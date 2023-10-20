package config_test

import (
	"os"
	"testing"

	"github.com/nais/console-backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("missing required environment variables", func(t *testing.T) {
		cfg, err := config.New()
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "either RUN_AS_USER or IAP_AUDIENCE must be set")
	})

	t.Run("incorrect format for static k8s cluster", func(t *testing.T) {
		assert.NoError(t, os.Setenv("KUBERNETES_CLUSTERS_STATIC", "foobar"))
		cfg, err := config.New()
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, `invalid static cluster entry: "foobar"`)
	})

	t.Run("process config", func(t *testing.T) {
		assert.NoError(t, os.Setenv("RUN_AS_USER", "some-user"))
		cfg, err := config.New()
		assert.Equal(t, "some-user", cfg.RunAsUser)
		assert.NoError(t, err)
	})
}
