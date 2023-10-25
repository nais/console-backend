package model

type Team struct {
	ID                  Ident                `json:"id"`
	Name                string               `json:"name"`
	Description         string               `json:"description,omitempty"`
	SlackChannel        string               `json:"slackChannel"`
	SlackAlertsChannels []SlackAlertsChannel `json:"slackAlertsChannels"`
	GcpProjects         []GcpProject         `json:"gcpProject"`
}

func (Team) IsSearchNode()  {}
func (Team) IsNode()        {}
func (t Team) GetID() Ident { return t.ID }
