package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

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
	user, err := r.TeamsClient.GetUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("getting user from Teams: %w", err)
	}
	return &model.User{
		ID:    model.Ident{ID: user.ID.String(), Type: "user"},
		Name:  user.Name,
		Email: email,
	}, nil
}

// Teams is the resolver for the teams field.
func (r *userResolver) Teams(ctx context.Context, obj *model.User, first *int, after *model.Cursor, last *int, before *model.Cursor) (*model.TeamConnection, error) {
	teams, err := r.TeamsClient.GetTeamsForUser(ctx, obj.Email)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Teams: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	e := userTeamEdges(teams, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	hasNext := len(teams) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.TeamConnection{
		TotalCount: len(teams),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// User returns UserResolver implementation.
func (r *Resolver) User() UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
