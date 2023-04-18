// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Node interface {
	IsNode()
	GetID() string
}

type App struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Team *Team  `json:"team"`
	Env  *Env   `json:"env"`
}

func (App) IsNode()            {}
func (this App) GetID() string { return this.ID }

type AppConnection struct {
	PageInfo *PageInfo  `json:"pageInfo"`
	Edges    []*AppEdge `json:"edges"`
}

type AppEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *App   `json:"node"`
}

type Env struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (Env) IsNode()            {}
func (this Env) GetID() string { return this.ID }

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *Cursor `json:"startCursor,omitempty"`
	EndCursor       *Cursor `json:"endCursor,omitempty"`
}

type Team struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Members *TeamMemberConnection `json:"members"`
	Apps    *AppConnection        `json:"apps"`
}

func (Team) IsNode()            {}
func (this Team) GetID() string { return this.ID }

type TeamConnection struct {
	PageInfo *PageInfo   `json:"pageInfo"`
	Edges    []*TeamEdge `json:"edges"`
}

type TeamEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *Team  `json:"node"`
}

type TeamMember struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (TeamMember) IsNode()            {}
func (this TeamMember) GetID() string { return this.ID }

type TeamMemberConnection struct {
	PageInfo *PageInfo         `json:"pageInfo"`
	Edges    []*TeamMemberEdge `json:"edges"`
}

type TeamMemberEdge struct {
	Cursor Cursor      `json:"cursor"`
	Node   *TeamMember `json:"node"`
}
