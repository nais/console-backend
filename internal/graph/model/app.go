package model

type AccessPolicyRule struct {
	Application string `json:"application"`
	Namespace   string `json:"namespace"`
}

type External struct {
	Host  string `json:"host"`
	Ports []struct {
		Port int `json:"port"`
	} `json:"ports"`
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

type App struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Image        string       `json:"image"`
	Env          *Env         `json:"env"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
	Ingresses    []string     `json:"ingresses"`
	Resources    Resources    `json:"resources"`
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
