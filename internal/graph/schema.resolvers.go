package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
)

// From is the resolver for the from field.
func (r *pageInfoResolver) From(ctx context.Context, obj *model.PageInfo) (int, error) {
	if obj.StartCursor == nil {
		return 0, nil
	}
	return obj.StartCursor.Offset + 1, nil
}

// To is the resolver for the to field.
func (r *pageInfoResolver) To(ctx context.Context, obj *model.PageInfo) (int, error) {
	if obj.EndCursor == nil {
		return 0, nil
	}
	return obj.EndCursor.Offset + 1, nil
}

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id model.Ident) (model.Node, error) {
	switch id.Type {
	case "user":
		u, err := r.TeamsClient.GetUserByID(ctx, id.ID)
		if err != nil {
			return nil, fmt.Errorf("getting user from Teams: %w", err)
		}
		return u, nil
	}
	return nil, fmt.Errorf("unknown type %q", id.Type)
}

// PageInfo returns PageInfoResolver implementation.
func (r *Resolver) PageInfo() PageInfoResolver { return &pageInfoResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type pageInfoResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
