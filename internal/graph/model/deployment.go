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

type DeploymentStatus struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Message *string   `json:"message,omitempty"`
	Created time.Time `json:"created"`
}

type DeployInfoGQLVars struct {
	App  string
	Job  string
	Env  string
	Team string
}
