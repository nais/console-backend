package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/search"
	naisv1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	naisv1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
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

func (c *Client) App(ctx context.Context, name, team, env string) (*model.App, error) {
	c.log.Debugf("getting app %q in namespace %q in env %q", name, team, env)
	if c.informers[env] == nil {
		return nil, fmt.Errorf("no appInformer for env %q", env)
	}
	obj, err := c.informers[env].AppInformer.Lister().ByNamespace(team).Get(name)
	if err != nil {
		return nil, c.error(ctx, err, "getting application")
	}
	return toApp(obj.(*unstructured.Unstructured), env)
}

func (c *Client) NaisJob(ctx context.Context, name, team, env string) (*model.NaisJob, error) {
	c.log.Infof("getting job %q in namespace %q in env %q", name, team, env)
	if c.informers[env] == nil {
		return nil, fmt.Errorf("no jobInformer for env %q", env)
	}
	obj, err := c.informers[env].NaisjobInformer.Lister().ByNamespace(team).Get(name)
	if err != nil {
		return nil, c.error(ctx, err, "getting job")
	}
	return toNaisJob(obj.(*unstructured.Unstructured), env)
}

func (c *Client) Manifest(ctx context.Context, name, team, env string) (string, error) {
	obj, err := c.informers[env].AppInformer.Lister().ByNamespace(team).Get(name)
	if err != nil {
		return "", c.error(ctx, err, "getting application")
	}
	u := obj.(*unstructured.Unstructured)

	tmp := map[string]any{}

	spec, _, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil {
		return "", c.error(ctx, err, "getting spec")
	}

	tmp["spec"] = spec
	tmp["apiVersion"] = u.GetAPIVersion()
	tmp["kind"] = u.GetKind()
	metadata := map[string]any{"labels": u.GetLabels()}
	metadata["name"] = u.GetName()
	metadata["namespace"] = u.GetNamespace()
	tmp["metadata"] = metadata
	b, err := yaml.Marshal(tmp)
	if err != nil {
		return "", c.error(ctx, err, "marshalling manifest")
	}

	return string(b), nil
}

func (c *Client) JobManifest(ctx context.Context, name, team, env string) (string, error) {
	obj, err := c.informers[env].NaisjobInformer.Lister().ByNamespace(team).Get(name)
	if err != nil {
		return "", c.error(ctx, err, "getting job")
	}
	u := obj.(*unstructured.Unstructured)

	tmp := map[string]any{}

	spec, _, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil {
		return "", c.error(ctx, err, "getting spec")
	}

	tmp["spec"] = spec
	tmp["apiVersion"] = u.GetAPIVersion()
	tmp["kind"] = u.GetKind()
	metadata := map[string]any{"labels": u.GetLabels()}
	metadata["name"] = u.GetName()
	metadata["namespace"] = u.GetNamespace()
	tmp["metadata"] = metadata
	b, err := yaml.Marshal(tmp)
	if err != nil {
		return "", c.error(ctx, err, "marshalling manifest")
	}

	return string(b), nil
}

func (c *Client) Apps(ctx context.Context, team string) ([]*model.App, error) {
	ret := []*model.App{}

	for env, infs := range c.informers {
		objs, err := infs.AppInformer.Lister().ByNamespace(team).List(labels.Everything())
		if err != nil {
			return nil, c.error(ctx, err, "listing applications")
		}
		for _, obj := range objs {
			app, err := toApp(obj.(*unstructured.Unstructured), env)
			if err != nil {
				return nil, c.error(ctx, err, "converting to app")
			}
			ret = append(ret, app)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})

	return ret, nil
}

func (c *Client) Jobs(ctx context.Context, team string) ([]*model.NaisJob, error) {
	ret := []*model.NaisJob{}

	for env, infs := range c.informers {
		objs, err := infs.NaisjobInformer.Lister().ByNamespace(team).List(labels.Everything())
		if err != nil {
			return nil, c.error(ctx, err, "listing jobs")
		}
		for _, obj := range objs {
			job, err := toNaisJob(obj.(*unstructured.Unstructured), env)
			if err != nil {
				return nil, c.error(ctx, err, "converting to job")
			}
			ret = append(ret, job)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})

	return ret, nil
}

