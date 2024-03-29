package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/apierror"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
)

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	email, err := auth.GetEmail(ctx)
	if err != nil {
		return nil, apierror.ErrNoEmailInSession
	}

	user, err := r.teamsClient.GetUser(ctx, email)
	if err != nil {
		return nil, apierror.ErrUserNotFound(email)
	}

	return &model.User{
		ID:    scalar.UserIdent(user.ID.String()),
		Name:  user.Name,
		Email: email,
	}, nil
}

// Teams is the resolver for the teams field.
func (r *userResolver) Teams(ctx context.Context, obj *model.User, first *int, after *scalar.Cursor, last *int, before *scalar.Cursor) (*model.TeamConnection, error) {
	teams, err := r.teamsClient.GetTeamsForUser(ctx, obj.Email)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Teams: %w", err)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}

	edges := userTeamEdges(teams, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(teams) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.TeamConnection{
		TotalCount: len(teams),
		Edges:      edges,
		PageInfo: model.PageInfo{
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
