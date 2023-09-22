package config

import (
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

type Config struct {
	Audience                 string
	BindHost                 string
	DBConnectionDSN          string
	FieldSelector            string
	HookdEndpoint            string
	HookdPSK                 string
	LogFormat                string
	LogLevel                 string
	Port                     string
	RunAsUser                string
	TeamsEndpoint            string
	TeamsToken               string
	Tenant                   string
	KubernetesClusters       []string
	KubernetesClustersStatic []string
}

func New() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Audience, "audience", os.Getenv("IAP_AUDIENCE"), "IAP audience")
	flag.StringVar(&cfg.BindHost, "bind-host", os.Getenv("BIND_HOST"), "Bind host")
	flag.StringVar(&cfg.DBConnectionDSN, "db-connection-dsn", getEnv("CONSOLE_DBCONN_STRING", "postgres://postgres:postgres@127.0.0.1:5432/console?sslmode=disable"), "database connection DSN")
	flag.StringVar(&cfg.TeamsEndpoint, "teams-endpoint", envOrDefault("TEAMS_ENDPOINT", "http://teams-backend/query"), "Teams endpoint")
	flag.StringVar(&cfg.TeamsToken, "teams-token", envOrDefault("TEAMS_TOKEN", "secret-admin-api-key"), "Teams token")
	flag.StringVar(&cfg.HookdEndpoint, "hookd-endpoint", envOrDefault("HOOKD_ENDPOINT", "http://hookd"), "Hookd endpoint")
	flag.StringVar(&cfg.HookdPSK, "hookd-psk", envOrDefault("HOOKD_PSK", "secret-frontend-psk"), "Hookd PSK")
	flag.StringVar(&cfg.LogFormat, "log-format", "json", "which log format to use")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "which log level to output")
	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "Port to listen on")
	flag.StringVar(&cfg.Tenant, "tenant", envOrDefault("TENANT", "dev-nais"), "Which tenant we are running in")
	flag.StringVar(&cfg.RunAsUser, "run-as-user", os.Getenv("RUN_AS_USER"), "Statically configured frontend user")
	flag.StringVar(&cfg.FieldSelector, "field-selector", os.Getenv("FIELD_SELECTOR"), "Field selector for k8s resources")
	flag.StringSliceVar(&cfg.KubernetesClusters, "kubernetes-clusters", splitEnv("KUBERNETES_CLUSTERS", ","), "Kubernetes clusters to watch (comma separated)")
	flag.StringSliceVar(&cfg.KubernetesClustersStatic, "kubernetes-clusters-static", splitEnv("KUBERNETES_CLUSTERS_STATIC", ","), "Kubernetes clusters to watch with static credentials (comma separated entries on the format 'name|apiserver-host|token')")
	flag.Parse()

	return cfg
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

func getEnv(key, fallback string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return fallback
}
