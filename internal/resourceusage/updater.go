package resourceusage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/sirupsen/logrus"
)

// utilizationMapForEnv is a map of team -> app -> utilizationMap
type utilizationMapForEnv map[string]map[string]utilizationMap

func (c *client) UpdateResourceUsage(ctx context.Context) (rowsUpserted int) {
	start := normalizeTime(time.Now().AddDate(0, 0, -30))
	end := start.Add(24 * time.Hour)

	resourceTypes := []model.ResourceType{
		model.ResourceTypeCPU,
		model.ResourceTypeMemory,
	}

	for _, env := range c.clusters {
		log := c.log.WithField("env", env)
		for _, resourceType := range resourceTypes {
			log = log.WithField("resource_type", resourceType)
			log.Debugf("fetch data from prometheus")
			values, err := c.utilizationInEnv(ctx, resourceType, env, start, end)
			if err != nil {
				log.WithError(err).Errorf("unable to fetch resource usage")
				continue
			}

			batchErrors := 0
			batch := getBatchParams(env, values)
			c.querier.ResourceUtilizationUpsert(ctx, batch).Exec(func(i int, err error) {
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

	return rowsUpserted
}

// utilizationInEnv returns resource utilization (usage and request) for all teams and apps in a given env
func (c *client) utilizationInEnv(ctx context.Context, resourceType model.ResourceType, env string, start, end time.Time) (utilizationMapForEnv, error) {
	start = normalizeTime(start)
	end = normalizeTime(end)
	log := c.log.WithFields(logrus.Fields{
		"env":           env,
		"resource_type": resourceType,
	})

	utilization := make(utilizationMapForEnv)
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
					utilization[team][app] = initUtilizationMap(resourceType, start, end)
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
				params = append(params, gensql.ResourceUtilizationUpsertParams{
					Date:         pgtype.Timestamptz{Time: value.Timestamp.In(time.UTC), Valid: true},
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
