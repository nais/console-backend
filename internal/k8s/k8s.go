package k8s

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/search"
	kafka_nais_io_v1 "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
	naisv1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	naisv1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	batchv1inf "k8s.io/client-go/informers/batch/v1"
	corev1inf "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
	"k8s.io/utils/strings/slices"
)

type Client struct {
	informers  map[string]*Informers
	clientSets map[string]*kubernetes.Clientset
	log        *logrus.Entry
	errors     metric.Int64Counter
}

type Informers struct {
	AppInformer     informers.GenericInformer
	PodInformer     corev1inf.PodInformer
	NaisjobInformer informers.GenericInformer
	JobInformer     batchv1inf.JobInformer
	TopicInformer   informers.GenericInformer
	EventInformer   corev1inf.EventInformer
}

func New(cfg config.K8S, errors metric.Int64Counter, log *logrus.Entry) (*Client, error) {
	restConfigs, err := CreateClusterConfigMap(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kubeconfig: %w", err)
	}

	infs := map[string]*Informers{}
	clientSets := map[string]*kubernetes.Clientset{}
	for cluster, restConfig := range restConfigs {
		infs[cluster] = &Informers{}

		clientSet, err := kubernetes.NewForConfig(&restConfig)
		if err != nil {
			return nil, fmt.Errorf("create clientset: %w", err)
		}

		dynamicClient, err := dynamic.NewForConfig(&restConfig)
		if err != nil {
			return nil, fmt.Errorf("create dynamic client: %w", err)
		}

		log.Debug("creating informers")
		dinf := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 4*time.Hour, "", func(options *metav1.ListOptions) {
			options.FieldSelector = cfg.FieldSelector
		})
		inf := informers.NewSharedInformerFactoryWithOptions(clientSet, 4*time.Hour, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = cfg.FieldSelector
		}))

		infs[cluster].PodInformer = inf.Core().V1().Pods()
		infs[cluster].AppInformer = dinf.ForResource(naisv1alpha1.GroupVersion.WithResource("applications"))
		infs[cluster].NaisjobInformer = dinf.ForResource(naisv1.GroupVersion.WithResource("naisjobs"))
		infs[cluster].JobInformer = inf.Batch().V1().Jobs()
		clientSets[cluster] = clientSet

		resources, err := discovery.NewDiscoveryClient(clientSet.RESTClient()).ServerResourcesForGroupVersion(kafka_nais_io_v1.GroupVersion.String())
		if err != nil && !strings.Contains(err.Error(), "the server could not find the requested resource") {
			return nil, fmt.Errorf("get server resources for group version: %w", err)
		}
		if err == nil {
			for _, r := range resources.APIResources {
				if r.Name == "topics" {
					infs[cluster].TopicInformer = dinf.ForResource(kafka_nais_io_v1.GroupVersion.WithResource("topics"))
				}
			}
		}
	}

	return &Client{
		informers:  infs,
		log:        log,
		errors:     errors,
		clientSets: clientSets,
	}, nil
}

func (c *Client) LogStream(ctx context.Context, cluster, namespace, selector, container string, instances []string) (<-chan *model.LogLine, error) {
	pods, err := c.clientSets[cluster].CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	ch := make(chan *model.LogLine, 10)
	for _, pod := range pods.Items {
		pod := pod
		if len(instances) > 0 && !slices.Contains(instances, pod.Name) {
			continue
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			logs, err := c.clientSets[cluster].CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
				Container:  container,
				Follow:     true,
				Timestamps: true,
				TailLines:  ptr.To[int64](int64(150 / len(pods.Items))),
			}).Stream(ctx)
			if err != nil {
				c.log.Error(err)
				return
			}
			defer logs.Close()
			sc := bufio.NewScanner(logs)

			for sc.Scan() {
				line := sc.Text()
				parts := strings.SplitN(line, " ", 2)
				if len(parts) != 2 {
					continue
				}
				time, err := time.Parse(time.RFC3339Nano, parts[0])
				if err != nil {
					continue
				}

				t := &model.LogLine{
					Time:     time,
					Message:  parts[1],
					Instance: pod.Name,
				}

				select {
				case <-ctx.Done():
					// Exit on cancellation
					c.log.Info("closing subscription")
					return

				case ch <- t:
					// Our message went through, do nothing
				}

			}

			c.log.Infof("Logs done, err: %v", sc.Err())
		}(wg)
	}
	go func() {
		wg.Wait()
		c.log.Info("closing subscription with explicit message")
		ch <- &model.LogLine{
			Time:     time.Now(),
			Message:  "Subscription closed.",
			Instance: "console-backend",
		}
		close(ch)
	}()
	return ch, nil
}

func (c *Client) Log(ctx context.Context, cluster, namespace, pod, container string, tailLines int64) ([]*model.LogLine, error) {
	logs, err := c.clientSets[cluster].CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{
		TailLines:  &tailLines,
		Container:  container,
		Follow:     false,
		Timestamps: true,
	}).Stream(ctx)
	if err != nil {
		return nil, err
	}
	defer logs.Close()

	sc := bufio.NewScanner(logs)

	ret := []*model.LogLine{}

	for sc.Scan() {
		line := sc.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		t, err := time.Parse(time.RFC3339Nano, parts[0])
		if err != nil {
			continue
		}
		ret = append(ret, &model.LogLine{
			Time:    t,
			Message: parts[1],
		})
	}

	return ret, nil
}

func (c *Client) Search(ctx context.Context, q string, filter *model.SearchFilter) []*search.Result {
	if !isFilterOrNoFilter(filter) {
		return nil
	}

	ret := []*search.Result{}

	for env, infs := range c.informers {
		if isFilterNaisjobOrNoFilter(filter) {
			jobs, err := infs.NaisjobInformer.Lister().List(labels.Everything())
			if err != nil {
				c.error(ctx, err, "listing jobs")
				return nil
			}

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

				ret = append(ret, &search.Result{
					Node: job,
					Rank: rank,
				})
			}
		}

		if isFilterAppOrNoFilter(filter) {
			apps, err := infs.AppInformer.Lister().List(labels.Everything())
			if err != nil {
				c.error(ctx, err, "listing applications")
				return nil
			}

			for _, obj := range apps {
				u := obj.(*unstructured.Unstructured)
				rank := search.Match(q, u.GetName())
				if rank == -1 {
					continue
				}
				app, err := c.toApp(ctx, u, env)
				if err != nil {
					c.error(ctx, err, "converting to app")
					return nil
				}

				ret = append(ret, &search.Result{
					Node: app,
					Rank: rank,
				})
			}
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
		if inf.TopicInformer != nil {
			go inf.TopicInformer.Informer().Run(ctx.Done())
		}
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
	c.errors.Add(ctx, 1, metric.WithAttributes(attribute.String("component", "k8s-client")))
	c.log.WithError(err).Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}

func isFilter(filter *model.SearchFilter) bool {
	if filter == nil {
		return false
	}

	if filter.Type == nil {
		return false
	}

	return true
}

func isFilterOrNoFilter(filter *model.SearchFilter) bool {
	if !isFilter(filter) {
		return true
	}

	return *filter.Type == model.SearchTypeApp || *filter.Type == model.SearchTypeNaisjob
}

func isFilterAppOrNoFilter(filter *model.SearchFilter) bool {
	if !isFilter(filter) {
		return true
	}

	return *filter.Type == model.SearchTypeApp
}

func isFilterNaisjobOrNoFilter(filter *model.SearchFilter) bool {
	if !isFilter(filter) {
		return true
	}

	return *filter.Type == model.SearchTypeNaisjob
}
