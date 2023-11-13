package resourceusage

import (
	"context"
	"sort"
	"time"

	"github.com/nais/console-backend/internal/graph/scalar"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/sirupsen/logrus"
)

type Client interface {
	// UtilizationForApp returns resource utilization (usage and request) for the given app, in the given time range
	UtilizationForApp(ctx context.Context, resource model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error)

	// UtilizationForTeam returns resource utilization (usage and request) for a given team in the given time range
	UtilizationForTeam(ctx context.Context, resource model.ResourceType, env, team string, start, end time.Time) ([]model.ResourceUtilization, error)

	// ResourceUtilizationOverageCostForTeam will return app overage cost for a given team in the given time range
	ResourceUtilizationOverageCostForTeam(ctx context.Context, team string, start, end time.Time) (*model.ResourceUtilizationOverageCostForTeam, error)

	// ResourceUtilizationRangeForApp will return the min and max timestamps for a specific app
	ResourceUtilizationRangeForApp(ctx context.Context, env, team, app string) (*model.ResourceUtilizationDateRange, error)

	// ResourceUtilizationRangeForTeam will return the min and max timestamps for a specific team
	ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error)
}

type utilizationMap map[time.Time]*model.ResourceUtilization

type client struct {
	querier gensql.Querier
	log     logrus.FieldLogger
}

// NewClient creates a new resourceusage client
func NewClient(querier gensql.Querier, log logrus.FieldLogger) Client {
	return &client{
		querier: querier,
		log:     log,
	}
}

func (c *client) ResourceUtilizationRangeForApp(ctx context.Context, env, team, app string) (*model.ResourceUtilizationDateRange, error) {
	dates, err := c.querier.ResourceUtilizationRangeForApp(ctx, gensql.ResourceUtilizationRangeForAppParams{
		Env:  env,
		Team: team,
		App:  app,
	})
	if err != nil {
		return nil, err
	}

	var from *scalar.Date
	var to *scalar.Date
	if !dates.From.Time.IsZero() {
		f := scalar.NewDate(dates.From.Time)
		from = &f
	}
	if !dates.To.Time.IsZero() {
		t := scalar.NewDate(dates.To.Time)
		to = &t
	}

	return &model.ResourceUtilizationDateRange{
		From: from,
		To:   to,
	}, nil
}

func (c *client) ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error) {
	dates, err := c.querier.ResourceUtilizationRangeForTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	var from *scalar.Date
	var to *scalar.Date
	if !dates.From.Time.IsZero() {
		f := scalar.NewDate(dates.From.Time)
		from = &f
	}
	if !dates.To.Time.IsZero() {
		t := scalar.NewDate(dates.To.Time)
		to = &t
	}

	return &model.ResourceUtilizationDateRange{
		From: from,
		To:   to,
	}, nil
}

func (c *client) ResourceUtilizationOverageCostForTeam(ctx context.Context, team string, start, end time.Time) (*model.ResourceUtilizationOverageCostForTeam, error) {
	startTs := pgtype.Timestamptz{}
	err := startTs.Scan(normalizeTime(start))
	if err != nil {
		return nil, err
	}

	endTs := pgtype.Timestamptz{}
	err = endTs.Scan(normalizeTime(end.Add(time.Hour * 24)))
	if err != nil {
		return nil, err
	}

	rows, err := c.querier.ResourceUtilizationOverageCostForTeam(ctx, gensql.ResourceUtilizationOverageCostForTeamParams{
		Team:  team,
		Start: startTs,
		End:   endTs,
	})
	if err != nil {
		return nil, err
	}

	costMap := make(map[string]map[string]float64)
	for _, row := range rows {
		if _, exists := costMap[row.App]; !exists {
			costMap[row.App] = make(map[string]float64)
		}
		if _, exists := costMap[row.App][row.Env]; !exists {
			costMap[row.App][row.Env] = 0
		}

		costMap[row.App][row.Env] += cost(row.ResourceType, row.Request-row.Usage)
	}

	var sum float64
	apps := make([]model.AppWithResourceUtilizationOverageCost, 0)
	for app, envs := range costMap {
		for env, cost := range envs {
			sum += cost
			apps = append(apps, model.AppWithResourceUtilizationOverageCost{
				Team:    team,
				App:     app,
				Env:     env,
				Overage: cost,
			})
		}
	}
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Overage > apps[j].Overage
	})
	return &model.ResourceUtilizationOverageCostForTeam{
		Sum:  sum,
		Apps: apps,
	}, nil
}

func (c *client) UtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error) {
	startTs := pgtype.Timestamptz{}
	err := startTs.Scan(normalizeTime(start))
	if err != nil {
		return nil, err
	}

	endTs := pgtype.Timestamptz{}
	err = endTs.Scan(normalizeTime(end.Add(time.Hour * 24)))
	if err != nil {
		return nil, err
	}

	rows, err := c.querier.ResourceUtilizationForApp(ctx, gensql.ResourceUtilizationForAppParams{
		Env:          env,
		Team:         team,
		App:          app,
		ResourceType: resourceType.ToDatabaseEnum(),
		Start:        startTs,
		End:          endTs,
	})
	if err != nil {
		return nil, err
	}

	data := make([]model.ResourceUtilization, 0)
	for _, row := range rows {
		usageCost := cost(resourceType.ToDatabaseEnum(), row.Usage)
		requestCost := cost(resourceType.ToDatabaseEnum(), row.Request)
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
	err = endTs.Scan(normalizeTime(end.Add(time.Hour * 24)))
	if err != nil {
		return nil, err
	}

	rows, err := c.querier.ResourceUtilizationForTeam(ctx, gensql.ResourceUtilizationForTeamParams{
		Env:          env,
		Team:         team,
		ResourceType: resourceType.ToDatabaseEnum(),
		Start:        startTs,
		End:          endTs,
	})
	if err != nil {
		return nil, err
	}

	data := make([]model.ResourceUtilization, 0)
	for _, row := range rows {
		usageCost := cost(resourceType.ToDatabaseEnum(), row.Usage)
		requestCost := cost(resourceType.ToDatabaseEnum(), row.Request)
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
func cost(resourceType gensql.ResourceType, value float64) (cost float64) {
	if resourceType == gensql.ResourceTypeCpu {
		cost = 131.0 / 30.0 * value
	} else {
		cost = 18.0 / 1024 / 1024 / 1024 / 30.0 * value
	}

	return cost / 24.0
}
