package k8s

import (
	"context"
	"strings"

	"google.golang.org/api/container/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

type cluster struct {
	Name     string
	Endpoint string
	CA       string
}

func createRestConfigs(projects []string) (map[string]rest.Config, error) {
	clusters, err := clusters(context.Background(), projects)
	if err != nil {
		return nil, err
	}

	ret := map[string]rest.Config{}
	for _, cluster := range clusters {
		// ca, err := base64.StdEncoding.DecodeString(cluster.CA)
		// if err != nil {
		// 	return nil, fmt.Errorf("base64 decoding CA for cluster %s: %w", cluster.Name, err)
		// }

		ret[cluster.Name] = rest.Config{
			Host: "https://apiserver." + cluster.Name + ".dev-nais.cloud.nais.io",
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

// clusters returns a list of cluster info entries for the given Google projects.
func clusters(ctx context.Context, projects []string) ([]cluster, error) {
	svc, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := []cluster{}

	for _, project := range projects {
		call := svc.Projects.Locations.Clusters.List("projects/" + project + "/locations/-")
		clusters, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, c := range clusters.Clusters {
			name := c.ResourceLabels["environment"]
			if name == "" {
				name = strings.TrimPrefix(c.Name, "nais-")
			}

			ret = append(ret, cluster{
				Name:     name,
				Endpoint: "https://" + c.Endpoint,
				CA:       c.MasterAuth.ClusterCaCertificate,
			})
		}
	}

	return ret, nil
}
