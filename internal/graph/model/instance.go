package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type Instance struct {
	ID       Ident         `json:"id"`
	Name     string        `json:"name"`
	Image    string        `json:"image"`
	State    InstanceState `json:"state"`
	Restarts int           `json:"restarts"`
	Message  string        `json:"message"`
	Created  time.Time     `json:"created"`
	GQLVars  struct {
		Env     string
		Team    string
		AppName string
	} `json:"-"`
}

type InstanceState string

const (
	InstanceStateFailing InstanceState = "FAILING"
	InstanceStateUnknown InstanceState = "UNKNOWN"
	InstanceStateRunning InstanceState = "RUNNING"
)

var AllInstanceState = []InstanceState{
	InstanceStateRunning,
	InstanceStateFailing,
	InstanceStateUnknown,
}

func (Instance) IsNode()        {}
func (i Instance) GetID() Ident { return i.ID }

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

func (e *InstanceState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = InstanceState(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid InstanceState", str)
	}
	return nil
}

func (e InstanceState) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
