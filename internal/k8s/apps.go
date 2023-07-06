package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	naisv1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

type AppCondition string

const (
	AppConditionRolloutComplete       AppCondition = "RolloutComplete"
	AppConditionFailedSynchronization AppCondition = "FailedSynchronization"
	AppConditionSynchronized          AppCondition = "Synchronized"
	AppConditionUnknown               AppCondition = "Unknown"
)

func (c *Client) App(ctx context.Context, name, team, env string) (*model.App, error) {
	c.log.Debugf("getting app %q in namespace %q in env %q", name, team, env)
	if c.informers[env] == nil {
		return nil, fmt.Errorf("no appInformer for env %q", env)
	}
	obj, err := c.informers[env].AppInformer.Lister().ByNamespace(team).Get(name)
	if err != nil {
		return nil, c.error(ctx, err, "getting application")
	}
	return c.toApp(ctx, obj.(*unstructured.Unstructured), env)
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

func (c *Client) Apps(ctx context.Context, team string) ([]*model.App, error) {
	ret := []*model.App{}

	for env, infs := range c.informers {
		objs, err := infs.AppInformer.Lister().ByNamespace(team).List(labels.Everything())
		if err != nil {
			return nil, c.error(ctx, err, "listing applications")
		}
		for _, obj := range objs {
			app, err := c.toApp(ctx, obj.(*unstructured.Unstructured), env)
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

func (c *Client) Instances(ctx context.Context, team, env, name string) ([]*model.Instance, error) {
	req, err := labels.NewRequirement("app", selection.Equals, []string{name})
	if err != nil {
		return nil, c.error(ctx, err, "creating label selector")
	}

	selector := labels.NewSelector().Add(*req)
	pods, err := c.informers[env].PodInformer.Lister().Pods(team).List(selector)
	if err != nil {
		return nil, c.error(ctx, err, "listing pods")
	}

	ret := []*model.Instance{}
	for _, pod := range pods {
		ret = append(ret, Instance(pod))
	}
	return ret, nil
}

func Instance(pod *corev1.Pod) *model.Instance {
	appName := pod.Labels["app"]

	image := "unknown"
	for _, c := range pod.Spec.Containers {
		if c.Name == appName {
			image = c.Image
		}
	}

	appCS := appContainerStatus(pod, appName)

	ret := &model.Instance{
		ID:       model.Ident{ID: string(pod.GetUID()), Type: "pod"},
		Name:     pod.GetName(),
		Image:    image,
		Restarts: int(appCS.RestartCount),
		Message:  messageFromCS(appCS),
		State:    stateFromCS(appCS),
		Created:  pod.GetCreationTimestamp().Time,
	}

	return ret
}

func stateFromCS(cs *corev1.ContainerStatus) model.InstanceState {
	switch {
	case cs == nil:
		return model.InstanceStateUnknown
	case cs.State.Running != nil:
		return model.InstanceStateRunning
	case cs.State.Waiting != nil:
		return model.InstanceStateFailing
	default:
		return model.InstanceStateUnknown
	}
}

func messageFromCS(cs *corev1.ContainerStatus) string {
	if cs == nil {
		return ""
	}

	if cs.State.Waiting != nil {
		return cs.State.Waiting.Reason
	}

	return ""
}

func appContainerStatus(pod *corev1.Pod, appName string) *corev1.ContainerStatus {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == appName {
			return &cs
		}
	}
	return nil
}

func (c *Client) toApp(ctx context.Context, u *unstructured.Unstructured, env string) (*model.App, error) {
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

	appSynchState := app.GetStatus().SynchronizationState

	if appSynchState == "RolloutComplete" {
		timestamp := time.Unix(0, app.GetStatus().RolloutCompleteTime)
		ret.DeployInfo.Timestamp = &timestamp
	} else if appSynchState == "Synchronized" {
		timestamp := time.Unix(0, app.GetStatus().SynchronizationTime)
		ret.DeployInfo.Timestamp = &timestamp
	} else {
		ret.DeployInfo.Timestamp = nil
	}

	ret.DeployInfo.CommitSha = app.GetAnnotations()["deploy.nais.io/github-sha"]
	ret.DeployInfo.Deployer = app.GetAnnotations()["deploy.nais.io/github-actor"]
	ret.DeployInfo.URL = app.GetAnnotations()["deploy.nais.io/github-workflow-run-url"]
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

	instances, err := c.Instances(ctx, app.GetNamespace(), env, app.GetName())
	if err != nil {
		return nil, fmt.Errorf("getting instances: %w", err)
	}

	setStatus(ret, *app.Status.Conditions, instances)

	return ret, nil
}

func setStatus(app *model.App, conditions []metav1.Condition, instances []*model.Instance) {
	currentCondition := getCurrentCondition(conditions)
	failing := failing(instances)

	if currentCondition == AppConditionRolloutComplete && failing == 0 {
		app.State = model.AppStateNais
		return
	}

	switch currentCondition {
	case AppConditionFailedSynchronization:
		app.Messages = append(app.Messages, "Invalid nais.yaml")
	case AppConditionSynchronized:
		app.Messages = append(app.Messages, "New instances failing")
	}

	if failing == len(instances) {
		app.State = model.AppStateFailing
		app.Messages = append(app.Messages, "No running instances")
		return
	}

	app.State = model.AppStateNotnais
}

// for instances map reason to something human understandable

func failing(instances []*model.Instance) int {
	ret := 0
	for _, instance := range instances {
		if instance.State == model.InstanceStateFailing {
			ret++
		}
	}
	return ret
}

func getCurrentCondition(conditions []metav1.Condition) AppCondition {
	for _, condition := range conditions {
		if condition.Status == metav1.ConditionTrue {
			switch condition.Reason {
			case "RolloutComplete":
				return AppConditionRolloutComplete
			case "FailedSynchronization":
				return AppConditionFailedSynchronization
			case "Synchronized":
				return AppConditionSynchronized
			}
		}
	}
	return AppConditionUnknown
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
