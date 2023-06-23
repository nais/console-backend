package graph

import (
	"context"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/model"
	t "github.com/nais/console-backend/internal/teams"
)

func teamEdges(teams []t.Team, p *model.Pagination) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	start, end := p.ForSlice(len(teams))

	for i, team := range teams[start:end] {
		team := team
		edges = append(edges, &model.TeamEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node: &model.Team{
				ID:           model.Ident{ID: team.Slug, Type: "teams_team"},
				Name:         team.Slug,
				Description:  &team.Purpose,
				SlackChannel: team.SlackChannel,
				SlackAlertsChannels: func(t []t.SlackAlertsChannel) []model.SlackAlertsChannel {
					ret := []model.SlackAlertsChannel{}
					for _, v := range t {
						ret = append(ret, model.SlackAlertsChannel{
							Env:  v.Environment,
							Name: v.ChannelName,
						})
					}
					return ret
				}(team.SlackAlertsChannels),
			},
		})
	}

	return edges
}

func appEdges(apps []*model.App, team string, p *model.Pagination) []*model.AppEdge {
	edges := []*model.AppEdge{}
	start, end := p.ForSlice(len(apps))

	for i, app := range apps[start:end] {
		app.GQLVars = struct{ Team string }{
			Team: team,
		}

		edges = append(edges, &model.AppEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node:   app,
		})
	}

	return edges
}

func memberEdges(members []t.Member, p *model.Pagination) []*model.TeamMemberEdge {
	edges := []*model.TeamMemberEdge{}

	start, end := p.ForSlice(len(members))

	for i, member := range members[start:end] {
		member := member
		edges = append(edges, &model.TeamMemberEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node: &model.TeamMember{
				ID:    model.Ident{ID: member.User.Email, Type: "user"},
				Name:  member.User.Name,
				Email: member.User.Email,
				Role:  model.TeamRole(member.Role),
			},
		})
	}

	return edges
}

func githubRepositoryEdges(repos []t.GitHubRepository, first int, after int) []*model.GithubRepositoryEdge {
	edges := []*model.GithubRepositoryEdge{}
	limit := first + after
	if limit > len(repos) {
		limit = len(repos)
	}
	for i := after; i < limit; i++ {
		repo := repos[i]
		edges = append(edges, &model.GithubRepositoryEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node: &model.GithubRepository{
				Name: repo.Name,
			},
		})
	}
	return edges
}

func (r *Resolver) hasAccess(ctx context.Context, team string) bool {
	email, err := auth.GetEmail(ctx)
	if err != nil {
		r.Log.Errorf("getting email from context: %v", err)
		return false
	}

	teams, err := r.TeamsClient.GetTeamsForUser(ctx, email)
	if err != nil {
		r.Log.Errorf("getting teams from Teams: %v", err)
		return false
	}

	for _, t := range teams {
		if t.Team.Slug == team {
			return true
		}
	}

	return false
}
