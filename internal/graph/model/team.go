package model

type Team struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	//	Members     *TeamMemberConnection `json:"members"`
	Apps *AppConnection `json:"apps"`
}

func (Team) IsNode()            {}
func (this Team) GetID() string { return this.ID }

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

func (TeamMember) IsNode()            {}
func (this TeamMember) GetID() string { return this.ID }

type TeamMemberConnection struct {
	TotalCount int               `json:"totalCount"`
	PageInfo   *PageInfo         `json:"pageInfo"`
	Edges      []*TeamMemberEdge `json:"edges"`
}

type TeamMemberEdge struct {
	Cursor Cursor      `json:"cursor"`
	Node   *TeamMember `json:"node"`
}
