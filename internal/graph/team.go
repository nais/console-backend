package graph

import (
	"context"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	t "github.com/nais/console-backend/internal/teams"
)

func teamEdges(teams []t.Team, p *model.Pagination) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	start, end := p.ForSlice(len(teams))

	for i, team := range teams[start:end] {
		team := team
		edges = append(edges, &model.TeamEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node: &model.Team{
				ID:           scalar.TeamIdent(team.Slug),
				Name:         team.Slug,
				Description:  team.Purpose,
				SlackChannel: team.SlackChannel,
				SlackAlertsChannels: func(t []t.SlackAlertsChannel) []*model.SlackAlertsChannel {
					ret := make([]*model.SlackAlertsChannel, 0)
					for _, v := range t {
						ret = append(ret, &model.SlackAlertsChannel{
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

func naisJobEdges(naisjobs []*model.NaisJob, team string, p *model.Pagination) []*model.NaisJobEdge {
	edges := []*model.NaisJobEdge{}
	start, end := p.ForSlice(len(naisjobs))

	for i, job := range naisjobs[start:end] {
		job.GQLVars = model.NaisJobGQLVars{Team: team}

		edges = append(edges, &model.NaisJobEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node:   job,
		})
	}

	return edges
}

func appEdges(apps []*model.App, team string, p *model.Pagination) []*model.AppEdge {
	edges := []*model.AppEdge{}
	start, end := p.ForSlice(len(apps))

	for i, app := range apps[start:end] {
		app.GQLVars = model.AppGQLVars{Team: team}

		edges = append(edges, &model.AppEdge{
			Cursor: scalar.Cursor{Offset: start + i},
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
			Cursor: scalar.Cursor{Offset: start + i},
			Node: &model.TeamMember{
				ID:    scalar.UserIdent(member.User.Email),
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
			Cursor: scalar.Cursor{Offset: i + 1},
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
		r.log.Errorf("getting email from context: %v", err)
		return false
	}

	teams, err := r.teamsClient.GetTeamsForUser(ctx, email)
	if err != nil {
		r.log.Errorf("getting teams from Teams: %v", err)
		return false
	}

	for _, t := range teams {
		if t.Team.Slug == team {
			return true
		}
	}

	return false
}
