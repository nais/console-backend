package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/search"
	naisv1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	naisv1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	batchv1inf "k8s.io/client-go/informers/batch/v1"
	corev1inf "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	informers map[string]*Informers
	log       *logrus.Entry
	errors    api.Int64Counter
}

type Informers struct {
	AppInformer     informers.GenericInformer
	PodInformer     corev1inf.PodInformer
	NaisjobInformer informers.GenericInformer
	JobInformer     batchv1inf.JobInformer
}

func New(clusters []string, tenant, fieldSelector string, errors api.Int64Counter, log *logrus.Entry) (*Client, error) {
	restConfigs, err := createRestConfigs(clusters, tenant)
	if err != nil {
		return nil, fmt.Errorf("create kubeconfig: %w", err)
	}

	infs := map[string]*Informers{}
	for cluster, cfg := range restConfigs {
		cfg := cfg
		infs[cluster] = &Informers{}

		clientSet, err := kubernetes.NewForConfig(&cfg)
		if err != nil {
			return nil, fmt.Errorf("create clientset: %w", err)
		}

		dynamicClient, err := dynamic.NewForConfig(&cfg)
		if err != nil {
			return nil, fmt.Errorf("create dynamic client: %w", err)
		}

		log.Debug("creating informers")
		dinf := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 4*time.Hour, "", func(options *metav1.ListOptions) {
			options.FieldSelector = fieldSelector
		})
		inf := informers.NewFilteredSharedInformerFactory(clientSet, 4*time.Hour, "", func(options *metav1.ListOptions) {
			options.FieldSelector = fieldSelector
		})

		infs[cluster].PodInformer = inf.Core().V1().Pods()
		infs[cluster].AppInformer = dinf.ForResource(naisv1alpha1.GroupVersion.WithResource("applications"))
		infs[cluster].NaisjobInformer = dinf.ForResource(naisv1.GroupVersion.WithResource("naisjobs"))
		infs[cluster].JobInformer = inf.Batch().V1().Jobs()
	}

	return &Client{
		informers: infs,
		log:       log,
		errors:    errors,
	}, nil
}

func (c *Client) Search(ctx context.Context, q string, filter *model.SearchFilter) []*search.SearchResult {
	// early exit if we're not searching for apps
	if filter != nil && filter.Type != nil && *filter.Type != model.SearchTypeApp {
		return nil
	}

	ret := []*search.SearchResult{}

	for env, infs := range c.informers {
		jobs, err := infs.NaisjobInformer.Lister().List(labels.Everything())
		if err != nil {
			c.error(ctx, err, "listing jobs")
			return nil
		}
		objs, err := infs.AppInformer.Lister().List(labels.Everything())
		if err != nil {
			c.error(ctx, err, "listing applications")
			return nil
		}

		fmt.Printf("jobs len: %d\n", len(jobs))
		for _, obj := range jobs {

			u := obj.(*unstructured.Unstructured)
			rank := search.Match(q, u.GetName())
			if rank == -1 {
				continue
			}
			job, err := toNaisJob(u, env)
			if err != nil {
				c.error(ctx, err, "converting to job")
				return nil
			}

			ret = append(ret, &search.SearchResult{
				Node: job,
				Rank: rank,
			})
		}

		for _, obj := range objs {
			u := obj.(*unstructured.Unstructured)
			rank := search.Match(q, u.GetName())
			if rank == -1 {
				continue
			}
			app, err := toApp(u, env)
			if err != nil {
				c.error(ctx, err, "converting to app")
				return nil
			}

			ret = append(ret, &search.SearchResult{
				Node: app,
				Rank: rank,
			})
		}

	}
	return ret
}

func (c *Client) Run(ctx context.Context) {
	for env, inf := range c.informers {
		c.log.Info("starting informers for ", env)
		go inf.PodInformer.Informer().Run(ctx.Done())
		go inf.AppInformer.Informer().Run(ctx.Done())
		go inf.NaisjobInformer.Informer().Run(ctx.Done())
		go inf.JobInformer.Informer().Run(ctx.Done())
	}
}

// convert takes a map[string]any / json, and converts it to the target struct
func convert(m any, target any) error {
	j, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshalling struct: %w", err)
	}
	err = json.Unmarshal(j, &target)
	if err != nil {
		return fmt.Errorf("unmarshalling json: %w", err)
	}
	return nil
}

func (c *Client) error(ctx context.Context, err error, msg string) error {
	c.errors.Add(context.Background(), 1, api.WithAttributes(attribute.String("component", "k8s-client")))
	c.log.WithError(err).Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}
