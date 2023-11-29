package resourceusage

import (
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/sirupsen/logrus"
)

type Client interface {
	// ResourceUtilizationForApp returns resource utilization (usage and request) for the given app, in the given time range
	ResourceUtilizationForApp(ctx context.Context, env, team, app string, start, end scalar.Date) (*model.ResourceUtilizationForApp, error)

	// ResourceUtilizationForTeam returns resource utilization (usage and request) for a given team in the given time range
	ResourceUtilizationForTeam(ctx context.Context, team string, start, end scalar.Date) ([]model.ResourceUtilizationForEnv, error)

	// ResourceUtilizationOverageForTeam will return latest overage data for a given team
	ResourceUtilizationOverageForTeam(ctx context.Context, team string) (*model.ResourceUtilizationOverageForTeam, error)

	// ResourceUtilizationRangeForApp will return the min and max timestamps for a specific app
	ResourceUtilizationRangeForApp(ctx context.Context, env, team, app string) (*model.ResourceUtilizationDateRange, error)

	// ResourceUtilizationRangeForTeam will return the min and max timestamps for a specific team
	ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*model.ResourceUtilizationDateRange, error)

	// CurrentResourceUtilizationForApp will return the current percentages of resource utilization for an app
	CurrentResourceUtilizationForApp(ctx context.Context, env, team, app string) (*model.CurrentResourceUtilization, error)

	// CurrentResourceUtilizationForTeam will return the current percentages of resource utilization for a team across all apps and environments
	CurrentResourceUtilizationForTeam(ctx context.Context, team string) (*model.CurrentResourceUtilization, error)
}

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

