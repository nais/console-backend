package graph

import (
	"context"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	t "github.com/nais/console-backend/internal/teams"
)


func naisJobEdges(naisjobs []*model.NaisJob, team string, p *model.Pagination) []model.NaisJobEdge {
	edges := make([]model.NaisJobEdge, 0)
	start, end := p.ForSlice(len(naisjobs))

	for i, job := range naisjobs[start:end] {
		job.GQLVars = model.NaisJobGQLVars{Team: team}

		edges = append(edges, model.NaisJobEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node:   *job,
		})
	}

	return edges
}


func (r *Resolver) hasAccess(ctx context.Context, teamName string) bool {
	email, err := auth.GetEmail(ctx)
	if err != nil {
		r.log.Errorf("getting email from context: %v", err)
		return false
	}

	teams, err := r.teamsClient.GetTeamsForUser(ctx, email)
	if err != nil {
		r.log.Errorf("getting teams from Teams: %v", err)
		return false
	}

	for _, team := range teams {
		if team.Team.Slug == teamName {
			return true
		}
	}

	return false
}
