package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"

	"github.com/nais/console-backend/internal/graph/model"
)

// Runs is the resolver for the runs field.
func (r *naisJobResolver) Runs(ctx context.Context, obj *model.NaisJob) ([]*model.Run, error) {
	return r.K8s.Runs(ctx, obj.GQLVars.Team, obj.Env.Name, obj.Name)
}

// Manifest is the resolver for the manifest field.
func (r *naisJobResolver) Manifest(ctx context.Context, obj *model.NaisJob) (string, error) {
	return r.K8s.NaisJobManifest(ctx, obj.Name, obj.GQLVars.Team, obj.Env.Name)
}

// Team is the resolver for the team field.
func (r *naisJobResolver) Team(ctx context.Context, obj *model.NaisJob) (*model.Team, error) {
	return r.TeamsClient.GetTeam(ctx, obj.GQLVars.Team)
}

// Naisjob is the resolver for the naisjob field.
func (r *queryResolver) Naisjob(ctx context.Context, name string, team string, env string) (*model.NaisJob, error) {
	return r.K8s.NaisJob(ctx, name, team, env)
}

// NaisJob returns NaisJobResolver implementation.
func (r *Resolver) NaisJob() NaisJobResolver { return &naisJobResolver{r} }

type naisJobResolver struct{ *Resolver }
