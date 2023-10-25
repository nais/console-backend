package graph

import (
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/hookd"
)

func deployEdges(deploys []hookd.Deploy, p *model.Pagination) []*model.DeploymentEdge {
	edges := make([]*model.DeploymentEdge, 0)

	start, end := p.ForSlice(len(deploys))

	for i, deploy := range deploys[start:end] {
		deploy := deploy
		edges = append(edges, &model.DeploymentEdge{
			Cursor: model.Cursor{Offset: start + i},
			Node: &model.Deployment{
				ID:        model.DeploymentIdent(deploy.DeploymentInfo.ID),
				Statuses:  mapStatuses(deploy.Statuses),
				Resources: mapResources(deploy.Resources),
				Team: &model.Team{
					Name: deploy.DeploymentInfo.Team,
					ID:   model.TeamIdent(deploy.DeploymentInfo.Team),
				},
				Env:        deploy.DeploymentInfo.Cluster,
				Created:    deploy.DeploymentInfo.Created,
				Repository: deploy.DeploymentInfo.GithubRepository,
			},
		})

	}

	return edges
}

func mapResources(resources []hookd.Resource) []*model.DeploymentResource {
	ret := make([]*model.DeploymentResource, 0)
	for _, resource := range resources {
		ret = append(ret, &model.DeploymentResource{
			ID:        model.DeploymentResourceIdent(resource.ID),
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
	ret := make([]*model.DeploymentStatus, 0)
	for _, status := range statuses {
		ret = append(ret, &model.DeploymentStatus{
			ID:      model.DeploymentStatusIdent(status.ID),
			Status:  status.Status,
			Message: &status.Message,
			Created: status.Created,
		})
	}
	return ret
}
