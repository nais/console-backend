package model

import "time"

type (
	AppGQLVars struct {
		Team string
	}

	DeployInfoGQLVars struct {
		App  string
		Job  string
		Env  string
		Team string
	}

	InstanceGQLVars struct {
		Env     string
		Team    string
		AppName string
	}

	NaisJobGQLVars struct {
		Team string
	}

	RunGQLVars struct {
		Env     string
		Team    string
		NaisJob string
	}

	ResourceUtilizationForEnvGQLVars struct {
		Start time.Time
		End   time.Time
		Team  string
	}
)
