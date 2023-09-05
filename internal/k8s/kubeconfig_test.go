package k8s_test

import (
	"testing"

	"github.com/nais/console-backend/internal/k8s"
	"github.com/stretchr/testify/assert"
)

func TestCreateClusterConfigMap(t *testing.T) {
	const tenant = "tenant"

	t.Run("invalid static cluster entry", func(t *testing.T) {
		cfg, err := k8s.CreateClusterConfigMap([]string{"cluster"}, []string{"invalid"}, tenant)
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, `invalid static cluster entry: "invalid". Must be on format 'name|apiserver-host|token'`)
	})

	t.Run("valid configuration with no static clusters", func(t *testing.T) {
		cfg, err := k8s.CreateClusterConfigMap([]string{"cluster"}, nil, tenant)
		assert.Equal(t, "https://apiserver.cluster.tenant.cloud.nais.io", cfg["cluster"].Host)
		assert.NoError(t, err)
	})

	t.Run("valid configuration with static clusters", func(t *testing.T) {
		cfg, err := k8s.CreateClusterConfigMap([]string{"cluster"}, []string{"static-cluster|host|token"}, tenant)
		assert.Equal(t, "https://apiserver.cluster.tenant.cloud.nais.io", cfg["cluster"].Host)
		assert.Equal(t, "host", cfg["static-cluster"].Host)
		assert.Equal(t, "token", cfg["static-cluster"].BearerToken)
		assert.NoError(t, err)
	})
}
