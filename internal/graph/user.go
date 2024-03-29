package graph

import (
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	t "github.com/nais/console-backend/internal/teams"
)

func userTeamEdges(teams []t.TeamMembership, p *model.Pagination) []model.TeamEdge {
	edges := make([]model.TeamEdge, 0)
	start, end := p.ForSlice(len(teams))

	for i, team := range teams[start:end] {
		team := team
		edges = append(edges, model.TeamEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node: model.Team{
				ID:           scalar.TeamIdent(team.Team.Slug),
				Name:         team.Team.Slug,
				Description:  team.Team.Purpose,
				SlackChannel: team.Team.SlackChannel,
				SlackAlertsChannels: func(t []t.SlackAlertsChannel) []model.SlackAlertsChannel {
					ret := make([]model.SlackAlertsChannel, 0)
					for _, v := range t {
						ret = append(ret, model.SlackAlertsChannel{
							Env:  v.Environment,
							Name: v.ChannelName,
						})
					}
					return ret
				}(team.Team.SlackAlertsChannels),
			},
		})
	}

	return edges
}
