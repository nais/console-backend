package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/nais/console-backend/internal/hookd"
)

// Deployments is the resolver for the deployments field.
func (r *queryResolver) Deployments(ctx context.Context, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor, limit *int) (*model.DeploymentConnection, error) {
	l := 100
	if limit != nil {
		l = *limit
	}
	deploys, err := r.hookdClient.Deployments(ctx, hookd.WithLimit(l), hookd.WithIgnoreTeams("nais-verification"))
	if err != nil {
		return nil, fmt.Errorf("getting deploys from Hookd: %w", err)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	e := deployEdges(deploys, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	hasNext := len(deploys) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.DeploymentConnection{
		TotalCount: len(deploys),
		Edges:      e,
		PageInfo: model.PageInfo{
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}, nil
}
