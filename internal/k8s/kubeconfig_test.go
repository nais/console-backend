package k8s_test

import (
	"testing"

	"github.com/nais/console-backend/internal/config"

	"github.com/nais/console-backend/internal/k8s"
	"github.com/stretchr/testify/assert"
)

func TestCreateClusterConfigMap(t *testing.T) {
	const tenant = "tenant"

	t.Run("valid configuration with no static clusters", func(t *testing.T) {
		cfg := config.K8S{
			Clusters: []string{"cluster"},
		}
		configMap, err := k8s.CreateClusterConfigMap(tenant, cfg)
		assert.Equal(t, "https://apiserver.cluster.tenant.cloud.nais.io", configMap["cluster"].Host)
		assert.NoError(t, err)
	})

	t.Run("valid configuration with static clusters", func(t *testing.T) {
		cfg := config.K8S{
			Clusters:       []string{"cluster"},
			StaticClusters: []config.StaticCluster{{Name: "static-cluster", Host: "host", Token: "token"}},
		}
		configMap, err := k8s.CreateClusterConfigMap(tenant, cfg)
		assert.Equal(t, "https://apiserver.cluster.tenant.cloud.nais.io", configMap["cluster"].Host)
		assert.Equal(t, "host", configMap["static-cluster"].Host)
		assert.Equal(t, "token", configMap["static-cluster"].BearerToken)
		assert.NoError(t, err)
	})
}
