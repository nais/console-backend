package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.30

import (
	"context"
	"fmt"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/hookd"
)

// ChangeDeployKey is the resolver for the changeDeployKey field.
func (r *mutationResolver) ChangeDeployKey(ctx context.Context, team string) (*model.DeploymentKey, error) {
	if !r.hasAccess(ctx, team) {
		return nil, fmt.Errorf("access denied")
	}

	new, err := r.Hookd.ChangeDeployKey(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("changing deploy key in Hookd: %w", err)
	}
	return &model.DeploymentKey{
		ID:      model.Ident{ID: team, Type: "deployKey"},
		Key:     new.Key,
		Created: new.Created,
		Expires: new.Expires,
	}, nil
}

// Teams is the resolver for the teams field.
func (r *queryResolver) Teams(ctx context.Context, first *int, last *int, after *model.Cursor, before *model.Cursor) (*model.TeamConnection, error) {
	teams, err := r.TeamsClient.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Teams: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	e := teamEdges(teams, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	hasNext := len(teams) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.TeamConnection{
		TotalCount: len(teams),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Team is the resolver for the team field.
func (r *queryResolver) Team(ctx context.Context, name string) (*model.Team, error) {
	team, err := r.TeamsClient.GetTeam(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("getting team from Teams: %w", err)
	}

	return team, nil
}

// Members is the resolver for the members field.
func (r *teamResolver) Members(ctx context.Context, obj *model.Team, first *int, after *model.Cursor, last *int, before *model.Cursor) (*model.TeamMemberConnection, error) {
	members, err := r.TeamsClient.GetMembers(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting members from Teams: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	e := memberEdges(members, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	hasNext := len(members) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.TeamMemberConnection{
		TotalCount: len(members),
		Edges:      e,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Apps is the resolver for the apps field.
func (r *teamResolver) Apps(ctx context.Context, obj *model.Team, first *int, last *int, after *model.Cursor, before *model.Cursor) (*model.AppConnection, error) {
	apps, err := r.K8s.Apps(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting apps from Kubernetes: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	a := appEdges(apps, obj.Name, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(a) > 0 {
		startCursor = &a[0].Cursor
		endCursor = &a[len(a)-1].Cursor
	}

	hasNext := len(apps) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.AppConnection{
		TotalCount: len(apps),
		Edges:      a,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Naisjobs is the resolver for the naisjobs field.
func (r *teamResolver) Naisjobs(ctx context.Context, obj *model.Team, first *int, last *int, after *model.Cursor, before *model.Cursor) (*model.NaisJobConnection, error) {
	naisjobs, err := r.K8s.NaisJobs(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting naisjobs from Kubernetes: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	j := naisJobEdges(naisjobs, obj.Name, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(j) > 0 {
		startCursor = &j[0].Cursor
		endCursor = &j[len(j)-1].Cursor
	}

	hasNext := len(naisjobs) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.NaisJobConnection{
		TotalCount: len(naisjobs),
		Edges:      j,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
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

	repos, err := r.TeamsClient.GetGithubRepositories(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Teams: %w", err)
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
func (r *teamResolver) Deployments(ctx context.Context, obj *model.Team, first *int, last *int, after *model.Cursor, before *model.Cursor, limit *int) (*model.DeploymentConnection, error) {
	if limit == nil {
		limit = new(int)
		*limit = 10
	}

	deploys, err := r.Hookd.Deployments(ctx, hookd.WithTeam(obj.Name), hookd.WithLimit(*limit))
	if err != nil {
		return nil, fmt.Errorf("getting deploys from Hookd: %w", err)
	}

	pagination := model.NewPagination(first, last, after, before)
	e := deployEdges(deploys, pagination)

	var startCursor *model.Cursor
	var endCursor *model.Cursor
	if len(e) > 0 {
		startCursor = &e[0].Cursor
		endCursor = &e[len(e)-1].Cursor
	}

	hasNext := len(deploys) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.DeploymentConnection{
		TotalCount: len(deploys),
		Edges:      e,
		PageInfo: &model.PageInfo{
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
		},
	}, nil
}

// DeployKey is the resolver for the deployKey field.
func (r *teamResolver) DeployKey(ctx context.Context, obj *model.Team) (*model.DeploymentKey, error) {
	if !r.hasAccess(ctx, obj.Name) {
		return nil, fmt.Errorf("access denied")
	}

	key, err := r.Hookd.DeployKey(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting deploy key from Hookd: %w", err)
	}

	return &model.DeploymentKey{
		ID:      model.Ident{ID: obj.Name, Type: "deployKey"},
		Key:     key.Key,
		Created: key.Created,
		Expires: key.Expires,
	}, nil
}

// ViewerIsMember is the resolver for the viewerIsMember field.
func (r *teamResolver) ViewerIsMember(ctx context.Context, obj *model.Team) (bool, error) {
	email, err := auth.GetEmail(ctx)
	if err != nil {
		return false, fmt.Errorf("getting email from context: %w", err)
	}

	members, err := r.TeamsClient.GetMembers(ctx, obj.Name)
	if err != nil {
		return false, fmt.Errorf("getting teams from Teams: %w", err)
	}

	for _, m := range members {
		if m.User.Email == email {
			if m.Role == "OWNER" || m.Role == "MEMBER" {
				return true, nil
			}
		}
	}

	return false, nil
}

// ViewerIsAdmin is the resolver for the viewerIsAdmin field.
func (r *teamResolver) ViewerIsAdmin(ctx context.Context, obj *model.Team) (bool, error) {
	email, err := auth.GetEmail(ctx)
	if err != nil {
		return false, fmt.Errorf("getting email from context: %w", err)
	}

	members, err := r.TeamsClient.GetMembers(ctx, obj.Name)
	if err != nil {
		return false, fmt.Errorf("getting members from Teams: %w", err)
	}

	for _, m := range members {
		if m.User.Email == email {
			if m.Role == "OWNER" {
				return true, nil
			}
		}
	}
	return false, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Team returns TeamResolver implementation.
func (r *Resolver) Team() TeamResolver { return &teamResolver{r} }

type mutationResolver struct{ *Resolver }
type teamResolver struct{ *Resolver }
