package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/cost"
	"github.com/nais/console-backend/internal/database"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/search"
	"github.com/nais/console-backend/internal/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/prometheus"
	met "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const (
	costUpdateSchedule  = 1 * time.Hour
	exitCodeSuccess     = 0
	exitCodeLoggerError = 1
	exitCodeRunError    = 2
)

func main() {
	cfg := config.New()
	log, err := newLogger(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		fmt.Printf("create logger: %s", err)
		os.Exit(exitCodeLoggerError)
	}

	err = run(cfg, log)
	if err != nil {
		log.WithError(err).Errorf("error in run()")
		os.Exit(exitCodeRunError)
	}

	os.Exit(exitCodeSuccess)
}

func run(cfg *config.Config, log *logrus.Logger) error {
	ctx := context.Background()

	meter, err := getMetricMeter()
	if err != nil {
		return fmt.Errorf("create metric meter: %w", err)
	}

	errorsCounter, err := meter.Int64Counter("errors")
	if err != nil {
		return fmt.Errorf("create error counter: %w", err)
	}

	log.Info("connecting to database")
	querier, closer, err := database.NewQuerier(ctx, cfg.DBConnectionDSN, log.WithField("subsystem", "database"))
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}
	defer closer()

	err = runCostUpdater(ctx, querier, cfg, log)
	if err != nil {
		log.WithError(err).Error("unable to setup and run cost updater. You might need to run `gcloud auth --update-adc` if running locally")
	}

	k8sClient, err := k8s.New(cfg.KubernetesClusters, cfg.KubernetesClustersStatic, cfg.Tenant, cfg.FieldSelector, errorsCounter, log.WithField("client", "k8s"))
	if err != nil {
		return fmt.Errorf("create k8s client: %w", err)
	}

	k8sClient.Run(ctx)

	teamsBackendClient := teams.New(cfg.TeamsToken, cfg.TeamsEndpoint, errorsCounter, log.WithField("client", "teams"))
	searcher := search.New(teamsBackendClient, k8sClient)

	graphConfig := graph.Config{
		Resolvers: &graph.Resolver{
			Hookd:       hookd.New(cfg.HookdPSK, cfg.HookdEndpoint, errorsCounter, log.WithField("client", "hookd")),
			TeamsClient: teamsBackendClient,
			K8s:         k8sClient,
			Searcher:    searcher,
			Log:         log,
			Queries:     querier,
			Clusters:    cfg.KubernetesClusters,
		},
	}

	srv := newServer(graph.NewExecutableSchema(graphConfig))

	metricsMW, err := graph.NewMetrics(meter)
	if err != nil {
		return fmt.Errorf("create metrics middleware: %w", err)
	}
	srv.Use(metricsMW)

	corsMW := cors.New(
		cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowCredentials: true,
			Debug:            cfg.LogLevel == "debug",
		})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	if cfg.RunAsUser != "" && cfg.Audience == "" {
		log.Infof("Running as user %s", cfg.RunAsUser)
		http.Handle("/query", corsMW.Handler(auth.StaticUser(cfg.RunAsUser, srv)))
	} else {
		http.Handle("/query", corsMW.Handler(auth.ValidateIAPJWT(cfg.Audience, srv)))
	}

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("connect to http://%s:%s/ for GraphQL playground", cfg.BindHost, cfg.Port)
	err = http.ListenAndServe(cfg.BindHost+":"+cfg.Port, nil)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("HTTP server finished, terminating...")
	return nil
}

func newLogger(logFormat, logLevel string) (*logrus.Logger, error) {
	log := logrus.StandardLogger()

	switch logFormat {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{})
	default:
		return nil, fmt.Errorf("invalid log format: %q", logFormat)
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	log.SetLevel(level)
	return log, nil
}

func newServer(es graphql.ExecutableSchema) *handler.Server {
	srv := handler.New(es)
	srv.AddTransport(transport.SSE{}) // Support subscriptions
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return srv
}

// getBigQueryClient will return a new BigQuery client for the specified project
func getBigQueryClient(ctx context.Context, projectID string) (*bigquery.Client, error) {
	bigQueryClient, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	bigQueryClient.Location = "EU"
	return bigQueryClient, nil
}

// getBigQueryClient will return a new cost updater instance
func getUpdater(ctx context.Context, querier gensql.Querier, cfg *config.Config, log logrus.FieldLogger) (*cost.Updater, error) {
	bigQueryClient, err := getBigQueryClient(ctx, cfg.BigQueryProjectID)
	if err != nil {
		return nil, err
	}

	opts := make([]cost.Option, 0)
	if cfg.CostDataReimport {
		opts = append(opts, cost.WithReimport(true))
	}

	return cost.NewCostUpdater(
		bigQueryClient,
		querier,
		cfg.Tenant,
		log.WithField("subsystem", "cost_updater"),
		opts...,
	), nil
}

// runCostUpdater will create an instance of the cost updater, and update the costs on a schedule
func runCostUpdater(ctx context.Context, querier gensql.Querier, cfg *config.Config, log logrus.FieldLogger) error {
	updater, err := getUpdater(ctx, querier, cfg, log)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second) // initial run
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				func() {
					ticker.Reset(costUpdateSchedule) // regular schedule
					log.Debugf("start scheduled cost update run")
					start := time.Now()

					if shouldUpdate, err := updater.ShouldUpdateCosts(ctx); err != nil {
						log.WithError(err).Errorf("unable to check if costs should be updated")
						return
					} else if !shouldUpdate {
						log.Debugf("no need to update costs yet")
						return
					}

					ctx, cancel := context.WithTimeout(ctx, costUpdateSchedule-5*time.Minute)
					defer cancel()

					done := make(chan bool)
					defer close(done)

					ch := make(chan gensql.CostUpsertParams, cost.UpsertBatchSize*2)

					go func() {
						err := updater.UpdateCosts(ctx, ch)
						if err != nil {
							log.WithError(err).Errorf("failed to update costs")
						}
						done <- true
					}()

					err = updater.FetchBigQueryData(ctx, ch)
					if err != nil {
						log.WithError(err).Errorf("failed to fetch bigquery data")
					}
					close(ch)
					<-done

					log.WithFields(logrus.Fields{
						"duration": time.Since(start),
					}).Infof("cost update run finished")
				}()
			}
		}
	}()
	return nil
}

// getMetricMeter will return a new metric meter that uses a Prometheus exporter
func getMetricMeter() (met.Meter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %w", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	return provider.Meter("github.com/nais/console-backend"), nil
}
