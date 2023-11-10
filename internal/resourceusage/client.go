package resourceusage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type Client interface {
	// UpdateResourceUsage will update resource usage data for all teams in all environments.
	UpdateResourceUsage(ctx context.Context) (rowsUpserted int, err error)

	// UtilizationForApp returns resource utilization (usage and request) for the given app, in the given time range
	UtilizationForApp(ctx context.Context, resource model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error)

	// UtilizationForTeam returns resource utilization (usage and request) for a given team in the given time range
	UtilizationForTeam(ctx context.Context, resource model.ResourceType, env, team string, start, end time.Time) ([]model.ResourceUtilization, error)
}

type utilizationMap map[time.Time]*model.ResourceUtilization

type client struct {
	clusters    []string
	querier     gensql.Querier
	promClients map[string]promv1.API
	log         logrus.FieldLogger
}

const (
	cpuUsageForTeam      = `sum(rate(container_cpu_usage_seconds_total{namespace=%q, container!~%q}[5m]))`
	cpuRequestForTeam    = `sum(kube_pod_container_resource_requests{namespace=%q, container!~%q, resource="cpu", unit="core"})`
	memoryUsageForTeam   = `sum(container_memory_working_set_bytes{namespace=%q, container!~%q})`
	memoryRequestForTeam = `sum(kube_pod_container_resource_requests{namespace=%q, container!~%q, resource="memory", unit="byte"})`

	cpuUsageForApp      = `max(rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[5m]))`
	cpuRequestForApp    = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="cpu", unit="core"})`
	memoryUsageForApp   = `max(container_memory_working_set_bytes{namespace=%q, container=%q})`
	memoryRequestForApp = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="memory", unit="byte"})`

	rangedQueryStep = time.Hour
)

// New creates a new resourceusage client
func New(clusters []string, tenant string, querier gensql.Querier, log logrus.FieldLogger) (Client, error) {
	promClients := map[string]promv1.API{}
	for _, cluster := range clusters {
		promClient, err := api.NewClient(api.Config{
			Address: fmt.Sprintf("https://prometheus.%s.%s.cloud.nais.io", cluster, tenant),
		})
		if err != nil {
			return nil, err
		}
		promClients[cluster] = promv1.NewAPI(promClient)
	}

	return &client{
		clusters:    clusters,
		querier:     querier,
		promClients: promClients,
		log:         log,
	}, nil
}

func (c *client) UtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error) {
	start = normalizeTime(start)
	end = normalizeTime(end)
	log := c.log.WithFields(logrus.Fields{
		"env":           env,
		"team":          team,
		"app":           app,
		"resource_type": resourceType,
	})

	promClient, exists := c.promClients[env]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	utilization := initUtilizationMap(resourceType, start, end)
	usageQuery, requestQuery := getAppQueries(resourceType, team, app)

	usage, err := rangedQuery(ctx, promClient, usageQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for usage data")
	} else if len(usage) > 0 {
		for _, val := range usage[0].Values {
			ts := val.Timestamp.Time().UTC()
			v := float64(val.Value)
			utilization[ts].Usage = v
			utilization[ts].UsageCost = cost(resourceType, v)
		}
	} else {
		log.Warningf("no usage data found")
	}

	request, err := rangedQuery(ctx, promClient, requestQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for request data")
	} else if len(request) > 0 {
		for _, val := range request[0].Values {
			ts := val.Timestamp.Time().UTC()
			v := float64(val.Value)
			utilization[ts].Request = v
			utilization[ts].RequestCost = cost(resourceType, v)
			utilization[ts].RequestCostOverage = utilization[ts].RequestCost - utilization[ts].UsageCost
		}
	} else {
		log.Warningf("no request data found")
	}

	return utilizationMapToSlice(utilization), nil
}

