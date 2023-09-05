package model

import (
	"time"
)

type AccessPolicyRule struct {
	Application string `json:"application"`
	Namespace   string `json:"namespace"`
}

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

type Port struct {
	Port int `json:"port"`
}

type Rule struct {
	Application string `json:"application"`
	Namespace   string `json:"namespace"`
	Cluster     string `json:"cluster"`
}

type Limits struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
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
	AppState     AppState     `json:"state"`

	GQLVars struct {
		Team string
	} `json:"-"`
}

func (App) IsSearchNode()  {}
func (App) IsNode()        {}
func (a App) GetID() Ident { return a.ID }

type AppConnection struct {
	TotalCount int        `json:"totalCount"`
	PageInfo   *PageInfo  `json:"pageInfo"`
	Edges      []*AppEdge `json:"edges"`
}

type AppEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *App   `json:"node"`
}

type Requests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Resources struct {
	Limits   *Limits   `json:"limits"`
	Requests *Requests `json:"requests"`
}
