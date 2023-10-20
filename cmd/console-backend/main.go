package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/cost"
	"github.com/nais/console-backend/internal/database"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/logger"
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
	log, err := logger.New(cfg.LogFormat, cfg.LogLevel)
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

func run(cfg *config.Config, log logrus.FieldLogger) error {
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

	go runCostUpdater(ctx, querier, cfg, log)

	k8sClient, teamsBackendClient, hookdClient, err := setupClients(cfg, errorsCounter, log)
	if err != nil {
		return fmt.Errorf("setup clients: %w", err)
	}
	k8sClient.Run(ctx)
	searcher := search.New(teamsBackendClient, k8sClient)

	graphHandler, err := graph.NewHandler(hookdClient, teamsBackendClient, k8sClient, searcher, querier, cfg.KubernetesClusters, log, meter)
	if err != nil {
		return fmt.Errorf("create graph handler: %w", err)
	}

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
		http.Handle("/query", corsMW.Handler(auth.StaticUser(cfg.RunAsUser, graphHandler)))
	} else {
		http.Handle("/query", corsMW.Handler(auth.ValidateIAPJWT(cfg.Audience, graphHandler)))
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

// runCostUpdater will create an instance of the cost updater, and update the costs on a schedule. This function will
// block until the context is cancelled, so it should be run in a goroutine.
func runCostUpdater(ctx context.Context, querier gensql.Querier, cfg *config.Config, log logrus.FieldLogger) {
	updater, err := getUpdater(ctx, querier, cfg, log)
	if err != nil {
		log.WithError(err).Error("unable to setup and run cost updater. You might need to run `gcloud auth --update-adc` if running locally")
	}

	ticker := time.NewTicker(1 * time.Second) // initial run
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				ticker.Reset(costUpdateSchedule) // regular schedule
				log.Infof("start scheduled cost update run")
				start := time.Now()

				if shouldUpdate, err := updater.ShouldUpdateCosts(ctx); err != nil {
					log.WithError(err).Errorf("unable to check if costs should be updated")
					return
				} else if !shouldUpdate {
					log.Infof("no need to update costs yet")
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

// setupClients will create and return the clients used by the application
func setupClients(cfg *config.Config, errorsCounter met.Int64Counter, log logrus.FieldLogger) (*k8s.Client, *teams.Client, *hookd.Client, error) {
	loggerFieldKey := "client"
	k8sClient, err := k8s.New(cfg.KubernetesClusters, cfg.KubernetesClustersStatic, cfg.Tenant, cfg.FieldSelector, errorsCounter, log.WithField(loggerFieldKey, "k8s"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create k8s client: %w", err)
	}

	teamsClient := teams.New(cfg.TeamsToken, cfg.TeamsEndpoint, errorsCounter, log.WithField(loggerFieldKey, "teams"))
	hookdClient := hookd.New(cfg.HookdPSK, cfg.HookdEndpoint, errorsCounter, log.WithField(loggerFieldKey, "hookd"))

	return k8sClient, teamsClient, hookdClient, nil
}
