package resourceusage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

// utilizationMapForEnv is a map of team -> app -> utilizationMap
type utilizationMapForEnv map[string]map[string]utilizationMap

type Updater struct {
	querier     gensql.Querier
	promClients map[string]promv1.API
	log         logrus.FieldLogger
}

const (
	cpuUsage      = `sum(rate(container_cpu_usage_seconds_total{namespace!~%q, container!~%q}[5m])) by (namespace, container)`
	cpuRequest    = `sum(kube_pod_container_resource_requests{namespace!~%q, container!~%q, resource="cpu", unit="core"}) by (namespace, container)`
	memoryUsage   = `sum(container_memory_working_set_bytes{namespace!~%q, container!~%q}) by (namespace, container)`
	memoryRequest = `sum(kube_pod_container_resource_requests{namespace!~%q, container!~%q, resource="memory", unit="byte"}) by (namespace, container)`

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

// NewUpdater creates a new resourceusage updater
func NewUpdater(promClients map[string]promv1.API, querier gensql.Querier, log logrus.FieldLogger) *Updater {
	return &Updater{
		querier:     querier,
		promClients: promClients,
		log:         log,
	}
}

func (u *Updater) UpdateResourceUsage(ctx context.Context) (rowsUpserted int, err error) {
	maxTimestamp, err := u.querier.MaxResourceUtilizationDate(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to fetch max timestamp from database: %w", err)
	}

	start, end := getQueryRange(maxTimestamp.Time)

	resourceTypes := []gensql.ResourceType{
		gensql.ResourceTypeCpu,
		gensql.ResourceTypeMemory,
	}

	for env, promClient := range u.promClients {
		log := u.log.WithField("env", env)
		for _, resourceType := range resourceTypes {
			log = log.WithField("resource_type", resourceType)
			values, err := utilizationInEnv(ctx, promClient, resourceType, start, end, log)
			if err != nil {
				log.WithError(err).Errorf("unable to fetch resource usage")
				continue
			}

			batchErrors := 0
			batch := getBatchParams(env, values)
			u.querier.ResourceUtilizationUpsert(ctx, batch).Exec(func(i int, err error) {
				if err != nil {
					batchErrors++
				}
			})
			log.WithFields(logrus.Fields{
				"num_rows":   len(batch),
				"num_errors": batchErrors,
			}).Debugf("batch upsert")
			rowsUpserted += len(batch) - batchErrors
		}
	}

	return rowsUpserted, nil
}

// utilizationInEnv returns resource utilization (usage and request) for all teams and apps in a given env
func utilizationInEnv(ctx context.Context, promClient promv1.API, resourceType gensql.ResourceType, start, end time.Time, log logrus.FieldLogger) (utilizationMapForEnv, error) {
	utilization := make(utilizationMapForEnv)
	usageQuery, requestQuery := getQueries(resourceType)

	log.WithField("query", usageQuery).Debugf("fetch usage data from prometheus")
	if usage, err := rangedQuery(ctx, promClient, usageQuery, start, end); err != nil {
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
					utilization[team][app] = initUtilizationMap(resourceType, start, end)
				}

				utilization[team][app][ts].Usage = float64(val.Value)
			}
		}
	}

	log.WithField("query", requestQuery).Debugf("fetch request data from prometheus")
	if request, err := rangedQuery(ctx, promClient, requestQuery, start, end); err != nil {
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
					utilization[team][app] = initUtilizationMap(resourceType, start, end)
				}

				utilization[team][app][ts].Request = float64(val.Value)
			}
		}
	}
	return utilization, nil
}

// getBatchParams converts ResourceUtilization to ResourceUtilizationUpsertParams
func getBatchParams(env string, utilization utilizationMapForEnv) []gensql.ResourceUtilizationUpsertParams {
	params := make([]gensql.ResourceUtilizationUpsertParams, 0)
	for team, apps := range utilization {
		for app, timestamps := range apps {
			for _, value := range timestamps {
				ts := &pgtype.Timestamptz{}
				_ = ts.Scan(value.Timestamp.UTC())

				params = append(params, gensql.ResourceUtilizationUpsertParams{
					Timestamp:    *ts,
					Env:          env,
					Team:         team,
					App:          app,
					ResourceType: gensql.ResourceType(strings.ToLower(string(value.Resource))),
					Usage:        value.Usage,
					Request:      value.Request,
				})
			}
		}
	}
	return params
}

// getQueries returns the prometheus queries for the given resource type
func getQueries(resourceType gensql.ResourceType) (usageQuery, requestQuery string) {
	if resourceType == gensql.ResourceTypeCpu {
		usageQuery = cpuUsage
		requestQuery = cpuRequest
	} else {
		usageQuery = memoryUsage
		requestQuery = memoryRequest
	}
	ignoreNamespaces := strings.Join(namespacesToIgnore, "|") + "|"
	ignoreContainers := strings.Join(containersToIgnore, "|") + "|"
	return fmt.Sprintf(usageQuery, ignoreNamespaces, ignoreContainers), fmt.Sprintf(requestQuery, ignoreNamespaces, ignoreContainers)
}

// getQueryRange returns the start and end time in UTC for a query, based on the given start time
func getQueryRange(start time.Time) (time.Time, time.Time) {
	now := time.Now()
	if start.IsZero() {
		start = now.AddDate(0, 0, -30)
	}

	end := start.Add(7 * 24 * time.Hour)
	if end.After(now) {
		end = now
	}

	return normalizeTime(start), normalizeTime(end)
}

// initUtilizationMap initializes a utilizationMap with the given time range without gaps
func initUtilizationMap(resourceType gensql.ResourceType, start, end time.Time) utilizationMap {
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
			Resource:  model.ResourceType(strings.ToUpper(string(resourceType))),
		}
	}
	return utilization
}

// rangedQuery queries prometheus for the given query in the given time range
func rangedQuery(ctx context.Context, client promv1.API, query string, start, end time.Time) (prom.Matrix, error) {
	value, warnings, err := client.QueryRange(ctx, query, promv1.Range{
		Start: start,
		End:   end,
		Step:  rangedQueryStep,
	})
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		return nil, fmt.Errorf("prometheus query warnings: %s", strings.Join(warnings, ", "))
	}

	matrix, ok := value.(prom.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected prometheus matrix, got %T", value)
	}

	return matrix, nil
}
