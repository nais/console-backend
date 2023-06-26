package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.30

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
)

// Manifest is the resolver for the manifest field.
func (r *jobResolver) Manifest(ctx context.Context, obj *model.Job) (string, error) {
	job, err := r.K8s.JobManifest(ctx, obj.Name, obj.GQLVars.Team, obj.Env.Name)
	if err != nil {
		return "", fmt.Errorf("getting job manifest from Kubernetes: %w", err)
	}
	return job, err
}

// Job is the resolver for the job field.
func (r *queryResolver) Job(ctx context.Context, name string, team string, env string) (*model.Job, error) {
	job, err := r.K8s.Job(ctx, name, team, env)
	if err != nil {
		return nil, fmt.Errorf("getting job from Kubernetes: %w", err)
	}
	return job, nil
}

// Job returns JobResolver implementation.
func (r *Resolver) Job() JobResolver { return &jobResolver{r} }

type jobResolver struct{ *Resolver }
