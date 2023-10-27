package graph

import (
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/nais/console-backend/internal/search"
)

func searchEdges(results []*search.Result, p *model.Pagination) []model.SearchEdge {
	edges := make([]model.SearchEdge, 0)
	start, end := p.ForSlice(len(results))

	for i, res := range results[start:end] {
		edges = append(edges, model.SearchEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node:   res.Node,
		},
		)
	}
	return edges
}
