package graph

import (
	"context"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	t "github.com/nais/console-backend/internal/teams"
)

func teamEdges(teams []t.Team, p *model.Pagination) []model.TeamEdge {
	edges := make([]model.TeamEdge, 0)
	start, end := p.ForSlice(len(teams))

	for i, team := range teams[start:end] {
		team := team
		edges = append(edges, model.TeamEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node: model.Team{
				ID:           scalar.TeamIdent(team.Slug),
				Name:         team.Slug,
				Description:  team.Purpose,
				SlackChannel: team.SlackChannel,
				SlackAlertsChannels: func(t []t.SlackAlertsChannel) []model.SlackAlertsChannel {
					ret := make([]model.SlackAlertsChannel, 0)
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

func appEdges(apps []*model.App, team string, p *model.Pagination) []model.AppEdge {
	edges := make([]model.AppEdge, 0)
	start, end := p.ForSlice(len(apps))

	for i, app := range apps[start:end] {
		app.GQLVars = model.AppGQLVars{Team: team}

		edges = append(edges, model.AppEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node:   *app,
		})
	}

	return edges
}

func memberEdges(members []t.Member, p *model.Pagination) []model.TeamMemberEdge {
	edges := make([]model.TeamMemberEdge, 0)

	start, end := p.ForSlice(len(members))

	for i, member := range members[start:end] {
		member := member
		edges = append(edges, model.TeamMemberEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node: model.TeamMember{
				ID:    scalar.UserIdent(member.User.Email),
				Name:  member.User.Name,
				Email: member.User.Email,
				Role:  model.TeamRole(member.Role),
			},
		})
	}

	return edges
}

func githubRepositoryEdges(repos []t.GitHubRepository, p *model.Pagination) []model.GithubRepositoryEdge {
	edges := make([]model.GithubRepositoryEdge, 0)
	start, end := p.ForSlice(len(repos))

	for i, repo := range repos[start:end] {
		repo := repo
		edges = append(edges, model.GithubRepositoryEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node: model.GithubRepository{
				Name:           repo.Name,
				Archived:       repo.Archived,
				Permissions:    mapPermissions(repo.Permissions),
				Authorizations: mapAuthorizations(repo.Authorizations),
			},
		})
	}
	return edges
}

func mapPermissions(permissions []t.GitHubRepositoryPermission) []string {
	ret := []string{}
	for _, p := range permissions {
		if p.Granted {
			ret = append(ret, p.Name)
		}
	}
	return ret
}

func mapAuthorizations(authorizations []t.RepositoryAuthorization) []model.RepositoryAuthorization {
	ret := []model.RepositoryAuthorization{}
	for _, a := range authorizations {
		ret = append(ret, model.RepositoryAuthorization(a))
	}
	return ret
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
