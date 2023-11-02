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
	UtilizationForApp(ctx context.Context, resource model.ResourceType, resolution model.Resolution, env, team, app string, start, end time.Time, step time.Duration) ([]model.ResourceUtilization, error)
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

const (
	cpuUsageQuery      = `max(rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[1h]))`
	cpuRequestQuery    = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="cpu", unit="core"})`
	memoryUsageQuery   = `max(container_memory_usage_bytes{namespace=%q, container=%q})`
	memoryRequestQuery = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="memory", unit="byte"})`
)

func (c *client) UtilizationForApp(ctx context.Context, resourceType model.ResourceType, resolution model.Resolution, env, team, app string, start, end time.Time, step time.Duration) ([]model.ResourceUtilization, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	utilization := make(map[time.Time]*model.ResourceUtilization)
	timestamps := make([]time.Time, 0)

	usageQuery, requestQuery := getQueries(resourceType, team, app)
	values, err := rangedQuery(ctx, promClient, usageQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	for _, val := range values {
		timestamps = append(timestamps, val.Timestamp.Time())
		utilization[val.Timestamp.Time()] = &model.ResourceUtilization{
			Timestamp: val.Timestamp.Time(),
			Usage:     float64(val.Value),
			UsageCost: cost(resourceType, float64(val.Value), resolution),
		}
	}

	values, err = rangedQuery(ctx, promClient, requestQuery, start, end, step)
	if err != nil {
		return nil, err
	}
	for _, val := range values {
		if _, exists := utilization[val.Timestamp.Time()]; !exists {
			continue
		}
		u := utilization[val.Timestamp.Time()]
		u.Request = float64(val.Value)
		u.RequestCost = cost(resourceType, float64(val.Value), resolution)
		u.RequestedFactor = u.Request / u.Usage
	}

	ret := make([]model.ResourceUtilization, 0)
	for _, t := range timestamps {
		ret = append(ret, *utilization[t])
	}
	return ret, nil
}

func getQueries(resourceType model.ResourceType, team, app string) (usageQuery, requestQuery string) {
	if resourceType == model.ResourceTypeCPU {
		usageQuery = cpuUsageQuery
		requestQuery = cpuRequestQuery
	} else {
		usageQuery = memoryUsageQuery
		requestQuery = memoryRequestQuery
	}

	return fmt.Sprintf(usageQuery, team, app), fmt.Sprintf(requestQuery, team, app)
}

func cost(resourceType model.ResourceType, value float64, resolution model.Resolution) (cost float64) {
	if resourceType == model.ResourceTypeCPU {
		cost = 131.0 / 31.0 * value
	} else {
		cost = 18.0 / 1024 / 1024 / 1024 / 31.0 * value
	}

	if resolution == model.ResolutionHourly {
		cost /= 24.0
	}

	return cost
}

func rangedQuery(ctx context.Context, client promv1.API, query string, start, end time.Time, step time.Duration) ([]prom.SamplePair, error) {
	value, _, err := client.QueryRange(ctx, query, promv1.Range{
		Start: time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC),
		End:   time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC),
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
