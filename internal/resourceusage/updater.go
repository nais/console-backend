package resourceusage

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/sirupsen/logrus"
)

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
			values, err := c.UtilizationInEnv(ctx, resourceType, env, start, end)
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

// getBatchParams converts ResourceUtilization to ResourceUtilizationUpsertParams
func getBatchParams(env string, values []ResourceUtilization) []gensql.ResourceUtilizationUpsertParams {
	params := make([]gensql.ResourceUtilizationUpsertParams, 0)
	for _, value := range values {
		params = append(params, gensql.ResourceUtilizationUpsertParams{
			Date:         pgtype.Timestamptz{Time: value.Timestamp.In(time.UTC), Valid: true},
			Env:          env,
			Team:         value.Team,
			App:          value.App,
			ResourceType: gensql.ResourceType(strings.ToLower(string(value.Resource))),
			Usage:        value.Usage,
			Request:      value.Request,
		})
	}
	return params
}
