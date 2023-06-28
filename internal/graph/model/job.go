package model

type Job struct {
	ID                         Ident             `json:"id"`
	Name                       string            `json:"name"`
	Image                      string            `json:"image"`
	DeployInfo                 DeployInfo        `json:"deployInfo"`
	Env                        *Env              `json:"env"`
	AccessPolicy               *AccessPolicy     `json:"accessPolicy"`
	Resources                  Resources         `json:"resources"`
	Schedule                   string            `json:"schedule"`
	ConcurrencyPolicy          ConcurrencyPolicy `json:"concurrencyPolicy"`
	ActiveDeadlineSeconds      int               `json:"activeDeadlineSeconds"`
	BackoffLimit               int               `json:"backoffLimit"`
	FailedJobsHistoryLimit     int               `json:"failedJobsHistoryLimit"`
	SuccessfulJobsHistoryLimit int               `json:"successfulJobsHistoryLimit"`

	GQLVars struct {
		Team string
	} `json:"-"`
}

func (Job) IsSearchNode()  {}
func (Job) IsNode()        {}
func (j Job) GetID() Ident { return j.ID }

type JobConnection struct {
	Edges      []*JobEdge `json:"edges"`
	PageInfo   *PageInfo  `json:"pageInfo"`
	TotalCount int        `json:"totalCount"`
}

type JobEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *Job   `json:"node"`
}
