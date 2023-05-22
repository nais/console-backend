package main

import (
	"context"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/hookd"
	"github.com/nais/console-backend/internal/k8s"
	"github.com/nais/console-backend/internal/search"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Audience        string
	BindHost        string
	ConsoleEndpoint string
	ConsoleToken    string
	HookdEndpoint   string
	HookdPSK        string
	Kubeconfig      string
	LogLevel        string
	Port            string
	RunAsUser       string
}

var cfg = &Config{}

func init() {
	flag.StringVar(&cfg.Audience, "audience", os.Getenv("IAP_AUDIENCE"), "IAP audience")
	flag.StringVar(&cfg.BindHost, "bind-host", os.Getenv("BIND_HOST"), "Bind host")
	flag.StringVar(&cfg.ConsoleEndpoint, "console-endpoint", envOrDefault("CONSOLE_ENDPOINT", "http://console.local.nais.io/query"), "Console endpoint")
	flag.StringVar(&cfg.ConsoleToken, "console-token", envOrDefault("CONSOLE_TOKEN", "secret-admin-api-key"), "Console Token")
	flag.StringVar(&cfg.HookdEndpoint, "hookd-endpoint", envOrDefault("HOOKD_ENDPOINT", "http://hookd.local.nais.io"), "Hookd endpoint")
	flag.StringVar(&cfg.HookdPSK, "hookd-psk", envOrDefault("HOOKD_PSK", "secret-frontend-psk"), "Hookd PSK")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "which log level to output")
	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "Port to listen on")
	flag.StringVar(&cfg.Kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "kubeconfig")
	flag.StringVar(&cfg.RunAsUser, "run-as-user", os.Getenv("RUN_AS_USER"), "Statically configured frontend user")
}

func main() {
	flag.Parse()
	log := newLogger()
	ctx := context.Background()

	k8s, err := k8s.New(cfg.Kubeconfig, log.WithField("client", "k8s"))
	if err != nil {
		log.Fatal(err)
	}

	k8s.Run(ctx)
	consoleClient := console.New(cfg.ConsoleToken, cfg.ConsoleEndpoint)
	searcher := search.New(consoleClient, k8s)

	graphConfig := graph.Config{
		Resolvers: &graph.Resolver{
			Hookd:    hookd.New(cfg.HookdPSK, cfg.HookdEndpoint),
			Console:  consoleClient,
			K8s:      k8s,
			Searcher: searcher,
		},
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graphConfig))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	if cfg.RunAsUser != "" && cfg.Audience == "" {
		log.Infof("Running as user %s", cfg.RunAsUser)
		http.Handle("/query", auth.StaticUser(cfg.RunAsUser, srv))
	} else {
		http.Handle("/query", auth.ValidateIAPJWT("audience")(srv))
	}

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.BindHost+":"+cfg.Port, nil))
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
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
