package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Cost is the configuration for the cost service
type Cost struct {
	Reimport          bool   `envconfig:"COST_DATA_REIMPORT" default:"false"`
	BigQueryProjectID string `envconfig:"BIGQUERY_PROJECTID" default:"*detect-project-id*"`
	Tenant            string `envconfig:"TENANT" default:"dev-nais"`
}

// Hookd is the configuration for the hookd service
type Hookd struct {
	Endpoint string `envconfig:"HOOKD_ENDPOINT" default:"http://hookd"`
	PSK      string `envconfig:"HOOKD_PSK" default:"secret-frontend-psk"`
}

// K8S is the configuration related to Kubernetes
type K8S struct {
	Clusters       []string        `envconfig:"KUBERNETES_CLUSTERS"`
	FieldSelector  string          `envconfig:"KUBERNETES_FIELD_SELECTOR"`
	StaticClusters []StaticCluster `envconfig:"KUBERNETES_CLUSTERS_STATIC"`
	Tenant         string          `envconfig:"TENANT" default:"dev-nais"`
}

// Logger is the configuration for the logger
type Logger struct {
	Format string `envconfig:"LOG_FORMAT" default:"json"`
	Level  string `envconfig:"LOG_LEVEL" default:"info"`
}

// Teams is the configuration for the teams backend service
type Teams struct {
	Endpoint string `envconfig:"TEAMS_ENDPOINT" default:"http://teams-backend/query"`
	Token    string `envconfig:"TEAMS_TOKEN" default:"secret-admin-api-key"`
}

// Config is the configuration for the console-backend application
type Config struct {
	Cost   Cost
	Hookd  Hookd
	K8S    K8S
	Logger Logger
	Teams  Teams

	// IapAudience is the audience for the IAP JWT token. Will not be used when RUN_AS_USER is set
	IapAudience string `envconfig:"IAP_AUDIENCE"`

	// DatabaseConnectionString is the database DSN
	DatabaseConnectionString string `envconfig:"CONSOLE_DATABASE_URL" default:"postgres://postgres:postgres@127.0.0.1:5432/console?sslmode=disable"`

	// ListenAddress is host:port combination used by the http server
	ListenAddress string `envconfig:"LISTEN_ADDRESS" default:":8080"`

	// RunAsUser is the static user to run as. Used for development purposes. Will override IAP_AUDIENCE when set
	RunAsUser string `envconfig:"RUN_AS_USER"`
}

// New creates a new configuration instance from environment variables
func New() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}
	if cfg.RunAsUser == "" && cfg.IapAudience == "" {
		return nil, fmt.Errorf("either RUN_AS_USER or IAP_AUDIENCE must be set")
	}
	return cfg, nil
}
