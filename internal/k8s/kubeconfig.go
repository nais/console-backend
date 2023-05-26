package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/container/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

type cluster struct {
	Name     string
	Endpoint string
	CA       string
}

func createKubeConfig(projects []string) (*api.Config, error) {
	clusters, err := clusters(context.Background(), projects)
	if err != nil {
		return nil, err
	}

	ret := &api.Config{
		Clusters:  map[string]*api.Cluster{},
		Contexts:  map[string]*api.Context{},
		AuthInfos: map[string]*api.AuthInfo{},
	}

	for _, cluster := range clusters {
		ca, err := base64.StdEncoding.DecodeString(cluster.CA)
		if err != nil {
			return nil, fmt.Errorf("base64 decoding CA for cluster %s: %w", cluster.Name, err)
		}

		ret.Clusters[cluster.Name] = &api.Cluster{
			Server:                   cluster.Endpoint,
			CertificateAuthorityData: ca,
		}
		ret.Contexts[cluster.Name] = &api.Context{
			Cluster:  cluster.Name,
			AuthInfo: "user",
		}
	}
	ret.AuthInfos["user"] = &api.AuthInfo{
		Exec: &api.ExecConfig{
			APIVersion:         "client.authentication.k8s.io/v1beta1",
			Command:            "gke-gcloud-auth-plugin",
			ProvideClusterInfo: true,
			InteractiveMode:    api.IfAvailableExecInteractiveMode,
		},
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
