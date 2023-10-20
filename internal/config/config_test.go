package config_test

import (
	"context"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/nais/console-backend/internal/config"
	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	t.Run("missing required environment variables", func(t *testing.T) {
		cfg, err := config.New(ctx, envconfig.MapLookuper(map[string]string{}))
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "either RUN_AS_USER or IAP_AUDIENCE must be set")
	})

	t.Run("incorrect format for static k8s cluster", func(t *testing.T) {
		cfg, err := config.New(ctx, envconfig.MapLookuper(map[string]string{
			"KUBERNETES_CLUSTERS_STATIC": "foobar",
		}))
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, `invalid static cluster entry: "foobar"`)
	})

	t.Run("process config", func(t *testing.T) {
		cfg, err := config.New(ctx, envconfig.MapLookuper(map[string]string{
			"RUN_AS_USER": "some-user",
		}))
		assert.NoError(t, err)
		assert.Equal(t, false, cfg.Cost.Reimport)
		assert.Equal(t, bigquery.DetectProjectID, cfg.Cost.BigQueryProjectID)
		assert.Equal(t, "dev-nais", cfg.Cost.Tenant)

		assert.Equal(t, "http://hookd", cfg.Hookd.Endpoint)
		assert.Equal(t, "secret-frontend-psk", cfg.Hookd.PSK)

		assert.Empty(t, cfg.K8S.Clusters)
		assert.Equal(t, "", cfg.K8S.FieldSelector)
		assert.Empty(t, cfg.K8S.StaticClusters)
		assert.Equal(t, "dev-nais", cfg.K8S.Tenant)

		assert.Equal(t, "json", cfg.Logger.Format)
		assert.Equal(t, "info", cfg.Logger.Level)

		assert.Equal(t, "http://teams-backend/query", cfg.Teams.Endpoint)
		assert.Equal(t, "secret-admin-api-key", cfg.Teams.Token)

		assert.Equal(t, "", cfg.IapAudience)
		assert.Equal(t, "postgres://postgres:postgres@127.0.0.1:5432/console?sslmode=disable", cfg.DatabaseConnectionString)
		assert.Equal(t, ":8080", cfg.ListenAddress)
		assert.Equal(t, "some-user", cfg.RunAsUser)
	})
}
