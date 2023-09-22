package model

import "time"

type Run struct {
	ID             Ident      `json:"id"`
	Name           string     `json:"name"`
	PodNames       []string   `json:"podNames"`
	StartTime      *time.Time `json:"startTime,omitempty"`
	CompletionTime *time.Time `json:"completionTime,omitempty"`
	Duration       string     `json:"duration"`
	Image          string     `json:"image"`
	Message        string     `json:"message"`
	Failed         bool       `json:"failed,omitempty"`
	GQLVars        struct {
		Env     string
		Team    string
		NaisJob string
	} `json:"-"`
}

func (Run) IsNode()        {}
func (r Run) GetID() Ident { return r.ID }

type NaisJob struct {
	ID           Ident         `json:"id"`
	Name         string        `json:"name"`
	Env          *Env          `json:"env"`
	DeployInfo   *DeployInfo   `json:"deployInfo"`
	Image        string        `json:"image"`
	AccessPolicy *AccessPolicy `json:"accessPolicy"`
	Resources    *Resources    `json:"resources"`
	Storage      []Storage     `json:"storage"`
	Authz        []Authz       `json:"authz"`
	Schedule     string        `json:"schedule"`
	Completions  int           `json:"completions"`
	Parallelism  int           `json:"parallelism"`
	Retries      int           `json:"retries"`
	JobState     JobState      `json:"jobState"`
	GQLVars      struct {
		Team string
	} `json:"-"`
}

func (NaisJob) IsNode()        {}
func (j NaisJob) GetID() Ident { return j.ID }

func (NaisJob) IsSearchNode() {}
