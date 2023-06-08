package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.30

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
)

// Instances is the resolver for the instances field.
func (r *appResolver) Instances(ctx context.Context, obj *model.App) ([]*model.Instance, error) {
	instances, err := r.K8s.Instances(ctx, obj.GQLVars.Team, obj.Env.Name, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting instances from Kubernetes: %w", err)
	}

	return instances, nil
}

// Manifest is the resolver for the manifest field.
func (r *appResolver) Manifest(ctx context.Context, obj *model.App) (string, error) {
	app, err := r.K8s.Manifest(ctx, obj.Name, obj.GQLVars.Team, obj.Env.Name)
	if err != nil {
		return "", fmt.Errorf("getting app from Kubernetes: %w", err)
	}
	return app, err
}

// Team is the resolver for the team field.
func (r *appResolver) Team(ctx context.Context, obj *model.App) (*model.Team, error) {
	return r.TeamsClient.GetTeam(ctx, obj.GQLVars.Team)
}

// History is the resolver for the history field.
func (r *deployInfoResolver) History(ctx context.Context, obj *model.DeployInfo, first *int, last *int, after *model.Cursor, before *model.Cursor) (model.DeploymentResponse, error) {
	deploys, err := r.Hookd.DeploymentsByApp(ctx, obj.GQLVars.App, obj.GQLVars.Team, obj.GQLVars.Env)
	if err != nil {
		return nil, fmt.Errorf("getting deploys from Hookd: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	e := deployEdges(deploys, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	hasNext := len(deploys) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.DeploymentConnection{
		Edges: e,
		PageInfo: &model.PageInfo{
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}, nil
}

// App is the resolver for the app field.
func (r *queryResolver) App(ctx context.Context, name string, team string, env string) (*model.App, error) {
	app, err := r.K8s.App(ctx, name, team, env)
	if err != nil {
		return nil, fmt.Errorf("getting app from Kubernetes: %w", err)
	}
	return app, nil
}

// App returns AppResolver implementation.
func (r *Resolver) App() AppResolver { return &appResolver{r} }

// DeployInfo returns DeployInfoResolver implementation.
func (r *Resolver) DeployInfo() DeployInfoResolver { return &deployInfoResolver{r} }

type appResolver struct{ *Resolver }
type deployInfoResolver struct{ *Resolver }
