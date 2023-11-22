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
	// ResourceUtilizationForApp returns resource utilization (usage and request) for the given app, in the given time range
	ResourceUtilizationForApp(ctx context.Context, env, team, app string, start, end time.Time) (*model.ResourceUtilizationForApp, error)

	// ResourceUtilizationForTeam returns resource utilization (usage and request) for a given team in the given time range
	ResourceUtilizationForTeam(ctx context.Context, team string, start, end time.Time) ([]model.ResourceUtilizationForEnv, error)

	// ResourceUtilizationOverageCostForTeam will return app overage cost for a given team in the given time range
	ResourceUtilizationOverageCostForTeam(ctx context.Context, team string, start, end time.Time) (*model.ResourceUtilizationOverageCostForTeam, error)

	// ResourceUtilizationRangeForApp will return the min and max timestamps for a specific app
	ResourceUtilizationRangeForApp(ctx context.Context, env, team, app string) (*model.ResourceUtilizationDateRange, error)

	// ResourceUtilizationRangeForTeam will return the min and max timestamps for a specific team
	ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error)

	// CurrentResourceUtilizationForApp will return the current percentages of resource utilization for an app
	CurrentResourceUtilizationForApp(ctx context.Context, env string, team string, app string) (*model.CurrentResourceUtilizationForApp, error)
}

type (
	utilizationMap map[time.Time]*model.ResourceUtilization
	overageCostMap map[string]map[string]float64 // env -> app -> cost
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

func (c *client) ResourceUtilizationForApp(ctx context.Context, env, team, app string, start, end time.Time) (*model.ResourceUtilizationForApp, error) {
	cpu, err := c.resourceUtilizationForApp(ctx, model.ResourceTypeCPU, env, team, app, start, end)
	if err != nil {
		return nil, err
	}

	memory, err := c.resourceUtilizationForApp(ctx, model.ResourceTypeMemory, env, team, app, start, end)
	if err != nil {
		return nil, err
	}

	return &model.ResourceUtilizationForApp{
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
	ret := make([]model.AppWithResourceUtilizationOverageCost, 0)
	for env, apps := range costMap {
		for app, cost := range apps {
			sum += cost
			ret = append(ret, model.AppWithResourceUtilizationOverageCost{
				Team:    team,
				App:     app,
				Env:     env,
				Overage: cost,
			})
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Overage > ret[j].Overage
	})
	return &model.ResourceUtilizationOverageCostForTeam{
		Sum:  sum,
		Apps: ret,
	}, nil
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
	return getDateRange(dates.From, dates.To), nil
}

func (c *client) ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error) {
	dates, err := c.querier.ResourceUtilizationRangeForTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	return getDateRange(dates.From, dates.To), nil
}

func (c *client) resourceUtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end time.Time) ([]model.ResourceUtilization, error) {
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
		usageCost := costPerHour(resourceType.ToDatabaseEnum(), row.Usage)
		requestCost := costPerHour(resourceType.ToDatabaseEnum(), row.Request)
		usagePercentage := float64(0)
		if row.Request > 0 {
			usagePercentage = row.Usage / row.Request * 100
		}
		data = append(data, model.ResourceUtilization{
			Resource:           resourceType,
			Timestamp:          row.Timestamp.Time.UTC(),
			Usage:              row.Usage,
			UsageCost:          usageCost,
			Request:            row.Request,
			RequestCost:        requestCost,
			RequestCostOverage: requestCost - usageCost,
			UsagePercentage:    usagePercentage,
		})
	}

	return data, nil
}

func (c *client) resourceUtilizationForTeam(ctx context.Context, resourceType model.ResourceType, env, team string, start, end time.Time) ([]model.ResourceUtilization, error) {
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
		usageCost := costPerHour(resourceType.ToDatabaseEnum(), row.Usage)
		requestCost := costPerHour(resourceType.ToDatabaseEnum(), row.Request)
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

func (c *client) CurrentResourceUtilizationForApp(ctx context.Context, env string, team string, app string) (*model.CurrentResourceUtilizationForApp, error) {
	cpu, err := c.querier.CurrentResourceUtilizationForApp(ctx, gensql.CurrentResourceUtilizationForAppParams{
		Env:          env,
		Team:         team,
		App:          app,
		ResourceType: gensql.ResourceTypeCpu,
	})
	if err != nil {
		return nil, err
	}

	memory, err := c.querier.CurrentResourceUtilizationForApp(ctx, gensql.CurrentResourceUtilizationForAppParams{
		Env:          env,
		Team:         team,
		App:          app,
		ResourceType: gensql.ResourceTypeMemory,
	})
	if err != nil {
		return nil, err
	}

	var cpuUtilization, memoryUtilization float64
	if cpu.Request > 0 {
		cpuUtilization = cpu.Usage / cpu.Request * 100
	}
	if memory.Request > 0 {
		memoryUtilization = memory.Usage / memory.Request * 100
	}

	return &model.CurrentResourceUtilizationForApp{
		CPU:    cpuUtilization,
		Memory: memoryUtilization,
	}, nil
}

// normalizeTime will truncate a time.Time down to the hour, and return it as UTC
func normalizeTime(ts time.Time) time.Time {
	return ts.Truncate(time.Hour).UTC()
}

// costPerHour calculates the cost for the given resource type
func costPerHour(resourceType gensql.ResourceType, value float64) (cost float64) {
	const costPerCpuCorePerMonthInNok = 131
	const costPerGBMemoryPerMonthInNok = 18
	const eurToNokExchangeRate = 11.5

	if resourceType == gensql.ResourceTypeCpu {
		cost = costPerCpuCorePerMonthInNok * value
	} else {
		// for memory the value is in bytes
		cost = (costPerGBMemoryPerMonthInNok / 1024 / 1024 / 1024) * value
	}

	return cost / 30.0 / 24.0 / eurToNokExchangeRate
}

// getCostMapFromRows converts a slice of ResourceUtilizationOverageCostForTeamRow to a overCostMap
func getCostMapFromRows(rows []*gensql.ResourceUtilizationOverageCostForTeamRow) overageCostMap {
	costMap := make(overageCostMap)
	for _, row := range rows {
		if _, exists := costMap[row.Env]; !exists {
			costMap[row.Env] = make(map[string]float64)
		}
		if _, exists := costMap[row.Env][row.App]; !exists {
			costMap[row.Env][row.App] = 0
		}

		costMap[row.Env][row.App] += costPerHour(row.ResourceType, row.Request-row.Usage)
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
