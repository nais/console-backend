package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
		objs, err := infs.AppInformer.Lister().ByNamespace(team).List(labels.Everything())
		if err != nil {
			return nil, fmt.Errorf("listing applications: %w", err)
		}
		for _, obj := range objs {
			app, err := toApp(obj, env)
			if err != nil {
				return nil, fmt.Errorf("converting to app: %w", err)
			}
			ret = append(ret, app)
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

func toApp(obj runtime.Object, env string) (*model.App, error) {
	u := obj.(*unstructured.Unstructured)
	ret := &model.App{}
	ret.ID = "app_" + env + "_" + u.GetNamespace() + "_" + u.GetName()
	ret.Name = u.GetName()
	ret.Env = &model.Env{
		Name: env,
		ID:   "env_" + env,
	}

	image, _, err := unstructured.NestedString(u.Object, "spec", "image")
	if err != nil {
		return nil, fmt.Errorf("getting image: %w", err)
	}
	ret.Image = image

	accessPolicy, _, err := unstructured.NestedMap(u.Object, "spec", "accessPolicy")
	if err != nil {
		return nil, fmt.Errorf("getting accessPolicy: %w", err)
	}
	apJsonString, err := json.Marshal(accessPolicy)
	if err != nil {
		return nil, fmt.Errorf("marshalling accessPolicy: %w", err)
	}
	a := model.AccessPolicy{}
	err = json.Unmarshal(apJsonString, &a)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling accessPolicy: %w", err)
	}
	ret.AccessPolicy = a

	resources, _, err := unstructured.NestedMap(u.Object, "spec", "resources")
	if err != nil {
		return nil, fmt.Errorf("getting resources: %w", err)
	}
	rJsonString, err := json.Marshal(resources)
	if err != nil {
		return nil, fmt.Errorf("marshalling resources: %w", err)
	}
	r := model.Resources{}
	err = json.Unmarshal(rJsonString, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling resources: %w", err)
	}
	ret.Resources = r

	ingresses, _, err := unstructured.NestedSlice(u.Object, "spec", "ingresses")
	if err != nil {
		return nil, fmt.Errorf("getting ingresses: %w", err)
	}

	fmt.Println(ingresses)
	iJsonString, err := json.Marshal(ingresses)
	if err != nil {
		return nil, fmt.Errorf("marshalling ingresses: %w", err)
	}
	i := []string{}
	err = json.Unmarshal(iJsonString, &i)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling ingresses: %w", err)
	}
	ret.Ingresses = i

	return ret, nil
}
