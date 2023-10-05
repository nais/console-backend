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

// Cost is the resolver for the cost field.
func (r *queryResolver) Cost(ctx context.Context, filter model.CostFilter) (*model.Cost, error) {
	err := ValidateDateInterval(filter.From, filter.To)
	if err != nil {
		return nil, err
	}

	if filter.App != "" && filter.Env != "" && filter.Team != "" {
		rows, err := r.Queries.DailyCostForApp(ctx, gensql.DailyCostForAppParams{
			App:      &filter.App,
			Team:     &filter.Team,
			Env:      &filter.Env,
			FromDate: filter.From.PgDate(),
			ToDate:   filter.To.PgDate(),
		})
		if err != nil {
			return nil, fmt.Errorf("cost query: %w", err)
		}

		costs, sum := DailyCostsFromDatabaseRows(filter.From, filter.To, rows)
		series := make([]*model.CostSeries, 0)
		for costType, data := range costs {
			costTypeSum := 0.0
			for _, cost := range data {
				costTypeSum += cost.Cost
			}
			series = append(series, &model.CostSeries{
				CostType: costType,
				Data:     data,
				App:      filter.App,
				Team:     filter.Team,
				Env:      filter.Env,
				Sum:      costTypeSum,
			})
		}

		return &model.Cost{
			From:   filter.From,
			To:     filter.To,
			Series: series,
			Sum:    sum,
		}, nil
	} else if filter.App == "" && filter.Env == "" && filter.Team != "" {
		rows, err := r.Queries.DailyCostForTeam(ctx, gensql.DailyCostForTeamParams{
			Team:     &filter.Team,
			FromDate: filter.From.PgDate(),
			ToDate:   filter.To.PgDate(),
		})
		if err != nil {
			return nil, fmt.Errorf("cost query: %w", err)
		}

		costs, sum := DailyCostsForTeamFromDatabaseRows(filter.From, filter.To, rows)
		series := make([]*model.CostSeries, 0)

		for costType, data := range costs {
			costTypeSum := 0.0
			for _, cost := range data {
				costTypeSum += cost.Cost
			}
			series = append(series, &model.CostSeries{
				CostType: costType,
				Data:     data,
				App:      filter.App,
				Team:     filter.Team,
				Env:      filter.Env,
				Sum:      costTypeSum,
			})
		}

		return &model.Cost{
			From:   filter.From,
			To:     filter.To,
			Series: series,
			Sum:    sum,
		}, nil
	}

	return nil, fmt.Errorf("not implemented")
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

		// sort appsCost by sum by using sort.Slice
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
