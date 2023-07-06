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
		return "", fmt.Errorf("getting app manifest from Kubernetes: %w", err)
	}
	return app, err
}

// Team is the resolver for the team field.
func (r *appResolver) Team(ctx context.Context, obj *model.App) (*model.Team, error) {
	return r.TeamsClient.GetTeam(ctx, obj.GQLVars.Team)
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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *appResolver) State(ctx context.Context, obj *model.App) (model.AppState, error) {
	panic(fmt.Errorf("not implemented: State - state"))
}
func (r *appResolver) Message(ctx context.Context, obj *model.App) (string, error) {
	panic(fmt.Errorf("not implemented: Message - message"))
}