func (c *Client) JobInstances(ctx context.Context, team, env, name string) ([]*model.Run, error) {
	ret := []*model.Run{}

	nameReq, err := labels.NewRequirement("app", selection.Equals, []string{name})
	if err != nil {
		return nil, c.error(ctx, err, "creating label selector")
	}

	selector := labels.NewSelector().Add(*nameReq)

	jobs, err := c.informers[env].JobInformer.Lister().Jobs(team).List(selector)
	if err != nil {
		return nil, c.error(ctx, err, "listing job instances")
	}

	for _, job := range jobs {
		var completionTime *time.Time
		if job.Status.CompletionTime != nil {
			completionTime = &job.Status.CompletionTime.Time
		}

		ret = append(ret, &model.Run{
			ID:             model.Ident{ID: job.Name, Type: "job"},
			Name:           job.Name,
			StartTime:      job.Status.StartTime.Time,
			CompletionTime: completionTime,
			Failed:         failed(job),
			RunDuration:    duration(job).String(),
			Image:          job.Spec.Template.Spec.Containers[0].Image,
			Message:        message(job),
		})
	}

	// sort ret by StartTime, newest first
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].StartTime.After(ret[j].StartTime)
	})

	return ret, nil
}

func message(job *batchv1.Job) string {
	target := completionTarget(*job)
	if failed(job) {
		return "Run failed"
	}
	if job.Status.Active > 0 {
		return fmt.Sprintf("%d run(s) in progress. %d/%d runs finished successfully with %d failed.", job.Status.Active, job.Status.Succeeded, target, job.Status.Failed)
	}
	if job.Status.Succeeded == target {
		return fmt.Sprintf("%d/%d runs finished successfully", job.Status.Succeeded, target)
	}
	return ""
}

// completion target is the number of successful runs we want to see based on parallelism and completions
func completionTarget(job batchv1.Job) int32 {
	if job.Spec.Completions == nil && job.Spec.Parallelism == nil {
		return 1
	}
	if job.Spec.Completions != nil {
		return *job.Spec.Completions
	}
	return *job.Spec.Parallelism
}

func duration(job *batchv1.Job) time.Duration {
	if job.Status.StartTime == nil {
		return time.Duration(0)
	}
	if job.Status.CompletionTime != nil {
		return job.Status.CompletionTime.Sub(job.Status.StartTime.Time)
	}
	if !failed(job) {
		return time.Since(job.Status.StartTime.Time)
	}
	for _, cs := range job.Status.Conditions {
		if cs.Status == corev1.ConditionTrue {
			if cs.Type == batchv1.JobFailed {
				return cs.LastTransitionTime.Time.Sub(job.Status.StartTime.Time)
			}
		}
	}

	return time.Duration(0)
}

func failed(job *batchv1.Job) bool {
	for _, cs := range job.Status.Conditions {
		if cs.Status == corev1.ConditionTrue {
			if cs.Type == batchv1.JobFailed {
				return true
			}
		}
	}
	return false
}

func (c *Client) Instances(ctx context.Context, team, env, name string) ([]*model.Instance, error) {
	ret := []*model.Instance{}
	req, err := labels.NewRequirement("app", selection.Equals, []string{name})
	if err != nil {
		return nil, c.error(ctx, err, "creating label selector")
	}

	selector := labels.NewSelector().Add(*req)
	pods, err := c.informers[env].PodInformer.Lister().Pods(team).List(selector)
	if err != nil {
		return nil, c.error(ctx, err, "listing pods")
	}

	for _, pod := range pods {
		restarts := 0
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == name {
				restarts = int(cs.RestartCount)
			}
		}

		image := "unknown"
		for _, c := range pod.Spec.Containers {
			if c.Name == name {
				image = c.Image
			}
		}
		ret = append(ret, &model.Instance{
			ID:       model.Ident{ID: string(pod.GetUID()), Type: "pod"},
			Name:     pod.GetName(),
			Status:   string(pod.Status.Phase),
			Restarts: restarts,
			Image:    image,
			Created:  pod.GetCreationTimestamp().Time,
		})
	}
	return ret, nil
}

