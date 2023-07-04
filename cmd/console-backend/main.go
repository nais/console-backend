package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/search"
	"github.com/nais/console-backend/internal/teams"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

type Config struct {
	Audience                 string
	BindHost                 string
	FieldSelector            string
	HookdEndpoint            string
	HookdPSK                 string
	LogLevel                 string
	Port                     string
	RunAsUser                string
	TeamsEndpoint            string
	TeamsToken               string
	Tenant                   string
	KubernetesClusters       []string
	KubernetesClustersStatic []string
}

var cfg = &Config{}

func init() {
	flag.StringVar(&cfg.Audience, "audience", os.Getenv("IAP_AUDIENCE"), "IAP audience")
	flag.StringVar(&cfg.BindHost, "bind-host", os.Getenv("BIND_HOST"), "Bind host")
	flag.StringVar(&cfg.TeamsEndpoint, "teams-endpoint", envOrDefault("TEAMS_ENDPOINT", "http://teams-backend/query"), "Teams endpoint")
	flag.StringVar(&cfg.TeamsToken, "teams-token", envOrDefault("TEAMS_TOKEN", "secret-admin-api-key"), "Teams token")
	flag.StringVar(&cfg.HookdEndpoint, "hookd-endpoint", envOrDefault("HOOKD_ENDPOINT", "http://hookd"), "Hookd endpoint")
	flag.StringVar(&cfg.HookdPSK, "hookd-psk", envOrDefault("HOOKD_PSK", "secret-frontend-psk"), "Hookd PSK")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "which log level to output")
	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "Port to listen on")
	flag.StringVar(&cfg.Tenant, "tenant", envOrDefault("TENANT", "dev-nais"), "Which tenant we are running in")
	flag.StringVar(&cfg.RunAsUser, "run-as-user", os.Getenv("RUN_AS_USER"), "Statically configured frontend user")
	flag.StringVar(&cfg.FieldSelector, "field-selector", os.Getenv("FIELD_SELECTOR"), "Field selector for k8s resources")
	flag.StringSliceVar(&cfg.KubernetesClusters, "kubernetes-clusters", splitEnv("KUBERNETES_CLUSTERS", ","), "Kubernetes clusters to watch (comma separated)")
	flag.StringSliceVar(&cfg.KubernetesClustersStatic, "kubernetes-clusters-static", splitEnv("KUBERNETES_CLUSTERS_STATIC", ","), "Kubernetes clusters to watch with static credentials (comma separated entries on the format 'name|apiserver-host|token')")
}

func main() {
	flag.Parse()
	log := newLogger()
	ctx := context.Background()

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("github.com/nais/console-backend")

	errors, err := meter.Int64Counter("errors")
	if err != nil {
		log.Fatalf("creating error counter: %v", err)
	}

	k8s, err := k8s.New(cfg.KubernetesClusters, cfg.KubernetesClustersStatic, cfg.Tenant, cfg.FieldSelector, errors, log.WithField("client", "k8s"))
	if err != nil {
		log.Fatal(err)
	}

	k8s.Run(ctx)

	teams := teams.New(cfg.TeamsToken, cfg.TeamsEndpoint, errors, log.WithField("client", "teams"))
	searcher := search.New(teams, k8s)

	graphConfig := graph.Config{
		Resolvers: &graph.Resolver{
			Hookd:       hookd.New(cfg.HookdPSK, cfg.HookdEndpoint, errors, log.WithField("client", "hookd")),
			TeamsClient: teams,
			K8s:         k8s,
			Searcher:    searcher,
			Log:         log,
		},
	}

	corsMW := cors.New(
		cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowCredentials: true,
			Debug:            cfg.LogLevel == "debug",
		})

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graphConfig))

	metricsMW, err := graph.NewMetrics(meter)
	if err != nil {
		log.WithError(err).Fatal("setting up metrics middleware")
	}
	srv.Use(metricsMW)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	if cfg.RunAsUser != "" && cfg.Audience == "" {
		log.Infof("Running as user %s", cfg.RunAsUser)
		http.Handle("/query", corsMW.Handler(auth.StaticUser(cfg.RunAsUser, srv)))
	} else {
		http.Handle("/query", corsMW.Handler(auth.ValidateIAPJWT(cfg.Audience)(srv)))
	}

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("connect to http://%s:%s/ for GraphQL playground", cfg.BindHost, cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.BindHost+":"+cfg.Port, nil))
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func splitEnv(key, sep string) []string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Split(value, sep)
	}
	return nil
}

func newLogger() *logrus.Logger {
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(l)
	return log
}
