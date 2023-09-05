package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/search"
	"github.com/nais/console-backend/internal/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

const (
	exitCodeSuccess     = 0
	exitCodeLoggerError = 1
	exitCodeRunError    = 2
)

func main() {
	cfg := config.New()
	log, err := newLogger(cfg.LogLevel)
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

	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("create prometheus exporter: %w", err)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("github.com/nais/console-backend")

	errorsCounter, err := meter.Int64Counter("errors")
	if err != nil {
		return fmt.Errorf("create error counter: %w", err)
	}

	k8s, err := k8s.New(cfg.KubernetesClusters, cfg.KubernetesClustersStatic, cfg.Tenant, cfg.FieldSelector, errorsCounter, log.WithField("client", "k8s"))
	if err != nil {
		return fmt.Errorf("create k8s client: %w", err)
	}

	k8s.Run(ctx)

	teams := teams.New(cfg.TeamsToken, cfg.TeamsEndpoint, errorsCounter, log.WithField("client", "teams"))
	searcher := search.New(teams, k8s)

	graphConfig := graph.Config{
		Resolvers: &graph.Resolver{
			Hookd:       hookd.New(cfg.HookdPSK, cfg.HookdEndpoint, errorsCounter, log.WithField("client", "hookd")),
			TeamsClient: teams,
			K8s:         k8s,
			Searcher:    searcher,
			Log:         log,
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
		http.Handle("/query", corsMW.Handler(auth.ValidateIAPJWT(cfg.Audience)(srv)))
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

func newLogger(logLevel string) (*logrus.Logger, error) {
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})

	l, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	log.SetLevel(l)
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
