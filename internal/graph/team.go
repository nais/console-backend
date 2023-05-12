package graph

import (
	"fmt"

	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph/model"
)

func teamEdges(teams []console.Team, first, last int, before *model.Cursor, after int) []*model.TeamEdge {
	edges := []*model.TeamEdge{}
	limit := first + after
	if limit > len(teams) {
		limit = len(teams)
	}
	if before != nil {
		fmt.Print("before and last")
		limit = last + before.Offset
		for i := before.Offset; i < limit; i-- {
			team := teams[i]
			edges = append(edges, &model.TeamEdge{
				Cursor: model.Cursor{Offset: i - 1},
				Node: &model.Team{
					ID:          model.Ident{ID: team.Slug, Type: "team"},
					Name:        team.Slug,
					Description: &team.Purpose,
				},
			})
		}
	} else {
		for i := after; i < limit; i++ {
			team := teams[i]
			edges = append(edges, &model.TeamEdge{
				Cursor: model.Cursor{Offset: i + 1},
				Node: &model.Team{
					ID:          model.Ident{ID: team.Slug, Type: "team"},
					Name:        team.Slug,
					Description: &team.Purpose,
				},
			})
		}
	}
	return edges
}

func appEdges(apps []*model.App, team string, p *model.Pagination) []*model.AppEdge {
	edges := []*model.AppEdge{}
	if p.Before() != nil {
		i := p.Before().Offset - p.Last()
		if i < 0 {
			i = 0
		}

		for ; i < p.Limit(); i++ {
			app := apps[i]
			app.GQLVars = struct{ Team string }{
				Team: team,
			}

			edges = append(edges, &model.AppEdge{
				Cursor: model.Cursor{Offset: i},
				Node:   app,
			})
		}
		return edges
	}

	for i := p.After().Offset + 1; i < p.Limit(); i++ {
		app := apps[i]
		app.GQLVars = struct{ Team string }{
			Team: team,
		}

		edges = append(edges, &model.AppEdge{
			Cursor: model.Cursor{Offset: i},
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
				ID:    model.Ident{ID: member.User.Email, Type: "user"},
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
