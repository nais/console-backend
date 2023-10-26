package model

import (
	"github.com/nais/console-backend/internal/graph/scalar"
)

type NaisJob struct {
	ID           scalar.Ident   `json:"id"`
	Name         string         `json:"name"`
	Env          *Env           `json:"env"`
	DeployInfo   *DeployInfo    `json:"deployInfo"`
	Image        string         `json:"image"`
	AccessPolicy *AccessPolicy  `json:"accessPolicy"`
	Resources    *Resources     `json:"resources"`
	Storage      []Storage      `json:"storage"`
	Authz        []Authz        `json:"authz"`
	Schedule     string         `json:"schedule"`
	Completions  int            `json:"completions"`
	Parallelism  int            `json:"parallelism"`
	Retries      int            `json:"retries"`
	JobState     JobState       `json:"jobState"`
	GQLVars      NaisJobGQLVars `json:"-"`
}

func (NaisJob) IsNode()               {}
func (j NaisJob) GetID() scalar.Ident { return j.ID }
func (NaisJob) IsSearchNode()         {}
