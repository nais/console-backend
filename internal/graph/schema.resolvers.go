package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
)

// Node is the resolver for the node field.
func (r *queryResolver) Node(ctx context.Context, id scalar.Ident) (model.Node, error) {
	switch id.Type {
	case scalar.IdentTypeTeam:
		t, err := r.teamsClient.GetTeam(ctx, id.ID)
		if err != nil {
			return nil, fmt.Errorf("getting team from Teams: %w", err)
		}
		return t, nil
	case scalar.IdentTypeUser:
		u, err := r.teamsClient.GetUserByID(ctx, id.ID)
		if err != nil {
			return nil, fmt.Errorf("getting user from Teams: %w", err)
		}
		return u, nil
	}
	return nil, fmt.Errorf("unsupported type %q in node query", id.Type)
}

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *pageInfoResolver) From(ctx context.Context, obj *model.PageInfo) (int, error) {
	if obj.StartCursor == nil {
		return 0, nil
	}
	return obj.StartCursor.Offset + 1, nil
}
func (r *pageInfoResolver) To(ctx context.Context, obj *model.PageInfo) (int, error) {
	if obj.EndCursor == nil {
		return 0, nil
	}
	return obj.EndCursor.Offset + 1, nil
}
func (r *Resolver) PageInfo() PageInfoResolver { return &pageInfoResolver{r} }

type pageInfoResolver struct{ *Resolver }
