package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	corev1inf "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"

type Client struct {
	informers map[string]*Informers
	log       *logrus.Entry
}
type Informers struct {
	AppInformer informers.GenericInformer
	PodInformer corev1inf.PodInformer
}

func New(kubeconfig string, log *logrus.Entry) (*Client, error) {
	infs := map[string]*Informers{}

	kubeConfig, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load config from file: %w", err)
	}

	for contextName, context := range kubeConfig.Contexts {
		infs[contextName] = &Informers{}
		contextSpec := &api.Context{Cluster: context.Cluster, AuthInfo: context.AuthInfo}
		restConfig, err := clientcmd.NewDefaultClientConfig(*kubeConfig, &clientcmd.ConfigOverrides{Context: *contextSpec}).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("create client config: %w", err)
		}

		clientSet, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("create clientset: %w", err)
		}

		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("create dynamic client: %w", err)
		}

		log.Debug("creating informers")
		dinf := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 4*time.Hour, "", nil)
		inf := informers.NewSharedInformerFactory(clientSet, 4*time.Hour)

		infs[contextName].PodInformer = inf.Core().V1().Pods()
		applicationGVK := schema.GroupVersion{Group: "nais.io", Version: "v1alpha1"}
		infs[contextName].AppInformer = dinf.ForResource(applicationGVK.WithResource("applications"))
	}

	return &Client{
		informers: infs,
		log:       log,
	}, nil
}

func (c *Client) Run(ctx context.Context) {
	for env, inf := range c.informers {
		c.log.Info("starting informers for ", env)
		go inf.PodInformer.Informer().Run(ctx.Done())
		go inf.AppInformer.Informer().Run(ctx.Done())
	}
}

func (c *Client) Apps(ctx context.Context, team string) ([]*model.App, error) {
	ret := []*model.App{}

	for env, infs := range c.informers {
		obj, err := infs.AppInformer.Lister().ByNamespace(team).List(labels.Everything())
		if err != nil {
			return nil, fmt.Errorf("listing applications: %w", err)
		}
		for _, app := range obj {
			application := app.(*unstructured.Unstructured)
			ret = append(ret, &model.App{
				Name: application.GetName(),
				Env: &model.Env{
					Name: env,
				},
			})
		}
	}
	return ret, nil
}
func (c *Client) Instances(ctx context.Context, team, env, name string) ([]*model.Instance, error) {
	ret := []*model.Instance{}
	req, err := labels.NewRequirement("app", selection.Equals, []string{name})
	if err != nil {
		return nil, fmt.Errorf("creating label selector: %w", err)
	}

	selector := labels.NewSelector().Add(*req)
	pods, err := c.informers[env].PodInformer.Lister().Pods(team).List(selector)
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	for _, pod := range pods {
		ret = append(ret, &model.Instance{
			ID:     string(pod.GetUID()),
			Name:   pod.GetName(),
			Status: string(pod.Status.Phase),
		})
	}
	return ret, nil
}