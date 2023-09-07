package model

import (
	"time"
)

type Deployment struct {
	ID         string                `json:"id"`
	Team       *Team                 `json:"team"`
	Type       string                `json:"type"`
	Resources  []*DeploymentResource `json:"resources"`
	Env        string                `json:"env"`
	Statuses   []*DeploymentStatus   `json:"statuses"`
	Created    time.Time             `json:"created"`
	Repository string                `json:"repository"`
}

type DeploymentConnection struct {
	TotalCount int               `json:"totalCount"`
	PageInfo   *PageInfo         `json:"pageInfo"`
	Edges      []*DeploymentEdge `json:"edges"`
}

func (d *DeploymentConnection) IsDeploymentResponse() {}

type DeploymentResource struct {
	ID        string `json:"id"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
}

type DeploymentStatus struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Message *string   `json:"message,omitempty"`
	Created time.Time `json:"created"`
}

type DeployInfo struct {
	Deployer  string     `json:"deployer"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
	CommitSha string     `json:"commitSha"`
	URL       string     `json:"url"`
	GQLVars   struct {
		App  string
		Job  string
		Env  string
		Team string
	} `json:"-"`
}

type Error struct {
	Message string `json:"message"`
}

func (e *Error) IsDeploymentResponse() {}
