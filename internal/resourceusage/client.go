package resourceusage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
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
	startTs := pgtype.Timestamptz{}
	err := startTs.Scan(normalizeTime(start))
	if err != nil {
		return nil, err
	}

	endTs := pgtype.Timestamptz{}
	err = endTs.Scan(normalizeTime(end))
	if err != nil {
		return nil, err
	}

	rows, err := c.querier.ResourceUtilizationForApp(ctx, gensql.ResourceUtilizationForAppParams{
		Env:          env,
		Team:         team,
		App:          app,
		ResourceType: gensql.ResourceType(strings.ToLower(string(resourceType))),
		Start:        startTs,
		End:          endTs,
	})
	if err != nil {
		return nil, err
	}

	data := make([]model.ResourceUtilization, 0)
	for _, row := range rows {
		usageCost := cost(resourceType, row.Usage)
		requestCost := cost(resourceType, row.Request)
		data = append(data, model.ResourceUtilization{
			Resource:           resourceType,
			Timestamp:          row.Timestamp.Time.UTC(),
			Usage:              row.Usage,
			UsageCost:          usageCost,
			Request:            row.Request,
			RequestCost:        requestCost,
			RequestCostOverage: requestCost - usageCost,
		})
	}

	return data, nil
}

func (c *client) UtilizationForTeam(ctx context.Context, resourceType model.ResourceType, env, team string, start, end time.Time) ([]model.ResourceUtilization, error) {
	startTs := pgtype.Timestamptz{}
	err := startTs.Scan(normalizeTime(start))
	if err != nil {
		return nil, err
	}

	endTs := pgtype.Timestamptz{}
	err = endTs.Scan(normalizeTime(end))
	if err != nil {
		return nil, err
	}

	rows, err := c.querier.ResourceUtilizationForTeam(ctx, gensql.ResourceUtilizationForTeamParams{
		Env:          env,
		Team:         team,
		ResourceType: gensql.ResourceType(strings.ToLower(string(resourceType))),
		Start:        startTs,
		End:          endTs,
	})
	if err != nil {
		return nil, err
	}

	data := make([]model.ResourceUtilization, 0)
	for _, row := range rows {
		usageCost := cost(resourceType, row.Usage)
		requestCost := cost(resourceType, row.Request)
		data = append(data, model.ResourceUtilization{
			Resource:           resourceType,
			Timestamp:          row.Timestamp.Time.UTC(),
			Usage:              row.Usage,
			UsageCost:          usageCost,
			Request:            row.Request,
			RequestCost:        requestCost,
			RequestCostOverage: requestCost - usageCost,
		})
	}

	return data, nil
}

// normalizeTime will truncate a time.Time down to the hour, and return it as UTC
func normalizeTime(ts time.Time) time.Time {
	return ts.Truncate(time.Hour).UTC()
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
