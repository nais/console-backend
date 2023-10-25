package model

import (
	"time"
)

type InstanceGQLVars struct {
	Env     string
	Team    string
	AppName string
}

type Instance struct {
	ID       Ident           `json:"id"`
	Name     string          `json:"name"`
	Image    string          `json:"image"`
	State    InstanceState   `json:"state"`
	Restarts int             `json:"restarts"`
	Message  string          `json:"message"`
	Created  time.Time       `json:"created"`
	GQLVars  InstanceGQLVars `json:"-"`
}

func (Instance) IsNode()        {}
func (i Instance) GetID() Ident { return i.ID }
