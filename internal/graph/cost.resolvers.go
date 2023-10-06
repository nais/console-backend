package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
)

// DailyCostForApp is the resolver for the dailyCostForApp field.
func (r *queryResolver) DailyCostForApp(ctx context.Context, team string, app string, env string, from model.Date, to model.Date) (*model.DailyCost, error) {
	err := ValidateDateInterval(from, to)
	if err != nil {
		return nil, err
	}

	rows, err := r.Queries.DailyCostForApp(ctx, gensql.DailyCostForAppParams{
		App:      &app,
		Team:     &team,
		Env:      &env,
		FromDate: from.PgDate(),
		ToDate:   to.PgDate(),
	})
	if err != nil {
		return nil, fmt.Errorf("cost query: %w", err)
	}

	costs, sum := DailyCostsFromDatabaseRows(from, to, rows)
	series := make([]*model.CostSeries, 0)
	for costType, data := range costs {
		costTypeSum := 0.0
		for _, cost := range data {
			costTypeSum += cost.Cost
		}
		series = append(series, &model.CostSeries{
			CostType: costType,
			Sum:      costTypeSum,
			Data:     data,
		})
	}

	return &model.DailyCost{
		Sum:    sum,
		Series: series,
	}, nil
}

// DailyCostForTeam is the resolver for the dailyCostForTeam field.
func (r *queryResolver) DailyCostForTeam(ctx context.Context, team string, from model.Date, to model.Date) (*model.DailyCost, error) {
	err := ValidateDateInterval(from, to)
	if err != nil {
		return nil, err
	}

	rows, err := r.Queries.DailyCostForTeam(ctx, gensql.DailyCostForTeamParams{
		Team:     &team,
		FromDate: from.PgDate(),
		ToDate:   to.PgDate(),
	})
	if err != nil {
		return nil, fmt.Errorf("cost query: %w", err)
	}

	costs, sum := DailyCostsForTeamFromDatabaseRows(from, to, rows)
	series := make([]*model.CostSeries, 0)

	for costType, data := range costs {
		costTypeSum := 0.0
		for _, cost := range data {
			costTypeSum += cost.Cost
		}
		series = append(series, &model.CostSeries{
			CostType: costType,
			Sum:      costTypeSum,
			Data:     data,
		})
	}

	return &model.DailyCost{
		Sum:    sum,
		Series: series,
	}, nil
}

// MonthlyCost is the resolver for the monthlyCost field.
func (r *queryResolver) MonthlyCost(ctx context.Context, filter model.MonthlyCostFilter) (*model.MonthlyCost, error) {
	if filter.App != "" && filter.Env != "" && filter.Team != "" {
		rows, err := r.Queries.MonthlyCostForApp(ctx, gensql.MonthlyCostForAppParams{
			Team: &filter.Team,
			App:  &filter.App,
			Env:  &filter.Env,
		})
		if err != nil {
			return nil, err
		}
		sum := 0.0
		cost := make([]*model.CostEntry, len(rows))
		for idx, row := range rows {
			sum += float64(row.DailyCost)
			// make date variable equal last day in month of row.LastRecordedDate

			cost[idx] = &model.CostEntry{
				Date: model.NewDate(row.LastRecordedDate.Time),
				Cost: float64(row.DailyCost),
			}
		}
		return &model.MonthlyCost{
			Sum:  sum,
			Cost: cost,
		}, nil
	} else if filter.App == "" && filter.Env == "" && filter.Team != "" {
		rows, err := r.Queries.MonthlyCostForTeam(ctx, &filter.Team)
		if err != nil {
			return nil, err
		}
		sum := 0.0
		cost := make([]*model.CostEntry, len(rows))
		for idx, row := range rows {
			sum += float64(row.DailyCost)
			// make date variable equal last day in month of row.LastRecordedDate

			cost[idx] = &model.CostEntry{
				Date: model.NewDate(row.LastRecordedDate.Time),
				Cost: float64(row.DailyCost),
			}
		}
		return &model.MonthlyCost{
			Sum:  sum,
			Cost: cost,
		}, nil
	}
	return nil, fmt.Errorf("not implemented")
}

// EnvCost is the resolver for the envCost field.
func (r *queryResolver) EnvCost(ctx context.Context, filter model.EnvCostFilter) ([]*model.EnvCost, error) {
	err := ValidateDateInterval(filter.From, filter.To)
	if err != nil {
		return nil, err
	}

	ret := make([]*model.EnvCost, len(r.Clusters))
	for idx, cluster := range r.Clusters {
		appsCost := make([]*model.AppCost, 0)
		rows, err := r.Queries.DailyEnvCostForTeam(ctx, gensql.DailyEnvCostForTeamParams{
			Team:     &filter.Team,
			Env:      &cluster,
			FromDate: filter.From.PgDate(),
			ToDate:   filter.To.PgDate(),
		})
		if err != nil {
			return nil, fmt.Errorf("cost query: %w", err)
		}

		costs, sum := DailyCostsForTeamPerEnvFromDatabaseRows(filter.From, filter.To, rows)

		for app, appCosts := range costs {
			appSum := 0.0
			for _, c := range appCosts {
				appSum += c.Cost
			}
			appsCost = append(appsCost, &model.AppCost{
				App:  app,
				Sum:  appSum,
				Cost: appCosts,
			})
		}

		sort.Slice(appsCost, func(i, j int) bool {
			return appsCost[i].Sum < appsCost[j].Sum
		})

		ret[idx] = &model.EnvCost{
			Env:  cluster,
			Apps: appsCost,
			Sum:  sum,
		}
	}

	return ret, nil
}
