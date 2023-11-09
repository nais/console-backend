package resourceusage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type (
	// envUtilizationMap is a map of team -> app -> utilizationMap
	envUtilizationMap map[string]map[string]utilizationMap
	utilizationMap    map[time.Time]*ResourceUtilization
)

type Client interface {
	// UpdateResourceUsage will update resource usage data for all teams in all environments.
	UpdateResourceUsage(ctx context.Context) (rowsUpserted int)

	// UtilizationForApp returns resource utilization (usage and request) for the given app, in the given time range
	UtilizationForApp(ctx context.Context, resource model.ResourceType, env, team, app string, start, end time.Time) ([]ResourceUtilization, error)

	// UtilizationInEnv returns resource utilization (usage and request) for all teams and apps in a given env
	UtilizationInEnv(ctx context.Context, resourceType model.ResourceType, env string, start, end time.Time) ([]ResourceUtilization, error)
}

type ResourceUtilization struct {
	*model.ResourceUtilization
	Team string
	App  string
}

type client struct {
	clusters    []string
	querier     gensql.Querier
	promClients map[string]promv1.API
	log         logrus.FieldLogger
}

const (
	cpuUsageForEnv      = `max(rate(container_cpu_usage_seconds_total{namespace!~%q, container!~%q}[5m])) by (namespace, container)`
	cpuRequestForEnv    = `max(kube_pod_container_resource_requests{namespace!~%q, container!~%q, resource="cpu", unit="core"}) by (namespace, container)`
	memoryUsageForEnv   = `max(container_memory_working_set_bytes{namespace!~%q, container!~%q}) by (namespace, container)`
	memoryRequestForEnv = `max(kube_pod_container_resource_requests{namespace!~%q, container!~%q, resource="memory", unit="byte"}) by (namespace, container)`

	cpuUsageForApp      = `max(rate(container_cpu_usage_seconds_total{namespace=%q, container=%q}[5m]))`
	cpuRequestForApp    = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="cpu", unit="core"})`
	memoryUsageForApp   = `max(container_memory_working_set_bytes{namespace=%q, container=%q})`
	memoryRequestForApp = `max(kube_pod_container_resource_requests{namespace=%q, container=%q, resource="memory", unit="byte"})`

	rangedQueryStep = time.Hour
)

var (
	namespacesToIgnore = []string{
		"default",
		"kube-system",
		"kyverno",
		"linkerd",
		"nais",
		"nais-system",
	}

	containersToIgnore = []string{
		"cloudsql-proxy",
		"elector",
		"linkerd-proxy",
		"secure-logs-configmap-reload",
		"secure-logs-fluentd",
		"vks-sidecar",
		"wonderwall",
	}
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

func (c *client) UtilizationInEnv(ctx context.Context, resourceType model.ResourceType, env string, start, end time.Time) ([]ResourceUtilization, error) {
	start = normalizeTime(start)
	end = normalizeTime(end)
	log := c.log.WithFields(logrus.Fields{
		"env":           env,
		"resource_type": resourceType,
	})

	utilization := make(envUtilizationMap)
	promClient, exists := c.promClients[env]
	if !exists {
		return nil, fmt.Errorf("no prometheus client for cluster: %q", env)
	}

	usageQuery, requestQuery := getEnvQueries(resourceType)
	usage, err := rangedQuery(ctx, promClient, usageQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for usage data")
	} else {
		for _, sample := range usage {
			for _, val := range sample.Values {
				ts := val.Timestamp.Time().UTC()
				team := string(sample.Metric["namespace"])
				app := string(sample.Metric["container"])

				if _, exists := utilization[team]; !exists {
					utilization[team] = make(map[string]utilizationMap)
				}

				if _, exists := utilization[team][app]; !exists {
					utilization[team][app] = initUtilizationMap(resourceType, team, app, start, end)
				}

				utilization[team][app][ts].Usage = float64(val.Value)
			}
		}
	}

	request, err := rangedQuery(ctx, promClient, requestQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for request data")
	} else {
		for _, sample := range request {
			for _, val := range sample.Values {
				ts := val.Timestamp.Time().UTC()
				team := string(sample.Metric["namespace"])
				app := string(sample.Metric["container"])

				if _, exists := utilization[team]; !exists {
					utilization[team] = make(map[string]utilizationMap)
				}

				if _, exists := utilization[team][app]; !exists {
					utilization[team][app] = initUtilizationMap(resourceType, team, app, start, end)
				}

				utilization[team][app][ts].Request = float64(val.Value)
			}
		}
	}

	ret := make([]ResourceUtilization, 0)
	for _, apps := range utilization {
		for _, timestamps := range apps {
			for _, util := range timestamps {
				ret = append(ret, *util)
			}
		}
	}
	return ret, nil
}

func (c *client) UtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end time.Time) ([]ResourceUtilization, error) {
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

	utilization := initUtilizationMap(resourceType, team, app, start, end)
	usageQuery, requestQuery := getAppQueries(resourceType, team, app)

	usage, err := rangedQuery(ctx, promClient, usageQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for usage data")
	} else if len(usage) > 0 {
		for _, val := range usage[0].Values {
			utilization[val.Timestamp.Time().UTC()].Usage = float64(val.Value)
		}
	} else {
		log.Warningf("no usage data found")
	}

	request, err := rangedQuery(ctx, promClient, requestQuery, start, end)
	if err != nil {
		log.WithError(err).Errorf("unable to query prometheus for request data")
	} else if len(request) > 0 {
		for _, val := range request[0].Values {
			utilization[val.Timestamp.Time().UTC()].Request = float64(val.Value)
		}
	} else {
		log.Warningf("no request data found")
	}

	ret := make([]ResourceUtilization, 0)
	for _, ut := range utilization {
		ret = append(ret, *ut)
	}
	return ret, nil
}

// initUtilizationMap initializes a utilizationMap with the given time range without gaps
func initUtilizationMap(resourceType model.ResourceType, team, app string, start, end time.Time) utilizationMap {
	timestamps := make([]time.Time, 0)
	ts := start
	for ; ts.Before(end); ts = ts.Add(rangedQueryStep) {
		timestamps = append(timestamps, ts)
	}
	timestamps = append(timestamps, ts)
	utilization := make(utilizationMap)
	for _, ts := range timestamps {
		utilization[ts] = &ResourceUtilization{
			ResourceUtilization: &model.ResourceUtilization{
				Timestamp: ts,
				Resource:  resourceType,
				Request:   0,
				Usage:     0,
			},
			Team: team,
			App:  app,
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

// getEnvQueries returns the prometheus queries for the given resource type
func getEnvQueries(resourceType model.ResourceType) (usageQuery, requestQuery string) {
	if resourceType == model.ResourceTypeCPU {
		usageQuery = cpuUsageForEnv
		requestQuery = cpuRequestForEnv
	} else {
		usageQuery = memoryUsageForEnv
		requestQuery = memoryRequestForEnv
	}
	ignoreNamespaces := strings.Join(namespacesToIgnore, "|") + "|"
	ignoreContainers := strings.Join(containersToIgnore, "|") + "|"
	return fmt.Sprintf(usageQuery, ignoreNamespaces, ignoreContainers), fmt.Sprintf(requestQuery, ignoreNamespaces, ignoreContainers)
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
