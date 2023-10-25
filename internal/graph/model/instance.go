package model

import (
	"context"
	"fmt"
	"io"
	"strconv"
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

type InstanceState string

const (
	InstanceStateFailing InstanceState = "FAILING"
	InstanceStateUnknown InstanceState = "UNKNOWN"
	InstanceStateRunning InstanceState = "RUNNING"
)

func (e InstanceState) IsValid() bool {
	switch e {
	case InstanceStateRunning, InstanceStateFailing, InstanceStateUnknown:
		return true
	}
	return false
}

func (e InstanceState) String() string {
	return string(e)
}

func (e *InstanceState) UnmarshalGQLContext(_ context.Context, v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("instance state must be a string")
	}

	*e = InstanceState(str)
	if !e.IsValid() {
		return fmt.Errorf("%q is not a valid InstanceState", str)
	}
	return nil
}

func (e InstanceState) MarshalGQLContext(_ context.Context, w io.Writer) error {
	_, err := fmt.Fprint(w, strconv.Quote(e.String()))
	return err
}
