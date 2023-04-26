package graph

import (
	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph/model"
)

func teamEdges(teams []console.Team, first int, after int) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	limit := first + after
	if limit > len(teams) {
		limit = len(teams)
	}
	for i := after; i < limit; i++ {
		team := teams[i]
		edges = append(edges, &model.TeamEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node: &model.Team{
				ID:          team.Slug,
				Name:        team.Slug,
				Description: &team.Purpose,
			},
		})
	}
	return edges
}

func appEdges(apps []*model.App, team string, first int, after int) []*model.AppEdge {
	edges := []*model.AppEdge{}
	limit := first + after
	if limit > len(apps) {
		limit = len(apps)
	}
	for i := after; i < limit; i++ {
		app := apps[i]
		app.GQLVars = struct{ Team string }{
			Team: team,
		}

		edges = append(edges, &model.AppEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node:   app,
		})
	}
	return edges
}

func memberEdges(members []console.Member, first int, after int) []*model.TeamMemberEdge {
	edges := []*model.TeamMemberEdge{}
	limit := first + after
	if limit > len(members) {
		limit = len(members)
	}
	for i := after; i < limit; i++ {
		member := members[i]
		edges = append(edges, &model.TeamMemberEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node: &model.TeamMember{
				ID:    member.User.Email,
				Email: member.User.Email,
				Name:  member.User.Name,
				Role:  member.Role,
			},
		})
	}
	return edges
}

func githubRepositoryEdges(repos []console.GitHubRepository, first int, after int) []*model.GithubRepositoryEdge {
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
