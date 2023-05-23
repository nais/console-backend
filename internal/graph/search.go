package graph

import (
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/search"
)

func searchEdges(results []*search.SearchResult, p *model.Pagination) []*model.SearchEdge {
	edges := []*model.SearchEdge{}
	start, end := p.ForSlice(len(results))

	for i, res := range results[start:end] {
		edges = append(edges, &model.SearchEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node:   res.Node,
		},
		)
	}
	return edges
}
