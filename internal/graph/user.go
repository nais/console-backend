package graph

import (
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph/model"
)

func edges(teams []console.TeamMembership, first int, after int) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	limit := first + after
	if limit > len(teams) {
		limit = len(teams)
	}
	for i := after; i < limit; i++ {
		team := teams[i].Team
		edges = append(edges, &model.TeamEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node: &model.Team{
				ID:          model.Ident{ID: team.Slug, Type: "team"},
				Name:        team.Slug,
				Description: &team.Purpose,
			},
		})
	}
	return edges
}
