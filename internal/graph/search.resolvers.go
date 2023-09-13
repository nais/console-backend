package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"

	"github.com/nais/console-backend/internal/graph/model"
)

// Search is the resolver for the search field.
func (r *queryResolver) Search(ctx context.Context, query string, filter *model.SearchFilter, first *int, last *int, after *model.Cursor, before *model.Cursor) (*model.SearchConnection, error) {
	results := r.Searcher.Search(ctx, query, filter)
	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	edges := searchEdges(results, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(results) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}
	return &model.SearchConnection{
		TotalCount: len(results),
		Edges:      edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}
