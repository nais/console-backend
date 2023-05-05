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

// Deploys is the resolver for the deploys field.
func (r *appResolver) Deploys(ctx context.Context, obj *model.App, first *int, after *model.Cursor) (*model.DeploymentConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &model.Cursor{Offset: 0}
	}
	deps, err := r.Hookd.GetDeploysForApp(ctx, obj.Name, obj.GQLVars.Team, obj.Env.Name)
	if err != nil {
		return nil, fmt.Errorf("getting deploys from Hookd: %w", err)
	}

	if *first > len(deps) {
		*first = len(deps)
	}

	e := deploymentEdges(deps, *first, after.Offset)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	return &model.DeploymentConnection{
		Edges: e,
		PageInfo: &model.PageInfo{
			StartCursor: startCursor,
			EndCursor:   endCursor,
			HasNextPage: len(deps) > *first+after.Offset,
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

type appResolver struct{ *Resolver }
