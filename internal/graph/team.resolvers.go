package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/console-backend/internal/auth"
	"github.com/nais/console-backend/internal/dependencytrack"
	"github.com/nais/console-backend/internal/graph/apierror"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/model/vulnerabilities"
	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/nais/console-backend/internal/hookd"
)

// ChangeDeployKey is the resolver for the changeDeployKey field.
func (r *mutationResolver) ChangeDeployKey(ctx context.Context, team string) (*model.DeploymentKey, error) {
	if !r.hasAccess(ctx, team) {
		return nil, fmt.Errorf("access denied")
	}

	deployKey, err := r.hookdClient.ChangeDeployKey(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("changing deploy key in Hookd: %w", err)
	}
	return &model.DeploymentKey{
		ID:      scalar.DeployKeyIdent(team),
		Key:     deployKey.Key,
		Created: deployKey.Created,
		Expires: deployKey.Expires,
	}, nil
}

// Teams is the resolver for the teams field.
func (r *queryResolver) Teams(ctx context.Context, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor) (*model.TeamConnection, error) {
	teams, err := r.teamsClient.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Teams: %w", err)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	e := teamEdges(teams, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
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
		PageInfo: model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Team is the resolver for the team field.
func (r *queryResolver) Team(ctx context.Context, name string) (*model.Team, error) {
	team, err := r.teamsClient.GetTeam(ctx, name)
	if err != nil {
		return nil, apierror.ErrTeamNotFound
	}

	return team, nil
}

// Members is the resolver for the members field.
func (r *teamResolver) Members(ctx context.Context, obj *model.Team, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor) (*model.TeamMemberConnection, error) {
	members, err := r.teamsClient.GetTeamMembers(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting members from Teams: %w", err)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	edges := memberEdges(members, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(members) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.TeamMemberConnection{
		TotalCount: len(members),
		Edges:      edges,
		PageInfo: model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Apps is the resolver for the apps field.
func (r *teamResolver) Apps(ctx context.Context, obj *model.Team, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor, orderBy *model.OrderBy) (*model.AppConnection, error) {
	apps, err := r.k8sClient.Apps(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting apps from Kubernetes: %w", err)
	}
	if orderBy != nil {
		switch orderBy.Field {
		case "NAME":
			model.SortWith(apps, func(a, b *model.App) bool {
				return model.Compare(a.Name, b.Name, orderBy.Direction)
			})
		case "ENV":
			model.SortWith(apps, func(a, b *model.App) bool {
				return model.Compare(a.Env.Name, b.Env.Name, orderBy.Direction)
			})
		}
	}
	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	edges := appEdges(apps, obj.Name, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(apps) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.AppConnection{
		TotalCount: len(apps),
		Edges:      edges,
		PageInfo: model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Naisjobs is the resolver for the naisjobs field.
func (r *teamResolver) Naisjobs(ctx context.Context, obj *model.Team, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor) (*model.NaisJobConnection, error) {
	naisjobs, err := r.k8sClient.NaisJobs(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting naisjobs from Kubernetes: %w", err)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	edges := naisJobEdges(naisjobs, obj.Name, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(naisjobs) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.NaisJobConnection{
		TotalCount: len(naisjobs),
		Edges:      edges,
		PageInfo: model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// GithubRepositories is the resolver for the githubRepositories field.
func (r *teamResolver) GithubRepositories(ctx context.Context, obj *model.Team, first *int, after *scalar.Cursor) (*model.GithubRepositoryConnection, error) {
	if first == nil {
		first = new(int)
		*first = 10
	}
	if after == nil {
		after = &scalar.Cursor{Offset: 0}
	}

	repos, err := r.teamsClient.GetGithubRepositories(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting teams from Teams: %w", err)
	}
	if *first > len(repos) {
		*first = len(repos)
	}

	edges := githubRepositoryEdges(repos, *first, after.Offset)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor

	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.GithubRepositoryConnection{
		TotalCount: len(repos),
		Edges:      edges,
		PageInfo: model.PageInfo{
			HasNextPage:     len(repos) > *first+after.Offset,
			HasPreviousPage: after.Offset > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Deployments is the resolver for the deployments field.
func (r *teamResolver) Deployments(ctx context.Context, obj *model.Team, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor, limit *int) (*model.DeploymentConnection, error) {
	if limit == nil {
		limit = new(int)
		*limit = 10
	}

	deploys, err := r.hookdClient.Deployments(ctx, hookd.WithTeam(obj.Name), hookd.WithLimit(*limit))
	if err != nil {
		return nil, fmt.Errorf("getting deploys from Hookd: %w", err)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	edges := deployEdges(deploys, pagination)

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(deploys) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.DeploymentConnection{
		TotalCount: len(deploys),
		Edges:      edges,
		PageInfo: model.PageInfo{
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

	key, err := r.hookdClient.DeployKey(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting deploy key from Hookd: %w", err)
	}

	return &model.DeploymentKey{
		ID:      scalar.DeployKeyIdent(obj.Name),
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

	members, err := r.teamsClient.GetTeamMembers(ctx, obj.Name)
	if err != nil {
		return false, fmt.Errorf("getting teams from Teams: %w", err)
	}

	for _, m := range members {
		if strings.EqualFold(m.User.Email, email) {
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

	members, err := r.teamsClient.GetTeamMembers(ctx, obj.Name)
	if err != nil {
		return false, fmt.Errorf("getting members from Teams: %w", err)
	}

	for _, m := range members {
		if strings.EqualFold(m.User.Email, email) {
			if m.Role == "OWNER" {
				return true, nil
			}
		}
	}
	return false, nil
}

// Vulnerabilities is the resolver for the vulnerabilities field.
func (r *teamResolver) Vulnerabilities(ctx context.Context, obj *model.Team, first *int, last *int, after *scalar.Cursor, before *scalar.Cursor, orderBy *model.OrderBy) (*model.VulnerabilitiesConnection, error) {
	apps, err := r.k8sClient.Apps(ctx, obj.Name)
	if err != nil {
		return nil, fmt.Errorf("getting apps from Kubernetes: %w", err)
	}

	instances := make([]*dependencytrack.AppInstance, 0)
	for _, app := range apps {
		instances = append(instances, &dependencytrack.AppInstance{
			Env:   app.Env.Name,
			App:   app.Name,
			Image: app.Image,
			Team:  obj.Name,
		})
	}

	nodes, err := r.dependencyTrackClient.GetVulnerabilities(ctx, instances)
	if err != nil {
		return nil, fmt.Errorf("getting vulnerabilities from DependencyTrack: %w", err)
	}

	if orderBy != nil {
		vulnerabilities.Sort(nodes, orderBy.Field, orderBy.Direction)
	}

	pagination, err := model.NewPagination(first, last, after, before)
	if err != nil {
		return nil, err
	}
	edges := make([]model.VulnerabilitiesEdge, 0)
	start, end := pagination.ForSlice(len(nodes))

	for i, n := range nodes[start:end] {
		edges = append(edges, model.VulnerabilitiesEdge{
			Cursor: scalar.Cursor{Offset: start + i},
			Node:   *n,
		})
	}

	var startCursor *scalar.Cursor
	var endCursor *scalar.Cursor
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNext := len(nodes) > pagination.First()+pagination.After().Offset+1
	hasPrevious := pagination.After().Offset > 0

	if pagination.Before() != nil && startCursor != nil {
		hasNext = true
		hasPrevious = startCursor.Offset > 0
	}

	return &model.VulnerabilitiesConnection{
		TotalCount: len(nodes),
		Edges:      edges,
		PageInfo: model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrevious,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}, nil
}

// Team returns TeamResolver implementation.
func (r *Resolver) Team() TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
