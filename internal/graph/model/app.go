package model

import (
	"time"
)

type AccessPolicy struct {
	Inbound  Inbound  `json:"inbound"`
	Outbound Outbound `json:"outbound"`
}

type App struct {
	ID           Ident        `json:"id"`
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

func (App) IsSearchNode()  {}
func (App) IsNode()        {}
func (a App) GetID() Ident { return a.ID }
