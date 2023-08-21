package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"

	"github.com/nais/console-backend/internal/graph/model"
)

// Log is the resolver for the log field.
func (r *instanceResolver) Log(ctx context.Context, obj *model.Instance, tailLines int) ([]*model.LogLine, error) {
	return r.K8s.Log(ctx, obj.GQLVars.Env, obj.GQLVars.Team, obj.Name, obj.GQLVars.AppName, int64(tailLines))
}

// Instance returns InstanceResolver implementation.
func (r *Resolver) Instance() InstanceResolver { return &instanceResolver{r} }

type instanceResolver struct{ *Resolver }