func (c *client) ResourceUtilizationForApp(ctx context.Context, env, team, app string, start, end scalar.Date) (*model.ResourceUtilizationForApp, error) {
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

func (c *client) ResourceUtilizationForTeam(ctx context.Context, team string, start, end scalar.Date) ([]model.ResourceUtilizationForEnv, error) {
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

func (c *client) ResourceUtilizationOverageForTeam(ctx context.Context, team string) (*model.ResourceUtilizationOverageForTeam, error) {
	cpu, cpuCost, timestamp, err := c.resourceUtilizationOverageForTeam(ctx, gensql.ResourceTypeCpu, team)
	if err != nil {
		return nil, err
	}

	memory, memoryCost, timestamp, err := c.resourceUtilizationOverageForTeam(ctx, gensql.ResourceTypeMemory, team)
	if err != nil {
		return nil, err
	}

	return &model.ResourceUtilizationOverageForTeam{
		OverageCost: cpuCost + memoryCost,
		CPU:         cpu,
		Memory:      memory,
		Timestamp:   &timestamp.Time,
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

func (c *client) CurrentResourceUtilizationForApp(ctx context.Context, env string, team string, app string) (*model.CurrentResourceUtilization, error) {
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

	return &model.CurrentResourceUtilization{
		CPU:    resourceUtilization(model.ResourceTypeCPU, cpu.Timestamp.Time.UTC(), cpu.Request, cpu.Usage),
		Memory: resourceUtilization(model.ResourceTypeMemory, memory.Timestamp.Time.UTC(), memory.Request, memory.Usage),
	}, nil
}

func (c *client) CurrentResourceUtilizationForTeam(ctx context.Context, team string) (*model.CurrentResourceUtilization, error) {
	currentCpu, err := c.querier.CurrentResourceUtilizationForTeam(ctx, gensql.CurrentResourceUtilizationForTeamParams{
		Team:         team,
		ResourceType: gensql.ResourceTypeCpu,
	})
	if err != nil {
		return nil, err
	}

	currentMemory, err := c.querier.CurrentResourceUtilizationForTeam(ctx, gensql.CurrentResourceUtilizationForTeamParams{
		Team:         team,
		ResourceType: gensql.ResourceTypeMemory,
	})
	if err != nil {
		return nil, err
	}

	return &model.CurrentResourceUtilization{
		CPU:    resourceUtilization(model.ResourceTypeCPU, currentCpu.Timestamp.Time.UTC(), currentCpu.Request, currentCpu.Usage),
		Memory: resourceUtilization(model.ResourceTypeMemory, currentMemory.Timestamp.Time.UTC(), currentMemory.Request, currentMemory.Usage),
	}, nil
}

func (c *client) resourceUtilizationForApp(ctx context.Context, resourceType model.ResourceType, env, team, app string, start, end scalar.Date) ([]model.ResourceUtilization, error) {
	s, err := start.Time()
	if err != nil {
		return nil, err
	}

	e, err := end.Time()
	if err != nil {
		return nil, err
	}
	e = e.AddDate(0, 0, 1)

	startTs := pgtype.Timestamptz{}
	err = startTs.Scan(s)
	if err != nil {
		return nil, err
	}

	endTs := pgtype.Timestamptz{}
	err = endTs.Scan(e)
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

	utilizationMap := initUtilizationMap(s, e)
	for _, row := range rows {
		ts := row.Timestamp.Time.UTC()
		utilizationMap[ts] = resourceUtilization(resourceType, ts, row.Request, row.Usage)
	}

	data := make([]model.ResourceUtilization, 0)
	for _, entry := range utilizationMap {
		data = append(data, entry)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Timestamp.Before(data[j].Timestamp)
	})

	return data, nil
}

func (c *client) resourceUtilizationForTeam(ctx context.Context, resourceType model.ResourceType, env, team string, start, end scalar.Date) ([]model.ResourceUtilization, error) {
	s, err := start.Time()
	if err != nil {
		return nil, err
	}

	e, err := end.Time()
	if err != nil {
		return nil, err
	}
	e = e.AddDate(0, 0, 1)

	startTs := pgtype.Timestamptz{}
	err = startTs.Scan(s)
	if err != nil {
		return nil, err
	}

	endTs := pgtype.Timestamptz{}
	err = endTs.Scan(e)
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

	utilizationMap := initUtilizationMap(s, e)
	for _, row := range rows {
		ts := row.Timestamp.Time.UTC()
		utilizationMap[ts] = resourceUtilization(resourceType, ts, row.Request, row.Usage)
	}

	data := make([]model.ResourceUtilization, 0)
	for _, entry := range utilizationMap {
		data = append(data, entry)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Timestamp.Before(data[j].Timestamp)
	})

	return data, nil
}

func (c *client) resourceUtilizationOverageForTeam(ctx context.Context, resource gensql.ResourceType, team string) (models []model.AppWithResourceUtilizationOverage, sumOverageCost float64, timestamp pgtype.Timestamptz, err error) {
	rows, err := c.querier.ResourceUtilizationOverageForTeam(ctx, gensql.ResourceUtilizationOverageForTeamParams{
		Team:         team,
		ResourceType: resource,
	})
	if err != nil {
		return
	}
	if len(rows) > 0 {
		timestamp = rows[0].Timestamp
	}
	for _, row := range rows {
		overageCost := costPerHour(resource, row.Overage)
		sumOverageCost += overageCost
		models = append(models, model.AppWithResourceUtilizationOverage{
			Overage: row.Overage,
			Env:     row.Env,
			Team:    team,
			App:     row.App,
		})
	}
	sort.Slice(models, func(i, j int) bool {
		return models[i].Overage > models[j].Overage
	})

	return
}

// costPerHour calculates the cost for the given resource type
func costPerHour(resourceType gensql.ResourceType, value float64) (cost float64) {
	const costPerCpuCorePerMonthInNok = 131.0
	const costPerGBMemoryPerMonthInNok = 18.0
	const eurToNokExchangeRate = 11.5

	if resourceType == gensql.ResourceTypeCpu {
		cost = costPerCpuCorePerMonthInNok * value
	} else {
		// for memory the value is in bytes
		cost = (costPerGBMemoryPerMonthInNok / 1024 / 1024 / 1024) * value
	}

	return cost / 30.0 / 24.0 / eurToNokExchangeRate
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

// resourceUtilization will return a resource utilization model
func resourceUtilization(resource model.ResourceType, ts time.Time, request, usage float64) model.ResourceUtilization {
	var utilization float64
	if request > 0 {
		utilization = usage / request * 100
	}

	requestCost := costPerHour(resource.ToDatabaseEnum(), request)
	usageCost := costPerHour(resource.ToDatabaseEnum(), usage)
	overageCostPerHour := requestCost - usageCost

	return model.ResourceUtilization{
		Timestamp:                  ts,
		Request:                    request,
		RequestCost:                requestCost,
		Usage:                      usage,
		UsageCost:                  usageCost,
		RequestCostOverage:         overageCostPerHour,
		Utilization:                utilization,
		EstimatedAnnualOverageCost: overageCostPerHour * 24 * 365,
	}
}

// initUtilizationMap will create a map of timestamps with empty resource utilization data. This is used to not have
// gaps in the graph. The last entry in the map will not be in the future.
func initUtilizationMap(start, end time.Time) map[time.Time]model.ResourceUtilization {
	now := time.Now().UTC()
	timestamps := make([]time.Time, 0)
	ts := start
	for ; ts.Before(end) && ts.Before(now); ts = ts.Add(rangedQueryStep) {
		timestamps = append(timestamps, ts)
	}

	utilization := make(map[time.Time]model.ResourceUtilization)
	for _, ts := range timestamps {
		utilization[ts] = model.ResourceUtilization{
			Timestamp: ts,
		}
	}
	return utilization
}
