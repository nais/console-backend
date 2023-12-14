package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/dependencytrack"
	"github.com/nais/console-backend/internal/graph/apierror"
	"github.com/nais/console-backend/internal/graph/model"
)

// Instances is the resolver for the instances field.
func (r *appResolver) Instances(ctx context.Context, obj *model.App) ([]*model.Instance, error) {
	instances, err := r.k8sClient.Instances(ctx, obj.GQLVars.Team, obj.Env.Name, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting instances from Kubernetes: %w", err)
	}

	return instances, nil
}

// Manifest is the resolver for the manifest field.
func (r *appResolver) Manifest(ctx context.Context, obj *model.App) (string, error) {
	app, err := r.k8sClient.Manifest(ctx, obj.Name, obj.GQLVars.Team, obj.Env.Name)
	if err != nil {
		return "", fmt.Errorf("getting app manifest from Kubernetes: %w", err)
	}
	return app, err
}

// Team is the resolver for the team field.
func (r *appResolver) Team(ctx context.Context, obj *model.App) (*model.Team, error) {
	team, err := r.teamsClient.GetTeam(ctx, obj.GQLVars.Team)
	if err != nil {
		return nil, apierror.ErrAppTeamNotFound
	}
	return team, nil
}

// Vulnerabilities is the resolver for the vulnerabilities field.
func (r *appResolver) Vulnerabilities(ctx context.Context, obj *model.App) (*model.Vulnerability, error) {
	return r.dependencyTrackClient.VulnerabilitySummary(ctx, &dependencytrack.AppInstance{Env: obj.Env.Name, Team: obj.GQLVars.Team, App: obj.Name, Image: obj.Image})
}

// App is the resolver for the app field.
func (r *queryResolver) App(ctx context.Context, name string, team string, env string) (*model.App, error) {
	app, err := r.k8sClient.App(ctx, name, team, env)
	if err != nil {
		return nil, apierror.ErrAppNotFound
	}
	return app, nil
}

// App returns AppResolver implementation.
func (r *Resolver) App() AppResolver { return &appResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type appResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
