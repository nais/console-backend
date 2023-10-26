package model

import (
	"time"

	"github.com/nais/console-backend/internal/graph/scalar"
)

type AccessPolicy struct {
	Inbound  Inbound  `json:"inbound,omitempty"`
	Outbound Outbound `json:"outbound,omitempty"`
}

type App struct {
	ID           scalar.Ident `json:"id"`
	Name         string       `json:"name"`
	Image        string       `json:"image"`
	DeployInfo   DeployInfo   `json:"deployInfo"`
	Env          *Env         `json:"env"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
	Authz        []Authz      `json:"authz"`
	AutoScaling  AutoScaling  `json:"autoScaling"`
	Deployed     time.Time    `json:"deployed"`
	Ingresses    []string     `json:"ingresses"`
	Resources    Resources    `json:"resources"`
	Storage      []Storage    `json:"storage"`
	Variables    []Variable   `json:"variables"`
	AppState     AppState     `json:"state"`
	GQLVars      AppGQLVars   `json:"-"`
}

func (App) IsSearchNode()         {}
func (App) IsNode()               {}
func (a App) GetID() scalar.Ident { return a.ID }