func toNaisJob(u *unstructured.Unstructured, env string) (*model.NaisJob, error) {
	naisjob := &naisv1.Naisjob{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, naisjob); err != nil {
		return nil, fmt.Errorf("converting to job: %w", err)
	}

	ret := &model.NaisJob{}
	ret.ID = model.Ident{ID: "job_" + env + "_" + naisjob.GetNamespace() + "_" + naisjob.GetName(), Type: "job"}
	ret.Name = naisjob.GetName()
	ret.Env = &model.Env{
		Name: env,
		ID:   model.Ident{ID: env, Type: "env"},
	}
	ret.DeployInfo = &model.DeployInfo{
		CommitSha: naisjob.GetAnnotations()["deploy.nais.io/github-sha"],
		Deployer:  naisjob.GetAnnotations()["deploy.nais.io/github-actor"],
		URL:       naisjob.GetAnnotations()["deploy.nais.io/github-workflow-run-url"],
	}
	ret.DeployInfo.GQLVars.Job = naisjob.GetName()
	ret.DeployInfo.GQLVars.Env = env
	ret.DeployInfo.GQLVars.Team = naisjob.GetNamespace()

	timestamp := time.Unix(0, naisjob.GetStatus().RolloutCompleteTime)
	ret.DeployInfo.Timestamp = &timestamp
	ret.GQLVars.Team = naisjob.GetNamespace()
	ret.Image = naisjob.Spec.Image

	ap := model.AccessPolicy{}
	if err := convert(naisjob.Spec.AccessPolicy, &ap); err != nil {
		return nil, fmt.Errorf("converting accessPolicy: %w", err)
	}
	ret.AccessPolicy = &ap

	r := model.Resources{}
	if err := convert(naisjob.Spec.Resources, &r); err != nil {
		return nil, fmt.Errorf("converting resources: %w", err)
	}

	if r.Requests == nil {
		r.Requests = &model.Requests{}
	}
	if r.Limits == nil {
		r.Limits = &model.Limits{}
	}
	ret.Resources = &r

	ret.Schedule = naisjob.Spec.Schedule

	if naisjob.Spec.Completions != nil {
		ret.Completions = int(*naisjob.Spec.Completions)
	}
	if naisjob.Spec.Parallelism != nil {
		ret.Parallelism = int(*naisjob.Spec.Parallelism)
	}
	ret.Retries = int(naisjob.Spec.BackoffLimit)

	return ret, nil
}

func toApp(u *unstructured.Unstructured, env string) (*model.App, error) {
	app := &naisv1alpha1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, app); err != nil {
		return nil, fmt.Errorf("converting to application: %w", err)
	}

	ret := &model.App{}
	ret.ID = model.Ident{ID: "app_" + env + "_" + app.GetNamespace() + "_" + app.GetName(), Type: "app"}
	ret.Name = app.GetName()

	ret.Env = &model.Env{
		Name: env,
		ID:   model.Ident{ID: env, Type: "env"},
	}
	timestamp := time.Unix(0, app.GetStatus().RolloutCompleteTime)
	ret.DeployInfo.CommitSha = app.GetAnnotations()["deploy.nais.io/github-sha"]
	ret.DeployInfo.Deployer = app.GetAnnotations()["deploy.nais.io/github-actor"]
	ret.DeployInfo.URL = app.GetAnnotations()["deploy.nais.io/github-workflow-run-url"]
	ret.DeployInfo.Timestamp = &timestamp
	ret.DeployInfo.GQLVars.App = app.GetName()
	ret.DeployInfo.GQLVars.Env = env
	ret.DeployInfo.GQLVars.Team = app.GetNamespace()
	ret.GQLVars.Team = app.GetNamespace()

	ret.Image = app.Spec.Image

	ingresses := []string{}
	if err := convert(app.Spec.Ingresses, &ingresses); err != nil {
		return nil, fmt.Errorf("converting ingresses: %w", err)
	}
	ret.Ingresses = ingresses

	if app.Spec.Replicas != nil {
		ret.AutoScaling = model.AutoScaling{
			Disabled:     app.Spec.Replicas.DisableAutoScaling,
			CPUThreshold: app.Spec.Replicas.CpuThresholdPercentage,
		}
		if app.Spec.Replicas.Min != nil {
			ret.AutoScaling.Min = *app.Spec.Replicas.Min
		}
		if app.Spec.Replicas.Max != nil {
			ret.AutoScaling.Max = *app.Spec.Replicas.Max
		}
	}

	ap := model.AccessPolicy{}
	if err := convert(app.Spec.AccessPolicy, &ap); err != nil {
		return nil, fmt.Errorf("converting accessPolicy: %w", err)
	}
	ret.AccessPolicy = ap

	r := model.Resources{}
	if err := convert(app.Spec.Resources, &r); err != nil {
		return nil, fmt.Errorf("converting resources: %w", err)
	}
	ret.Resources = r

	reps := model.Replicas{}
	if err := convert(app.Spec.Replicas, &reps); err != nil {
		return nil, fmt.Errorf("converting replicas: %w", err)
	}

	ret.Replicas = reps

	storage, err := appStorage(app)
	if err != nil {
		return nil, fmt.Errorf("getting storage: %w", err)
	}

	ret.Storage = storage

	authz, err := appAuthz(app)
	if err != nil {
		return nil, fmt.Errorf("getting authz: %w", err)
	}

	ret.Authz = authz

	for _, v := range app.Spec.Env {
		m := model.Variable{
			Name:  v.Name,
			Value: v.Value,
		}
		ret.Variables = append(ret.Variables, m)
	}

	if app.Status.RolloutCompleteTime > 0 {
		ret.Deployed = time.Unix(0, app.Status.RolloutCompleteTime)
	}

	return ret, nil
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

