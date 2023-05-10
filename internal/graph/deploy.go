package graph

import (
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/hookd"
)

func deployEdges(deploys []hookd.Deploy, first int, after int) []*model.DeploymentEdge {
	edges := []*model.DeploymentEdge{}
	limit := first + after
	if limit > len(deploys) {
		limit = len(deploys)
	}
	for i := after; i < limit; i++ {
		deploy := deploys[i]
		edges = append(edges, &model.DeploymentEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node: &model.Deployment{
				ID:         deploy.DeploymentInfo.ID,
				Statuses:   mapStatuses(deploy.Statuses),
				Resources:  mapResources(deploy.Resources),
				Type:       "Application",
				Team:       &model.Team{Name: deploy.DeploymentInfo.Team},
				Env:        deploy.DeploymentInfo.Cluster,
				Created:    deploy.DeploymentInfo.Created,
				Repository: deploy.DeploymentInfo.GithubRepository,
			},
		})
	}
	return edges
}

func mapResources(resources []hookd.Resource) []*model.DeploymentResource {
	ret := []*model.DeploymentResource{}
	for _, resource := range resources {
		ret = append(ret, &model.DeploymentResource{
			ID:        resource.ID,
			Group:     resource.Group,
			Kind:      resource.Kind,
			Name:      resource.Name,
			Namespace: resource.Namespace,
			Version:   resource.Version,
		})
	}
	return ret
}

func mapStatuses(statuses []hookd.Status) []*model.DeploymentStatus {
	ret := []*model.DeploymentStatus{}
	for _, status := range statuses {
		ret = append(ret, &model.DeploymentStatus{
			ID:      status.ID,
			Status:  status.Status,
			Message: &status.Message,
			Created: status.Created,
		})
	}
	return ret
}
