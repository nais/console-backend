package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi/v5"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"

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
	"github.com/nais/console-backend/internal/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/prometheus"
	met "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const (
	exitCodeSuccess = iota
	exitCodeLoggerError
	exitCodeRunError
	exitCodeConfigError
)

const (
	costUpdateSchedule = 1 * time.Hour
)

func main() {
	ctx := context.Background()
	cfg, err := config.New(ctx, envconfig.OsLookuper())
	if err != nil {
		fmt.Printf("error when processing configuration: %s", err)
		os.Exit(exitCodeConfigError)
	}

	log, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("error when creating logger: %s", err)
		os.Exit(exitCodeLoggerError)
	}

	err = run(ctx, cfg, log)
	if err != nil {
		log.WithError(err).Errorf("error in run()")
		os.Exit(exitCodeRunError)
	}

	os.Exit(exitCodeSuccess)
}

func run(ctx context.Context, cfg *config.Config, log logrus.FieldLogger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	signals := make(chan os.Signal, 1)
	defer close(signals)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	meter, err := getMetricMeter()
	if err != nil {
		return fmt.Errorf("create metric meter: %w", err)
	}

	errorsCounter, err := meter.Int64Counter("errors")
	if err != nil {
		return fmt.Errorf("create error counter: %w", err)
	}

	log.Info("connecting to database")
	querier, closer, err := database.NewQuerier(ctx, cfg.DatabaseConnectionString, log.WithField("subsystem", "database"))
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}
	defer closer()

	k8sClient, teamsBackendClient, hookdClient, err := setupClients(cfg, errorsCounter, log)
	if err != nil {
		return fmt.Errorf("setup clients: %w", err)
	}

	resolver := graph.NewResolver(hookdClient, teamsBackendClient, k8sClient, querier, cfg.K8S.Clusters, log)
	graphHandler, err := graph.NewHandler(graph.Config{Resolvers: resolver}, meter)
	if err != nil {
		return fmt.Errorf("create graph handler: %w", err)
	}

	// k8s informers
	go func() {
		stopCh := ctx.Done()
		for cluster, informer := range k8sClient.Informers() {
			log.WithField("cluster", cluster).Infof("starting informers")
			go informer.PodInformer.Informer().Run(stopCh)
			go informer.AppInformer.Informer().Run(stopCh)
			go informer.NaisjobInformer.Informer().Run(stopCh)
			go informer.JobInformer.Informer().Run(stopCh)
			if informer.TopicInformer != nil {
				_, err := informer.TopicInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						log.Infof("added topic: %s", obj.(*unstructured.Unstructured).GetName())
					},
					UpdateFunc: func(oldObj, newObj interface{}) {
						log.Infof("updated topic: %s", newObj.(*unstructured.Unstructured).GetName())
					},
					DeleteFunc: func(obj interface{}) {
						log.Infof("deleted topic: %s", obj.(*unstructured.Unstructured).GetName())
					},
				})
				if err != nil {
					log.WithError(err).Errorf("error adding event handler")
				}

				/*listWatcher := cache.NewListWatchFromClient(
					k8sClient.ClientSets[cluster].RESTClient(),
					informer.TopicInformer.Informer().AddEventHandler(),
					"kafka.nais.io", // Namespace
					fields.Everything(),
				)
				resyncPeriod := 30 * time.Second
				_, controller := cache.NewInformer(
					listWatcher,
					&kafka_nais_io_v1.Topic{}, // Change to your custom resource's API type
					resyncPeriod,
					cache.ResourceEventHandlerFuncs{
						AddFunc: func(obj interface{}) {
							log.Infof("added topic: %s", obj)
						},
						UpdateFunc: func(oldObj, newObj interface{}) {
							log.Infof("updated topic: %s", newObj)
						},
						DeleteFunc: func(obj interface{}) {
							log.Infof("deleted topic: %s", obj)
						},
					},
				)
				go controller.Run(stopCh)*/

				go informer.TopicInformer.Informer().Run(stopCh)
			}
		}
	}()

	// cost updater
	go func() {
		defer cancel()
		err = runCostUpdater(ctx, querier, cfg.Cost, log)
		if err != nil {
			log.WithError(err).Errorf("error in cost updater")
		}
	}()

	// HTTP server
	go func() {
		defer cancel()
		srv := getHttpServer(cfg, graphHandler)
		err = srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Infof("unexpected error from HTTP server")
		}
		log.Infof("HTTP server finished, terminating...")
	}()

	// signal handling
	go func() {
		defer cancel()
		sig := <-signals
		log.Infof("received signal %s, terminating...", sig)
	}()

	<-ctx.Done()
	return ctx.Err()
}

// getHttpServer will return a new HTTP server with the specified configuration
func getHttpServer(cfg *config.Config, graphHandler *handler.Server) *http.Server {
	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	router.Get("/healthz", func(_ http.ResponseWriter, _ *http.Request) {})
	router.Get("/", playground.Handler("GraphQL playground", "/query"))

	middlewares := []func(http.Handler) http.Handler{
		cors.New(
			cors.Options{
				AllowedOrigins:   []string{"https://*", "http://*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowCredentials: true,
			},
		).Handler,
	}

	authMiddlware := auth.ValidateIAPJWT(cfg.IapAudience)
	if cfg.RunAsUser != "" {
		authMiddlware = auth.StaticUser(cfg.RunAsUser)
	}
	middlewares = append(middlewares, authMiddlware)

	router.Route("/query", func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/", graphHandler.ServeHTTP)
	})

	return &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: router,
	}
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
func getUpdater(ctx context.Context, querier gensql.Querier, cfg config.Cost, log logrus.FieldLogger) (*cost.Updater, error) {
	bigQueryClient, err := getBigQueryClient(ctx, cfg.BigQueryProjectID)
	if err != nil {
		return nil, err
	}

	opts := make([]cost.Option, 0)
	if cfg.Reimport {
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
func runCostUpdater(ctx context.Context, querier gensql.Querier, cfg config.Cost, log logrus.FieldLogger) error {
	updater, err := getUpdater(ctx, querier, cfg, log)
	if err != nil {
		return fmt.Errorf("unable to set up and run cost updater: %w", err)
	}

	ticker := time.NewTicker(1 * time.Second) // initial run
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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

				done := make(chan struct{})
				defer close(done)

				ch := make(chan gensql.CostUpsertParams, cost.UpsertBatchSize*2)

				go func() {
					err := updater.UpdateCosts(ctx, ch)
					if err != nil {
						log.WithError(err).Errorf("failed to update costs")
					}
					done <- struct{}{}
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
func setupClients(cfg *config.Config, errorsCounter met.Int64Counter, log logrus.FieldLogger) (*k8s.Client, *teams.Client, hookd.Client, error) {
	loggerFieldKey := "client"
	k8sClient, err := k8s.New(cfg.K8S, errorsCounter, log.WithField(loggerFieldKey, "k8s"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create k8s client: %w", err)
	}

	teamsClient := teams.New(cfg.Teams, errorsCounter, log.WithField(loggerFieldKey, "teams"))
	hookdClient := hookd.New(cfg.Hookd, errorsCounter, log.WithField(loggerFieldKey, "hookd"))

	return k8sClient, teamsClient, hookdClient, nil
}
