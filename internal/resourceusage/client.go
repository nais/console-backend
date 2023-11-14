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
	// ResourceUtilization returns resource utilization (usage and request) for the given app/job, in the given time range
	ResourceUtilization(ctx context.Context, env, team, name string, kind gensql.Kind, start, end time.Time) (*model.ResourceUtilization, error)

	// ResourceUtilizationForTeam returns resource utilization (usage and request) for a given team in the given time range
	ResourceUtilizationForTeam(ctx context.Context, team string, start, end time.Time) ([]model.ResourceUtilizationForEnv, error)

	// ResourceUtilizationOverageCostForTeam will return app overage cost for a given team in the given time range
	ResourceUtilizationOverageCostForTeam(ctx context.Context, team string, start, end time.Time) (*model.ResourceUtilizationOverageCostForTeam, error)

	// ResourceUtilizationDateRange will return the min and max timestamps for a specific app/job
	ResourceUtilizationDateRange(ctx context.Context, env, team, name string, kind gensql.Kind) (*model.ResourceUtilizationDateRange, error)

	// ResourceUtilizationDateRangeForTeam will return the min and max timestamps for a specific team
	ResourceUtilizationDateRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error)
}

type (
	utilizationMap map[time.Time]*model.ResourceUtilizationMetrics
	overageCostMap map[string]map[string]map[gensql.Kind]float64 // env -> name -> kind -> cost
)

type client struct {
	clusters []string
	querier  gensql.Querier
	log      logrus.FieldLogger
}

// NewClient creates a new resourceusage client
func NewClient(clusters []string, querier gensql.Querier, log logrus.FieldLogger) Client {
	return &client{
		clusters: clusters,
		querier:  querier,
		log:      log,
	}
}

func (c *client) ResourceUtilization(ctx context.Context, env, team, name string, kind gensql.Kind, start, end time.Time) (*model.ResourceUtilization, error) {
	cpu, err := c.resourceUtilization(ctx, model.ResourceTypeCPU, env, team, name, kind, start, end)
	if err != nil {
		return nil, err
	}

	memory, err := c.resourceUtilization(ctx, model.ResourceTypeMemory, env, team, name, kind, start, end)
	if err != nil {
		return nil, err
	}

	return &model.ResourceUtilization{
		CPU:    cpu,
		Memory: memory,
	}, nil
}

func (c *client) ResourceUtilizationForTeam(ctx context.Context, team string, start, end time.Time) ([]model.ResourceUtilizationForEnv, error) {
	ret := make([]model.ResourceUtilizationForEnv, 0)
	for _, env := range c.clusters {
		cpu, err := c.resourceUtilizationForTeam(ctx, model.ResourceTypeCPU, env, team, start, end)
		if err != nil {
			return nil, err
		}

		memory, err := c.resourceUtilizationForTeam(ctx, model.ResourceTypeMemory, env, team, start, end)
		if err != nil {
			return nil, err
		}

		ret = append(ret, model.ResourceUtilizationForEnv{
			Env:    env,
			CPU:    cpu,
			Memory: memory,
		})
	}
	return ret, nil
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

	costMap := getCostMapFromRows(rows)
	var sum float64
	ret := make([]model.OverageEntry, 0)
	for env, apps := range costMap {
		for app, kinds := range apps {
			for kind, cost := range kinds {
				sum += cost
				ret = append(ret, model.OverageEntry{
					Team:    team,
					Name:    app,
					Env:     env,
					Overage: cost,
					IsApp:   kind == gensql.KindApp,
				})
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Overage > ret[j].Overage
	})
	return &model.ResourceUtilizationOverageCostForTeam{
		Sum:     sum,
		Entries: ret,
	}, nil
}

func (c *client) ResourceUtilizationDateRange(ctx context.Context, env, team, name string, kind gensql.Kind) (*model.ResourceUtilizationDateRange, error) {
	dates, err := c.querier.ResourceUtilizationDateRange(ctx, gensql.ResourceUtilizationDateRangeParams{
		Env:  env,
		Team: team,
		Name: name,
		Kind: kind,
	})
	if err != nil {
		return nil, err
	}
	return getDateRange(dates.From, dates.To), nil
}

func (c *client) ResourceUtilizationDateRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error) {
	dates, err := c.querier.ResourceUtilizationDateRangeForTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	return getDateRange(dates.From, dates.To), nil
}

func (c *client) resourceUtilization(ctx context.Context, resourceType model.ResourceType, env, team, name string, kind gensql.Kind, start, end time.Time) ([]model.ResourceUtilizationMetrics, error) {
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
		Kind:         kind,
		Env:          env,
		Team:         team,
		Name:         name,
		ResourceType: resourceType.ToDatabaseEnum(),
		Start:        startTs,
		End:          endTs,
	})
	if err != nil {
		return nil, err
	}

	data := make([]model.ResourceUtilizationMetrics, 0)
	for _, row := range rows {
		usageCost := cost(resourceType.ToDatabaseEnum(), row.Usage)
		requestCost := cost(resourceType.ToDatabaseEnum(), row.Request)
		data = append(data, model.ResourceUtilizationMetrics{
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

func (c *client) resourceUtilizationForTeam(ctx context.Context, resourceType model.ResourceType, env, team string, start, end time.Time) ([]model.ResourceUtilizationMetrics, error) {
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

	data := make([]model.ResourceUtilizationMetrics, 0)
	for _, row := range rows {
		usageCost := cost(resourceType.ToDatabaseEnum(), row.Usage)
		requestCost := cost(resourceType.ToDatabaseEnum(), row.Request)
		data = append(data, model.ResourceUtilizationMetrics{
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

// getCostMapFromRows converts a slice of ResourceUtilizationOverageCostForTeamRow to a overCostMap
func getCostMapFromRows(rows []*gensql.ResourceUtilizationOverageCostForTeamRow) overageCostMap {
	costMap := make(overageCostMap)
	for _, row := range rows {
		if _, exists := costMap[row.Env]; !exists {
			costMap[row.Env] = make(map[string]map[gensql.Kind]float64)
		}
		if _, exists := costMap[row.Env][row.Name]; !exists {
			costMap[row.Env][row.Name] = make(map[gensql.Kind]float64)
		}
		if _, exists := costMap[row.Env][row.Name][row.Kind]; !exists {
			costMap[row.Env][row.Name][row.Kind] = 0
		}

		costMap[row.Env][row.Name][row.Kind] += cost(row.ResourceType, row.Request-row.Usage)
	}
	return costMap
}

// getDateRange returns a date range model from two timestamps
func getDateRange(from, to pgtype.Timestamptz) *model.ResourceUtilizationDateRange {
	var fromDate, toDate *scalar.Date

	if !from.Time.IsZero() {
		f := scalar.NewDate(from.Time)
		fromDate = &f
	}
	if !to.Time.IsZero() {
		t := scalar.NewDate(to.Time)
		toDate = &t
	}

	return &model.ResourceUtilizationDateRange{
		From: fromDate,
		To:   toDate,
	}
}
