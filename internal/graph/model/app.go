package model

import (
	"time"
)

type External struct {
	Host  string `json:"host"`
	Ports []Port `json:"ports"`
}

type Outbound struct {
	Rules    []*Rule     `json:"rules"`
	External []*External `json:"external"`
}

type Inbound struct {
	Rules []*Rule `json:"rules"`
}

type AccessPolicy struct {
	Inbound  Inbound  `json:"inbound"`
	Outbound Outbound `json:"outbound"`
}

type Replicas struct {
	MinReplicas            int  `json:"min"`
	MaxReplicas            int  `json:"max"`
	DisableAutoScaling     bool `json:"disableAutoScaling"`
	CPUThresholdPercentage int  `json:"cpuThresholdPercentage"`
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
	Replicas     Replicas     `json:"replicas"`
	Resources    Resources    `json:"resources"`
	Storage      []Storage    `json:"storage"`
	Variables    []Variable   `json:"variables"`
	State        AppState     `json:"state"`
	Messages     []string     `json:"messages"`

	GQLVars struct {
		Team string
	} `json:"-"`
}

func (App) IsSearchNode()  {}
func (App) IsNode()        {}
func (a App) GetID() Ident { return a.ID }
