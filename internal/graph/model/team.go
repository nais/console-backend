package model

type Team struct {
	ID                  Ident                `json:"id"`
	Name                string               `json:"name"`
	Description         *string              `json:"description,omitempty"`
	SlackChannel        string               `json:"slackChannel"`
	SlackAlertsChannels []SlackAlertsChannel `json:"slackAlertsChannels"`
}

func (Team) IsSearchNode()  {}
func (Team) IsNode()        {}
func (t Team) GetID() Ident { return t.ID }

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
	ID    Ident    `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Role  TeamRole `json:"role"`
}

func (TeamMember) IsNode()        {}
func (t TeamMember) GetID() Ident { return t.ID }

type TeamMemberConnection struct {
	TotalCount int               `json:"totalCount"`
	PageInfo   *PageInfo         `json:"pageInfo"`
	Edges      []*TeamMemberEdge `json:"edges"`
}

type TeamMemberEdge struct {
	Cursor Cursor      `json:"cursor"`
	Node   *TeamMember `json:"node"`
}
