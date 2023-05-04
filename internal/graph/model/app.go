package model

import "time"

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
}

type Limits struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type AccessPolicy struct {
	Inbound struct {
		Rules []AccessPolicyRule `json:"rules"`
	} `json:"inbound"`
	Outbound struct {
		External []External         `json:"external"`
		Rules    []AccessPolicyRule `json:"rules"`
	} `json:"outbound"`
}

type Replicas struct {
	MinReplicas            int  `json:"min"`
	MaxReplicas            int  `json:"max"`
	DisableAutoScaling     bool `json:"disableAutoScaling"`
	CPUThresholdPercentage int  `json:"cpuThresholdPercentage"`
}

type App struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Image        string       `json:"image"`
	Env          *Env         `json:"env"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
	Ingresses    []string     `json:"ingresses"`
	Resources    Resources    `json:"resources"`
	Deployed     time.Time    `json:"deployed"`
	Replicas     Replicas     `json:"replicas"`
	GQLVars      struct {
		Team string
	} `json:"-"`
}

func (App) IsNode()         {}
func (a App) GetID() string { return a.ID }

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
