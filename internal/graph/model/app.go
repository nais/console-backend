package model

type AccessPolicyRule struct {
	Application string `json:"application"`
	Namespace   string `json:"namespace"`
}

type External struct {
	Host string `json:"host"`
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
