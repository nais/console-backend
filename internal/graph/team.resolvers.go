package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.29

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/console"
	"github.com/nais/console-backend/internal/graph/model"
)

// Teams is the resolver for the teams field.
func (r *queryResolver) Teams(ctx context.Context, first *int, after *model.Cursor) (*model.TeamConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &model.Cursor{Offset: 0}
	}

	teams, err := r.Console.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Console: %w", err)
	}
	if *first > len(teams) {
		*first = len(teams)
	}

	e := teamEdges(teams, *first, after.Offset)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	return &model.TeamConnection{
		TotalCount: len(teams),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     len(teams) > *first+after.Offset,
			HasPreviousPage: after.Offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Team is the resolver for the team field.
func (r *queryResolver) Team(ctx context.Context, name string) (*model.Team, error) {
	team, err := r.Console.GetTeam(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("getting team from Console: %w", err)
	}

	return &model.Team{
		ID:           team.Slug,
		Name:         team.Slug,
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
		Description: &team.Purpose,
	}, nil
}

// Members is the resolver for the members field.
func (r *teamResolver) Members(ctx context.Context, obj *model.Team, first *int, after *model.Cursor) (*model.TeamMemberConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &model.Cursor{Offset: 0}
	}

	members, err := r.Console.GetMembers(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Console: %w", err)
	}
	if *first > len(members) {
		*first = len(members)
	}

	e := memberEdges(members, *first, after.Offset)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	return &model.TeamMemberConnection{
		TotalCount: len(members),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     len(members) > *first+after.Offset,
			HasPreviousPage: after.Offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Apps is the resolver for the apps field.
func (r *teamResolver) Apps(ctx context.Context, obj *model.Team, first *int, after *model.Cursor) (*model.AppConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &model.Cursor{Offset: 0}
	}

	apps, err := r.K8s.Apps(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting apps from Kubernetes: %w", err)
	}

	if *first > len(apps) {
		*first = len(apps)
	}

	a := appEdges(apps, obj.Name, *first, after.Offset)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(a) > 0 {
		startCursor = &a[0].Cursor
		endCursor = &a[len(a)-1].Cursor
	}

	return &model.AppConnection{
		TotalCount: len(apps),
		Edges:      a,
		PageInfo: &model.PageInfo{
			HasNextPage:     len(apps) > *first+after.Offset,
			HasPreviousPage: after.Offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// GithubRepositories is the resolver for the githubRepositories field.
func (r *teamResolver) GithubRepositories(ctx context.Context, obj *model.Team, first *int, after *model.Cursor) (*model.GithubRepositoryConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &model.Cursor{Offset: 0}
	}

	repos, err := r.Console.GetGithubRepositories(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Console: %w", err)
	}
	if *first > len(repos) {
		*first = len(repos)
	}

	e := githubRepositoryEdges(repos, *first, after.Offset)

	var startCursor *model.Cursor
	var endCursor *model.Cursor

	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	return &model.GithubRepositoryConnection{
		TotalCount: len(repos),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     len(repos) > *first+after.Offset,
			HasPreviousPage: after.Offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Deployments is the resolver for the deployments field.
func (r *teamResolver) Deployments(ctx context.Context, obj *model.Team, first *int, after *model.Cursor) (*model.DeploymentConnection, error) {
	deploys, err := r.Hookd.GetDeploysForTeam(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting team deploys from Hookd: %w", err)
	}

	fmt.Println("deploys", deploys)

	return nil, nil
}

// Team returns TeamResolver implementation.
func (r *Resolver) Team() TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *teamResolver) Instances(ctx context.Context) (*model.Instance, error) {
	return &model.Instance{}, nil
}
