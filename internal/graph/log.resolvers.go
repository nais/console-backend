package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
)

// Log is the resolver for the log field.
func (r *subscriptionResolver) Log(ctx context.Context, input *model.LogSubscriptionInput) (<-chan *model.LogLine, error) {
	container := ""
	selector := ""
	switch {
	case input.App != nil:
		selector = "app=" + *input.App
		container = *input.App
	case input.Job != nil:
		selector = "app=" + *input.Job
		container = *input.Job
	default:
		return nil, fmt.Errorf("must specify either app or job")
	}

	return r.K8s.LogStream(ctx, input.Env, input.Team, selector, container, input.Instances)
}

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type subscriptionResolver struct{ *Resolver }
