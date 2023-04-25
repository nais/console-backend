package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	Clientsets map[string]*kubernetes.Clientset
}

func New(kubeconfig string) (*Client, error) {
	kcs := map[string]*kubernetes.Clientset{}

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
		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("create clientset: %w", err)
		}
		kcs[contextName] = clientset
	}

	return &Client{Clientsets: kcs}, nil
}