func appStorage(app *naisv1alpha1.Application) ([]model.Storage, error) {
	ret := []model.Storage{}

	if app.Spec.GCP != nil {
		for _, v := range app.Spec.GCP.Buckets {
			bucket := model.Bucket{}
			if err := convert(v, &bucket); err != nil {
				return nil, fmt.Errorf("converting buckets: %w", err)
			}
			ret = append(ret, bucket)
		}
		for _, v := range app.Spec.GCP.SqlInstances {
			sqlInstance := model.SQLInstance{}
			if err := convert(v, &sqlInstance); err != nil {
				return nil, fmt.Errorf("converting sqlInstance: %w", err)
			}
			if sqlInstance.Name == "" {
				sqlInstance.Name = app.Name
			}
			ret = append(ret, sqlInstance)
		}

		for _, v := range app.Spec.GCP.BigQueryDatasets {
			bqDataset := model.BigQueryDataset{}
			if err := convert(v, &bqDataset); err != nil {
				return nil, fmt.Errorf("converting bigQueryDataset: %w", err)
			}
			ret = append(ret, bqDataset)
		}
	}

	if app.Spec.OpenSearch != nil {
		os := model.OpenSearch{
			Name:   app.Spec.OpenSearch.Instance,
			Access: app.Spec.OpenSearch.Access,
		}
		ret = append(ret, os)
	}

	if app.Spec.Kafka != nil {
		kafka := model.Kafka{
			Name:    app.Spec.Kafka.Pool,
			Streams: app.Spec.Kafka.Streams,
		}
		ret = append(ret, kafka)
	}
	return ret, nil
}

func appAuthz(app *naisv1alpha1.Application) ([]model.Authz, error) {
	ret := []model.Authz{}
	if app.Spec.Azure != nil {
		isApp := app.Spec.Azure.Application != nil && app.Spec.Azure.Application.Enabled
		isSidecar := app.Spec.Azure.Sidecar != nil && app.Spec.Azure.Sidecar.Enabled
		if isApp || isSidecar {
			azureAd := model.AzureAd{}
			if err := convert(app.Spec.Azure, &azureAd); err != nil {
				return nil, fmt.Errorf("converting azureAd: %w", err)
			}
			ret = append(ret, azureAd)
		}
	}

	if app.Spec.IDPorten != nil && app.Spec.IDPorten.Enabled {
		idPorten := model.IDPorten{}
		if err := convert(app.Spec.IDPorten, &idPorten); err != nil {
			return nil, fmt.Errorf("converting idPorten: %w", err)
		}
		ret = append(ret, idPorten)
	}

	if app.Spec.Maskinporten != nil && app.Spec.Maskinporten.Enabled {
		maskinporten := model.Maskinporten{}
		if err := convert(app.Spec.Maskinporten, &maskinporten); err != nil {
			return nil, fmt.Errorf("converting maskinporten: %w", err)
		}
		ret = append(ret, maskinporten)
	}

	if app.Spec.TokenX != nil && app.Spec.TokenX.Enabled {
		tokenX := model.TokenX{}
		if err := convert(app.Spec.TokenX, &tokenX); err != nil {
			return nil, fmt.Errorf("converting tokenX: %w", err)
		}
		ret = append(ret, tokenX)
	}

	return ret, nil
}

func (c *Client) error(ctx context.Context, err error, msg string) error {
	c.errors.Add(context.Background(), 1, api.WithAttributes(attribute.String("component", "k8s-client")))
	c.log.WithError(err).Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}
