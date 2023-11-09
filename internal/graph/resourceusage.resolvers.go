package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"time"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
)

// ResourceUtilizationForTeam is the resolver for the resourceUtilizationForTeam field.
func (r *queryResolver) ResourceUtilizationForTeam(ctx context.Context, resource model.ResourceType, team string, from *scalar.Date, to *scalar.Date) ([]model.ResourceUtilizationInEnv, error) {
	end := time.Now()
	start := end.Add(-24 * time.Hour * 6)

	var err error
	if to != nil {
		end, err = to.Time()
		if err != nil {
			return nil, err
		}
	}

	if from != nil {
		start, err = from.Time()
		if err != nil {
			return nil, err
		}
	}

	ret := make([]model.ResourceUtilizationInEnv, 0)
	for _, env := range r.clusters {
		data, err := r.resourceUsageClient.UtilizationForTeam(ctx, resource, env, team, start, end)
		if err != nil {
			return nil, err
		}
		ret = append(ret, model.ResourceUtilizationInEnv{
			Env:  env,
			Data: data,
		})
	}
	return ret, nil
}

// ResourceUtilizationForApp is the resolver for the resourceUtilizationForApp field.
func (r *queryResolver) ResourceUtilizationForApp(ctx context.Context, resource model.ResourceType, env string, team string, app string, from *scalar.Date, to *scalar.Date) ([]model.ResourceUtilization, error) {
	end := time.Now()
	start := end.Add(-24 * time.Hour * 6)

	var err error
	if to != nil {
		end, err = to.Time()
		if err != nil {
			return nil, err
		}
	}

	if from != nil {
		start, err = from.Time()
		if err != nil {
			return nil, err
		}
	}

	return r.resourceUsageClient.UtilizationForApp(ctx, resource, env, team, app, start, end)
}
