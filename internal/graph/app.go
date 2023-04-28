package graph

import (
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/hookd"
)

func deploymentEdges(deps []hookd.Deploy, first int, after int) []*model.DeploymentEdge {
	edges := []*model.DeploymentEdge{}
	limit := first + after
	if limit > len(deps) {
		limit = len(deps)
	}
	for i := after; i < limit; i++ {
		dep := deps[i]
		edge := &model.DeploymentEdge{
			Cursor: model.Cursor{Offset: i + 1},
			Node: &model.Deployment{
				ID:      dep.DeploymentInfo.ID,
				Type:    "Application",
				Env:     dep.DeploymentInfo.Cluster,
				Created: dep.DeploymentInfo.Created,
			},
		}
		for _, status := range dep.Statuses {
			edge.Node.Statuses = append(edge.Node.Statuses, &model.DeploymentStatus{
				ID:      status.ID,
				Status:  status.Status,
				Message: &status.Message,
				Created: status.Created,
			})
		}
		for _, resource := range dep.Resources {
			edge.Node.Resources = append(edge.Node.Resources, &model.DeploymentResource{
				ID:        resource.ID,
				Group:     resource.Group,
				Kind:      resource.Kind,
				Name:      resource.Name,
				Namespace: resource.Namespace,
				Version:   resource.Version,
			})
		}

		edges = append(edges, edge)
	}
	return edges
}
