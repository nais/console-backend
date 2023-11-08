package resourceusage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type Client interface {
	// UtilizationForApp returns resource utilization for the given app, in the given time range
	UtilizationForApp(ctx context.Context, resource model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error)
}

type clusterName string

type client struct {
	promClients map[clusterName]promv1.API
	log         logrus.FieldLogger
}

type utilizationMap map[time.Time]*model.ResourceUtilization

const (
	cpuAppUsageQuery      = `max(rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[5m]))`
	cpuAppRequestQuery    = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="cpu", unit="core"})`
	memoryAppUsageQuery   = `max(container_memory_working_set_bytes{namespace=%q, container=%q})`
	memoryAppRequestQuery = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="memory", unit="byte"})`
)

// New creates a new resourceusage client
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

func (c *client) UtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error) {
	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	usageQuery, requestQuery := getQueries(resourceType, team, app)
	utilization := make(utilizationMap)

	samples, err := rangedQuery(ctx, promClient, usageQuery, start, end)
	if err != nil {
		return nil, err
	}
	for _, val := range samples {
		ts := val.Timestamp.Time()
		utilization[ts] = &model.ResourceUtilization{
			Timestamp: ts,
			Resource:  resourceType,
			Usage:     float64(val.Value),
		}
	}

	samples, err = rangedQuery(ctx, promClient, requestQuery, start, end)
	if err != nil {
		return nil, err
	}
	for _, val := range samples {
		ts := val.Timestamp.Time()
		if _, exists := utilization[ts]; !exists {
			continue
		}
		utilization[ts].Request = float64(val.Value)
	}

	return mapToResourceUtilization(utilization), nil
}

// mapToResourceUtilization converts a utilizationMap to []model.ResourceUtilization, sorted by timestamp
func mapToResourceUtilization(utilization utilizationMap) []model.ResourceUtilization {
	timestamps := make([]time.Time, 0)
	for ts := range utilization {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	ret := make([]model.ResourceUtilization, 0)
	for _, ts := range timestamps {
		if _, exists := utilization[ts]; !exists {
			continue
		}
		ret = append(ret, *utilization[ts])
	}
	return ret
}

// getQueries returns the prometheus queries for the given resource type
func getQueries(resourceType model.ResourceType, team, app string) (usageQuery, requestQuery string) {
	if resourceType == model.ResourceTypeCPU {
		usageQuery = cpuAppUsageQuery
		requestQuery = cpuAppRequestQuery
	} else {
		usageQuery = memoryAppUsageQuery
		requestQuery = memoryAppRequestQuery
	}
	return fmt.Sprintf(usageQuery, team, app), fmt.Sprintf(requestQuery, team, app)
}

// rangedQuery queries prometheus for the given query, in the given time range.
func rangedQuery(ctx context.Context, client promv1.API, query string, start, end time.Time) ([]prom.SamplePair, error) {
	value, _, err := client.QueryRange(ctx, query, promv1.Range{
		Start: time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()),
		End:   time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location()),
		Step:  getStep(start, end),
	})
	if err != nil {
		return nil, err
	}

	matrix, ok := value.(prom.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected prometheus matrix, got %T", value)
	}

	if len(matrix) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	return matrix[0].Values, nil
}

// getStep returns the step to use for the given time range
func getStep(start, end time.Time) time.Duration {
	step := 24 * time.Hour
	if end.Sub(start) < 7*24*time.Hour {
		step = time.Hour
	}
	return step
}
