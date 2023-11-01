package resourceusage

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/console-backend/internal/graph/model"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type clusterName string

type Client interface {
	CPUUtilizationForApp(ctx context.Context, env, team, app string, now time.Time, dur, step time.Duration) ([]model.ResourceUtilization, error)
	MemoryUtilizationForApp(ctx context.Context, env, team, app string, now time.Time, dur, step time.Duration) ([]model.ResourceUtilization, error)
}

type client struct {
	promClients map[clusterName]promv1.API
	log         logrus.FieldLogger
}

func New(clusters []string, tenant string, log logrus.FieldLogger) (Client, error) {
	promClients := map[clusterName]promv1.API{}
	for _, cluster := range clusters {
		promClient, err := api.NewClient(api.Config{
			Address: fmt.Sprintf("https://prometheus.%s.%s.cloud.nais.io", cluster, tenant),
		})
		if err != nil {
			return nil, err
		}
		promClients[clusterName(cluster)] = promv1.NewAPI(promClient)
	}

	return &client{
		promClients: promClients,
		log:         log,
	}, nil
}

func (c *client) CPUUtilizationForApp(ctx context.Context, env, team, app string, now time.Time, dur, step time.Duration) ([]model.ResourceUtilization, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	utilization := make(map[time.Time]*model.ResourceUtilization)
	timestamps := make([]time.Time, 0)
	query := fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[1h]))`,
		team,
		app,
	)
	values, err := rangedQuery(ctx, promClient, query, now, dur, step)
	if err != nil {
		return nil, err
	}

	for _, val := range values {
		timestamps = append(timestamps, val.Timestamp.Time())
		utilization[val.Timestamp.Time()] = &model.ResourceUtilization{
			Timestamp: val.Timestamp.Time(),
			Usage:     float64(val.Value),
			UsageCost: 131.0 / 31.0 / 24.0 * float64(val.Value),
		}
	}

	query = fmt.Sprintf(
		`max(max_over_time(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="cpu", unit="core"}[5m]))`,
		team,
		app,
	)
	values, err = rangedQuery(ctx, promClient, query, now, dur, step)
	if err != nil {
		return nil, err
	}
	for _, val := range values {
		if _, exists := utilization[val.Timestamp.Time()]; !exists {
			continue
		}
		u := utilization[val.Timestamp.Time()]
		u.Request = float64(val.Value)
		u.RequestCost = 131.0 / 31.0 / 24.0 * float64(val.Value)
		u.RequestedFactor = u.Request / u.Usage
	}

	ret := make([]model.ResourceUtilization, 0)
	for _, t := range timestamps {
		ret = append(ret, *utilization[t])
	}
	return ret, nil
}

func (c *client) MemoryUtilizationForApp(ctx context.Context, env, team, app string, now time.Time, dur, step time.Duration) ([]model.ResourceUtilization, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	utilization := make(map[time.Time]*model.ResourceUtilization)
	timestamps := make([]time.Time, 0)
	query := fmt.Sprintf(
		`sum(container_memory_usage_bytes{namespace=%q, container=%q})`,
		team,
		app,
	)
	values, err := rangedQuery(ctx, promClient, query, now, dur, step)
	if err != nil {
		return nil, err
	}

	for _, val := range values {
		timestamps = append(timestamps, val.Timestamp.Time())
		utilization[val.Timestamp.Time()] = &model.ResourceUtilization{
			Timestamp: val.Timestamp.Time(),
			Usage:     float64(val.Value),
			UsageCost: 18.0 / 1024 / 1024 / 1024 / 31.0 / 24.0 * float64(val.Value),
		}
	}

	query = fmt.Sprintf(
		`sum(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="memory", unit="byte"})`,
		team,
		app,
	)
	values, err = rangedQuery(ctx, promClient, query, now, dur, step)
	if err != nil {
		return nil, err
	}
	for _, val := range values {
		if _, exists := utilization[val.Timestamp.Time()]; !exists {
			continue
		}
		u := utilization[val.Timestamp.Time()]
		u.Request = float64(val.Value)
		u.RequestCost = 18.0 / 1024 / 1024 / 1024 / 31.0 / 24.0 * float64(val.Value)
		u.RequestedFactor = u.Request / u.Usage
	}

	ret := make([]model.ResourceUtilization, 0)
	for _, t := range timestamps {
		ret = append(ret, *utilization[t])
	}
	return ret, nil
}

func rangedQuery(ctx context.Context, client promv1.API, query string, ts time.Time, dur, step time.Duration) ([]prom.SamplePair, error) {
	to := time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, ts.Location())
	from := to.Add(-dur)
	value, _, err := client.QueryRange(ctx, query, promv1.Range{
		Start: from,
		End:   to,
		Step:  step,
	})
	if err != nil {
		return nil, err
	}

	if len(value.(prom.Matrix)) == 0 {
		return make([]prom.SamplePair, 0), nil
	}

	return value.(prom.Matrix)[0].Values, nil
}
