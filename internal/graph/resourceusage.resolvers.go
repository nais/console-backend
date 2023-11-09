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
func (r *queryResolver) ResourceUtilizationForTeam(ctx context.Context, team string, from *scalar.Date, to *scalar.Date) ([]model.ResourceUtilizationInEnv, error) {
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
		ret = append(ret, model.ResourceUtilizationInEnv{
			Env: env,
			GQLVars: model.ResourceUtilizationInEnvGQLVars{
				Start: start,
				End:   end,
				Team:  team,
			},
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

// CPU is the resolver for the cpu field.
func (r *resourceUtilizationInEnvResolver) CPU(ctx context.Context, obj *model.ResourceUtilizationInEnv) ([]model.ResourceUtilization, error) {
	return r.resourceUsageClient.UtilizationForTeam(ctx, model.ResourceTypeCPU, obj.Env, obj.GQLVars.Team, obj.GQLVars.Start, obj.GQLVars.End)
}

// Memory is the resolver for the memory field.
func (r *resourceUtilizationInEnvResolver) Memory(ctx context.Context, obj *model.ResourceUtilizationInEnv) ([]model.ResourceUtilization, error) {
	return r.resourceUsageClient.UtilizationForTeam(ctx, model.ResourceTypeMemory, obj.Env, obj.GQLVars.Team, obj.GQLVars.Start, obj.GQLVars.End)
}

// ResourceUtilizationInEnv returns ResourceUtilizationInEnvResolver implementation.
func (r *Resolver) ResourceUtilizationInEnv() ResourceUtilizationInEnvResolver {
	return &resourceUtilizationInEnvResolver{r}
}

type resourceUtilizationInEnvResolver struct{ *Resolver }
