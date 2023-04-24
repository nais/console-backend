package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph"
	"github.com/nais/console-backend/internal/hookd"
)

type Config struct {
	BindHost        string
	Port            string
	Audience        string
	HookdEndpoint   string
	HookdPSK        string
	ConsoleEndpoint string
	ConsoleToken    string
}

var cfg = &Config{}

func init() {
	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "Port to listen on")
	flag.StringVar(&cfg.BindHost, "bind-host", os.Getenv("BIND_HOST"), "Bind host")
	flag.StringVar(&cfg.Audience, "audience", os.Getenv("IAP_AUDIENCE"), "IAP audience")
	flag.StringVar(&cfg.HookdEndpoint, "hookd-endpoint", envOrDefault("HOOKD_ENDPOINT", "http://hookd.local.nais.io"), "Hookd endpoint")
	flag.StringVar(&cfg.HookdPSK, "hookd-psk", envOrDefault("HOOKD_PSK", "secret-frontend-psk"), "Hookd PSK")
	flag.StringVar(&cfg.ConsoleEndpoint, "console-endpoint", envOrDefault("CONSOLE_ENDPOINT", "http://console.local.nais.io/query"), "Console endpoint")
	flag.StringVar(&cfg.ConsoleToken, "console-token", envOrDefault("CONSOLE_TOKEN", "secret"), "Console Token")
}

func main() {
	flag.Parse()
	graphConfig := graph.Config{
		Resolvers: &graph.Resolver{
			Hookd:   hookd.New(cfg.HookdPSK, cfg.HookdEndpoint),
			Console: console.New(cfg.ConsoleToken, cfg.ConsoleEndpoint),
		},
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graphConfig))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", auth.InsecureValidateMW(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.BindHost+":"+cfg.Port, nil))
}

func envOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
