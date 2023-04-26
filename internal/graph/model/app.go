package model

type App struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Env     *Env   `json:"env"`
	GQLVars struct {
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
