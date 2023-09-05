package k8s

import (
	"fmt"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

type ClusterConfigMap map[string]rest.Config

func CreateClusterConfigMap(clusters, static []string, tenant string) (ClusterConfigMap, error) {
	configs := ClusterConfigMap{}

	for _, cluster := range clusters {
		configs[cluster] = rest.Config{
			Host: fmt.Sprintf("https://apiserver.%s.%s.cloud.nais.io", cluster, tenant),
			AuthProvider: &api.AuthProviderConfig{
				Name: googleAuthPlugin,
			},
		}
	}

	staticConfigs, err := getStaticClusterConfigs(static)
	if err != nil {
		return nil, err
	}

	for cluster, cfg := range staticConfigs {
		configs[cluster] = cfg
	}

	return configs, nil
}

func getStaticClusterConfigs(static []string) (ClusterConfigMap, error) {
	configs := ClusterConfigMap{}

	for _, entry := range static {
		parts := strings.Split(entry, "|")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid static cluster entry: %q. Must be on format 'name|apiserver-host|token'", entry)
		}

		cluster := parts[0]
		host := parts[1]
		token := parts[2]

		configs[cluster] = rest.Config{
			Host:        host,
			BearerToken: token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	}
	return configs, nil
}
