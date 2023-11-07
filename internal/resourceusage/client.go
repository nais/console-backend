package resourceusage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type Client interface {
	UtilizationForApp(ctx context.Context, resource model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error)
}

type clusterName string

type client struct {
	promClients map[clusterName]promv1.API
	querier     gensql.Querier
	log         logrus.FieldLogger
}

type utilizationMap map[time.Time]*metrics

type metrics struct {
	request float64
	pods    map[string]float64
}

const (
	cpuAppUsageQuery      = `rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[5m])`
	cpuAppRequestQuery    = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="cpu", unit="core"})`
	memoryAppUsageQuery   = `container_memory_working_set_bytes{namespace=%q, container=%q}`
	memoryAppRequestQuery = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="memory", unit="byte"})`
)

func New(clusters []string, tenant string, querier gensql.Querier, log logrus.FieldLogger) (Client, error) {
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
		querier:     querier,
		log:         log,
	}, nil
}

func (c *client) UtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error) {
	step := 24 * time.Hour
	if end.Sub(start) < 7*24*time.Hour {
		step = time.Hour
	}

	promClient, exists := c.promClients[clusterName(env)]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	usageQuery, requestQuery := getAppQueries(resourceType, team, app)
	series, err := rangedQuery(ctx, promClient, usageQuery, start, end, step)
	if err != nil {
		return nil, err
	}

	utilization := make(utilizationMap)
	for _, pod := range series {
		podName := string(pod.Metric["pod"])
		if podName == "" {
			podName = "<unknown>"
		}
		for _, val := range pod.Values {
			ts := val.Timestamp.Time()
			if _, exists := utilization[ts]; !exists {
				utilization[ts] = &metrics{
					pods: make(map[string]float64),
				}
			}
			utilization[ts].pods[podName] = float64(val.Value)
		}
	}

	series, err = rangedQuery(ctx, promClient, requestQuery, start, end, step)
	if err != nil {
		return nil, err
	}
	for _, val := range series[0].Values {
		ts := val.Timestamp.Time()
		v := float64(val.Value)
		if _, exists := utilization[ts]; !exists {
			continue
		}
		utilization[ts].request = v
	}

	return utilizationMapToResourceUtilization(utilization, resourceType), nil
}

func utilizationMapToResourceUtilization(utilization utilizationMap, resourceType model.ResourceType) []model.ResourceUtilization {
	timestamps := make([]time.Time, 0)
	for ts := range utilization {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	ret := make([]model.ResourceUtilization, 0)
	for _, ts := range timestamps {
		pods := make([]model.ResourceUtilizationPodUsage, 0)
		for podName, usage := range utilization[ts].pods {
			pods = append(pods, model.ResourceUtilizationPodUsage{
				Pod:   podName,
				Usage: usage,
			})
		}

		req := utilization[ts].request
		ret = append(ret, model.ResourceUtilization{
			Resource:     resourceType,
			Timestamp:    ts,
			Request:      req,
			RequestTotal: req * float64(len(pods)),
			Pods:         pods,
		})
	}
	return ret
}

func getAppQueries(resourceType model.ResourceType, team string, app string) (usageQuery, requestQuery string) {
	if resourceType == model.ResourceTypeCPU {
		usageQuery = cpuAppUsageQuery
		requestQuery = cpuAppRequestQuery
	} else {
		usageQuery = memoryAppUsageQuery
		requestQuery = memoryAppRequestQuery
	}
	return fmt.Sprintf(usageQuery, team, app), fmt.Sprintf(requestQuery, team, app)
}

func rangedQuery(ctx context.Context, client promv1.API, query string, start, end time.Time, step time.Duration) (prom.Matrix, error) {
	value, _, err := client.QueryRange(ctx, query, promv1.Range{
		Start: time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC),
		End:   time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC),
		Step:  step,
	})
	if err != nil {
		return nil, err
	}

	return value.(prom.Matrix), nil
}
