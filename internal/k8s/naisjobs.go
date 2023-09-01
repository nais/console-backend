package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	naisv1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
)

func (c *Client) NaisJob(ctx context.Context, name, team, env string) (*model.NaisJob, error) {
	c.log.Debugf("getting job %q in namespace %q in env %q", name, team, env)
	if c.informers[env] == nil {
		return nil, fmt.Errorf("no jobInformer for env %q", env)
	}
	obj, err := c.informers[env].NaisjobInformer.Lister().ByNamespace(team).Get(name)
	if err != nil {
		return nil, c.error(ctx, err, "getting job")
	}

	job, err := toNaisJob(obj.(*unstructured.Unstructured), env)
	if err != nil {
		return nil, c.error(ctx, err, "converting to job")
	}

	topics, err := c.getTopics(ctx, name, team, env)
	if err != nil {
		return nil, c.error(ctx, err, "getting topics")
	}

	storage, err := naisjobStorage(obj.(*unstructured.Unstructured), topics)
	if err != nil {
		return nil, c.error(ctx, err, "getting storage")
	}

	job.Storage = storage

	return job, nil
}

func (c *Client) NaisJobs(ctx context.Context, team string) ([]*model.NaisJob, error) {
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

func (c *Client) NaisJobManifest(ctx context.Context, name, team, env string) (string, error) {
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

func (c *Client) Runs(ctx context.Context, team, env, name string) ([]*model.Run, error) {
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
		var startTime, completionTime *time.Time
		if job.Status.CompletionTime != nil {
			completionTime = &job.Status.CompletionTime.Time
		}
		if job.Status.StartTime != nil {
			startTime = &job.Status.StartTime.Time
		}

		podReq, err := labels.NewRequirement("job-name", selection.Equals, []string{job.Name})
		if err != nil {
			return nil, c.error(ctx, err, "creating label selector")
		}
		podSelector := labels.NewSelector().Add(*podReq)
		pods, err := c.informers[env].PodInformer.Lister().Pods(team).List(podSelector)
		if err != nil {
			return nil, c.error(ctx, err, "listing job instance pods")
		}

		var podNames []string
		for _, pod := range pods {
			podNames = append(podNames, pod.Name)
		}

		ret = append(ret, &model.Run{
			ID:             model.Ident{ID: job.Name, Type: "job"},
			Name:           job.Name,
			PodNames:       podNames,
			StartTime:      startTime,
			CompletionTime: completionTime,
			Failed:         failed(job),
			Duration:       duration(job).String(),
			Image:          job.Spec.Template.Spec.Containers[0].Image,
			Message:        Message(job),
			GQLVars: struct {
				Env     string
				Team    string
				NaisJob string
			}{Env: env, Team: team, NaisJob: name},
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		if ret[i].StartTime == nil {
			return false
		}
		if ret[j].StartTime == nil {
			return true
		}

		return ret[i].StartTime.After(*ret[j].StartTime)
	})

	return ret, nil
}

func Message(job *batchv1.Job) string {
	if failed(job) {
		return fmt.Sprintf("Run failed after %d attempts", job.Status.Failed)
	}
	target := completionTarget(*job)
	if job.Status.Active > 0 {
		msg := ""
		if job.Status.Active == 1 {
			msg = "1 instance running"
		} else {
			msg = fmt.Sprintf("%d instances running", job.Status.Active)
		}
		return fmt.Sprintf("%s. %d/%d completed (%d failed %s)", msg, job.Status.Succeeded, target, job.Status.Failed, pluralize("attempt", job.Status.Failed))
	} else if job.Status.Succeeded == target {
		return fmt.Sprintf("%d/%d instances completed (%d failed %s)", job.Status.Succeeded, target, job.Status.Failed, pluralize("attempt", job.Status.Failed))
	}
	return ""
}

func pluralize(s string, count int32) string {
	if count == 1 {
		return s
	}
	return s + "s"
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
		if cs.Status == corev1.ConditionTrue && cs.Type == batchv1.JobFailed {
			return true
		}
	}
	return false
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

	authz, err := jobAuthz(naisjob)
	if err != nil {
		return nil, fmt.Errorf("getting authz: %w", err)
	}

	ret.Authz = authz

	return ret, nil
}

func naisjobStorage(u *unstructured.Unstructured, topics []*model.Topic) ([]model.Storage, error) {
	naisjob := &naisv1.Naisjob{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, naisjob); err != nil {
		return nil, fmt.Errorf("converting to application: %w", err)
	}

	ret := []model.Storage{}

	if naisjob.Spec.GCP != nil {
		for _, v := range naisjob.Spec.GCP.Buckets {
			bucket := model.Bucket{}
			if err := convert(v, &bucket); err != nil {
				return nil, fmt.Errorf("converting buckets: %w", err)
			}
			ret = append(ret, bucket)
		}
		for _, v := range naisjob.Spec.GCP.SqlInstances {
			sqlInstance := model.SQLInstance{}
			if err := convert(v, &sqlInstance); err != nil {
				return nil, fmt.Errorf("converting sqlInstance: %w", err)
			}
			if sqlInstance.Name == "" {
				sqlInstance.Name = naisjob.Name
			}
			ret = append(ret, sqlInstance)
		}

		for _, v := range naisjob.Spec.GCP.BigQueryDatasets {
			bqDataset := model.BigQueryDataset{}
			if err := convert(v, &bqDataset); err != nil {
				return nil, fmt.Errorf("converting bigQueryDataset: %w", err)
			}
			ret = append(ret, bqDataset)
		}
	}

	if naisjob.Spec.OpenSearch != nil {
		os := model.OpenSearch{
			Name:   naisjob.Spec.OpenSearch.Instance,
			Access: naisjob.Spec.OpenSearch.Access,
		}
		ret = append(ret, os)
	}

	if naisjob.Spec.Kafka != nil {
		kafka := model.Kafka{
			Name:    naisjob.Spec.Kafka.Pool,
			Streams: naisjob.Spec.Kafka.Streams,
			Topics:  topics,
		}
		ret = append(ret, kafka)
	}
	return ret, nil
}

func jobAuthz(job *naisv1.Naisjob) ([]model.Authz, error) {
	ret := []model.Authz{}
	if job.Spec.Azure != nil {
		isApp := job.Spec.Azure.Application != nil && job.Spec.Azure.Application.Enabled
		if isApp {
			azureAd := model.AzureAd{}
			if err := convert(job.Spec.Azure, &azureAd); err != nil {
				return nil, fmt.Errorf("converting azureAd: %w", err)
			}
			ret = append(ret, azureAd)
		}
	}

	if job.Spec.Maskinporten != nil && job.Spec.Maskinporten.Enabled {
		maskinporten := model.Maskinporten{}
		if err := convert(job.Spec.Maskinporten, &maskinporten); err != nil {
			return nil, fmt.Errorf("converting maskinporten: %w", err)
		}
		ret = append(ret, maskinporten)
	}

	return ret, nil
}
