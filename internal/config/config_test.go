package config_test

import (
	"os"
	"testing"

	"github.com/nais/console-backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.NoError(t, os.Setenv("TEAMS_ENDPOINT", "http://some/endpoint"))
	assert.NoError(t, os.Setenv("KUBERNETES_CLUSTERS", "cluster1,cluster2"))
	cfg := config.New()
	assert.Equal(t, "http://some/endpoint", cfg.TeamsEndpoint)
	assert.Equal(t, "secret-admin-api-key", cfg.TeamsToken)
	assert.Equal(t, "cluster1", cfg.KubernetesClusters[0])
	assert.Equal(t, "cluster2", cfg.KubernetesClusters[1])
	assert.Empty(t, cfg.KubernetesClustersStatic)
}