func (c *client) UtilizationForTeam(ctx context.Context, resourceType model.ResourceType, env, team string, start, end time.Time) ([]model.ResourceUtilization, error) {
	start = normalizeTime(start)
	end = normalizeTime(end)
	log := c.log.WithFields(logrus.Fields{
		"team":          team,
		"resource_type": resourceType,
	})

	promClient, exists := c.promClients[env]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	utilization := initUtilizationMap(resourceType, start, end)
	usageQuery, requestQuery := getTeamQueries(resourceType, team)

	usage, err := rangedQuery(ctx, promClient, usageQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for usage data")
	} else if len(usage) > 0 {
		for _, val := range usage[0].Values {
			ts := val.Timestamp.Time().UTC()
			v := float64(val.Value)
			utilization[ts].Usage = v
			utilization[ts].UsageCost = cost(resourceType, v)
		}
	} else {
		log.Warningf("no usage data found")
	}

	request, err := rangedQuery(ctx, promClient, requestQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for request data")
	} else if len(request) > 0 {
		for _, val := range request[0].Values {
			ts := val.Timestamp.Time().UTC()
			v := float64(val.Value)
			utilization[ts].Request = v
			utilization[ts].RequestCost = cost(resourceType, v)
			utilization[ts].RequestCostOverage = utilization[ts].RequestCost - utilization[ts].UsageCost
		}
	} else {
		log.Warningf("no request data found")
	}

	return utilizationMapToSlice(utilization), nil
}

// initUtilizationMap initializes a utilizationMap with the given time range without gaps
func initUtilizationMap(resourceType model.ResourceType, start, end time.Time) utilizationMap {
	timestamps := make([]time.Time, 0)
	ts := start
	for ; ts.Before(end); ts = ts.Add(rangedQueryStep) {
		timestamps = append(timestamps, ts)
	}
	timestamps = append(timestamps, ts)
	utilization := make(utilizationMap)
	for _, ts := range timestamps {
		utilization[ts] = &model.ResourceUtilization{
			Timestamp: ts,
			Resource:  resourceType,
		}
	}
	return utilization
}

// getAppQueries returns the prometheus queries for the given resource type
func getAppQueries(resourceType model.ResourceType, team, app string) (usageQuery, requestQuery string) {
	if resourceType == model.ResourceTypeCPU {
		usageQuery = cpuUsageForApp
		requestQuery = cpuRequestForApp
	} else {
		usageQuery = memoryUsageForApp
		requestQuery = memoryRequestForApp
	}
	return fmt.Sprintf(usageQuery, team, app), fmt.Sprintf(requestQuery, team, app)
}

// getTeamQueries returns the prometheus queries for the given team and resource type
func getTeamQueries(resourceType model.ResourceType, team string) (usageQuery, requestQuery string) {
	if resourceType == model.ResourceTypeCPU {
		usageQuery = cpuUsageForTeam
		requestQuery = cpuRequestForTeam
	} else {
		usageQuery = memoryUsageForTeam
		requestQuery = memoryRequestForTeam
	}
	ignoreContainers := strings.Join(containersToIgnore, "|") + "|"
	return fmt.Sprintf(usageQuery, team, ignoreContainers), fmt.Sprintf(requestQuery, team, ignoreContainers)
}

// rangedQuery queries prometheus for the given query in the given time range
func rangedQuery(ctx context.Context, client promv1.API, query string, start, end time.Time) (prom.Matrix, error) {
	value, _, err := client.QueryRange(ctx, query, promv1.Range{
		Start: start,
		End:   end,
		Step:  rangedQueryStep,
	})
	if err != nil {
		return nil, err
	}

	matrix, ok := value.(prom.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected prometheus matrix, got %T", value)
	}

	return matrix, nil
}

// normalizeTime will truncate a time.Time down to the hour, and return it as UTC
func normalizeTime(ts time.Time) time.Time {
	return ts.Truncate(time.Hour).UTC()
}

// utilizationMapToSlice converts a utilizationMap to a slice of ResourceUtilization, sorted by the timestamp
func utilizationMapToSlice(util utilizationMap) []model.ResourceUtilization {
	ret := make([]model.ResourceUtilization, 0)
	for _, u := range util {
		ret = append(ret, *u)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Timestamp.Before(ret[j].Timestamp)
	})
	return ret
}

// cost calculates the cost for the given resource type
func cost(resourceType model.ResourceType, value float64) (cost float64) {
	if resourceType == model.ResourceTypeCPU {
		cost = 131.0 / 30.0 * value
	} else {
		cost = 18.0 / 1024 / 1024 / 1024 / 30.0 * value
	}

	return cost / 24.0
}
