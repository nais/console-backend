package model

type Team struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Description         *string              `json:"description,omitempty"`
	Apps                *AppConnection       `json:"apps"`
	SlackChannel        string               `json:"slackChannel"`
	SlackAlertsChannels []SlackAlertsChannel `json:"slackAlertsChannels"`
}

type SlackAlertsChannel struct {
	Env  string `json:"env"`
	Name string `json:"name"`
}

func (Team) IsNode()         {}
func (t Team) GetID() string { return t.ID }

type TeamConnection struct {
	TotalCount int         `json:"totalCount"`
	PageInfo   *PageInfo   `json:"pageInfo"`
	Edges      []*TeamEdge `json:"edges"`
}

type TeamEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *Team  `json:"node"`
}

type TeamMember struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (TeamMember) IsNode()         {}
func (t TeamMember) GetID() string { return t.ID }

type TeamMemberConnection struct {
	TotalCount int               `json:"totalCount"`
	PageInfo   *PageInfo         `json:"pageInfo"`
	Edges      []*TeamMemberEdge `json:"edges"`
}

type TeamMemberEdge struct {
	Cursor Cursor      `json:"cursor"`
	Node   *TeamMember `json:"node"`
}
