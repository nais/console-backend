package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
)

// Cost is the resolver for the cost field.
func (r *queryResolver) Cost(ctx context.Context, filter *model.CostFilter) (*model.Cost, error) {
	if filter == nil {
		return nil, fmt.Errorf("cost filter is nil")
	}

	if filter.StartDate == nil {
		start := time.Now().Add(time.Duration(-7 * time.Hour * 24))
		filter.StartDate = &start
	}

	if filter.EndDate == nil {
		end := time.Now()
		filter.EndDate = &end
	}

	if filter.App != "" && filter.Env != "" && filter.Team != "" && filter.StartDate != nil && filter.EndDate != nil {
		params := gensql.CostForAppParams{}
		params.App = &filter.App
		params.Team = &filter.Team
		params.Env = &filter.Env
		params.FromDate.Time = filter.StartDate.UTC()
		params.FromDate.Valid = true
		params.ToDate.Time = filter.EndDate.UTC()
		params.ToDate.Valid = true
		rows, err := r.Queries.CostForApp(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("cost query: %w", err)
		}

		mapTypeDailyCost := make(map[string][]*model.DailyCost)

		for _, row := range rows {
			mapTypeDailyCost[row.CostType] = append(mapTypeDailyCost[row.CostType], &model.DailyCost{
				Date: row.Date.Time,
				Cost: float64(row.Cost),
			})
		}

		cost := &model.Cost{
			From: *filter.StartDate,
			To:   *filter.EndDate,
		}

		for costType, dailyCost := range mapTypeDailyCost {
			cost.Series = append(cost.Series, &model.CostSeries{
				CostType: costType,
				Data:     dailyCost,
				App:      filter.App,
				Team:     filter.Team,
				Env:      filter.Env,
			})
		}

		return cost, nil
	} /*else if filter.Env != "" && filter.Team != "" && filter.StartDate != nil && filter.EndDate != nil {
		params := gensql.CostForAppParams{}
		params.Team = &filter.Team
		params.Env = &filter.Env
		params.FromDate.Time = filter.StartDate.UTC()
		params.FromDate.Valid = true
		params.ToDate.Time = filter.EndDate.UTC()
		params.ToDate.Valid = true
		rows, err := r.Queries.CostForApp(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("cost query: %w", err)
		}

		cost := &model.Cost{}
		cost.Series = make([]*model.CostSeries, 0)

		for _, row := range rows {
			cost.Series = append(cost.Series, &model.CostSeries{
				CostType: row.CostType,
				Data:     []float64{float64(row.Cost)},
				App:      *row.App,
				Team:     *row.Team,
				Env:      *row.Env,
			})

		}
	}
	*/
	return nil, fmt.Errorf("not implemented")
}
