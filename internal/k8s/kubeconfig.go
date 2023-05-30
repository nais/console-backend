package k8s

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

func createRestConfigs(clusters []string, tenant string) (map[string]rest.Config, error) {
	ret := map[string]rest.Config{}
	for _, cluster := range clusters {
		ret[cluster] = rest.Config{
			Host: "https://apiserver." + cluster + "." + tenant + ".cloud.nais.io",
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
			AuthProvider: &api.AuthProviderConfig{
				Name: googleAuthPlugin,
			},
		}
	}

	return ret, nil
}
