package k8s

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"

type Client struct {
	DynamicClients map[string]*dynamic.DynamicClient
}

func New(kubeconfig string) (*Client, error) {
	kcs := map[string]*dynamic.DynamicClient{}

	kubeConfig, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load config from file: %w", err)
	}

	for contextName, context := range kubeConfig.Contexts {
		contextSpec := &api.Context{Cluster: context.Cluster, AuthInfo: context.AuthInfo}
		restConfig, err := clientcmd.NewDefaultClientConfig(*kubeConfig, &clientcmd.ConfigOverrides{Context: *contextSpec}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("create client config: %w", err)
		}
		clientset, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("create clientset: %w", err)
		}
		kcs[contextName] = clientset
	}

	return &Client{DynamicClients: kcs}, nil
}

func (c *Client) Apps(ctx context.Context, team string) ([]*model.App, error) {
	ret := []*model.App{}
	applicationGVK := schema.GroupVersion{Group: "nais.io", Version: "v1alpha1"}

	for env, client := range c.DynamicClients {
		apps, err := client.Resource(applicationGVK.WithResource("applications")).Namespace(team).List(ctx, v1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("listing applications: %w", err)
		}
		for _, app := range apps.Items {
			ret = append(ret, &model.App{
				Name: app.GetName(),
				Env: &model.Env{
					Name: env,
				},
			})
		}
	}
	return ret, nil
}
