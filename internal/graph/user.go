package graph

import (
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph/model"
)

func edges(teams []console.TeamMembership, p *model.Pagination) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	start, end := p.ForSlice(len(teams))

	for i, team := range teams[start:end] {
		edges = append(edges, &model.TeamEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node:   consoleTeamToModelTeam(team.Team),
		})
	}

	return edges
}

func consoleTeamToModelTeam(team console.Team) *model.Team {
	return &model.Team{
		ID:           model.Ident{ID: team.Slug, Type: "team"},
		Name:         team.Slug,
		Description:  &team.Purpose,
		SlackChannel: team.SlackChannel,
		SlackAlertsChannels: func(t []console.SlackAlertsChannel) []model.SlackAlertsChannel {
			ret := []model.SlackAlertsChannel{}
			for _, v := range t {
				ret = append(ret, model.SlackAlertsChannel{
					Env:  v.Environment,
					Name: v.ChannelName,
				})
			}
			return ret
		}(team.SlackAlertsChannels),
	}
}
