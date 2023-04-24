package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.29

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/model"
)

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	email, err := auth.GetEmail(ctx)
	if err != nil {
		return nil, err
	}
	user, err := r.Console.GetUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("getting user from Console: %w", err)
	}
	return &model.User{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: email,
	}, nil
}

// Teams is the resolver for the teams field.
func (r *userResolver) Teams(ctx context.Context, obj *model.User, first *int, after *model.Cursor) (*model.TeamConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &model.Cursor{Offset: 0}
	}

	teams, err := r.Console.GetTeamsForUser(ctx, obj.Email)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Console: %w", err)
	}
	if *first > len(teams) {
		*first = len(teams)
	}

	e := edges(teams, *first, after.Offset)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	return &model.TeamConnection{
		TotalCount: len(teams),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     len(teams) > *first+after.Offset,
			HasPreviousPage: after.Offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// User returns UserResolver implementation.
func (r *Resolver) User() UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
