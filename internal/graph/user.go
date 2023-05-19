package graph

import (
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph/model"
)

func userTeamEdges(teams []console.TeamMembership, p *model.Pagination) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	start, end := p.ForSlice(len(teams))

	for i, team := range teams[start:end] {
		team := team
		edges = append(edges, &model.TeamEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node: &model.Team{
				ID:           model.Ident{ID: team.Team.Slug, Type: "user_team"},
				Name:         team.Team.Slug,
				Description:  &team.Team.Purpose,
				SlackChannel: team.Team.SlackChannel,
				SlackAlertsChannels: func(t []console.SlackAlertsChannel) []model.SlackAlertsChannel {
					ret := []model.SlackAlertsChannel{}
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
