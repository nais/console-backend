// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type Authz interface {
	IsAuthz()
}

// Connection interface.
type Connection interface {
	IsConnection()
	// The total count of items in the connection.
	GetTotalCount() int
	// Pagination information.
	GetPageInfo() *PageInfo
	// A list of edges.
	GetEdges() []Edge
}

type DeploymentResponse interface {
	IsDeploymentResponse()
}

// Edge interface.
type Edge interface {
	IsEdge()
	// A cursor for use in pagination.
	GetCursor() Cursor
}

// Node interface.
type Node interface {
	IsNode()
	// The unique ID of an object.
	GetID() Ident
}

type SearchNode interface {
	IsSearchNode()
}

type StateError interface {
	IsStateError()
	GetRevision() string
	GetLevel() ErrorLevel
}

type Storage interface {
	IsStorage()
	GetName() string
}

type ACL struct {
	Access      string `json:"access"`
	Application string `json:"application"`
	Team        string `json:"team"`
}

type AppConnection struct {
	TotalCount int        `json:"totalCount"`
	PageInfo   *PageInfo  `json:"pageInfo"`
	Edges      []*AppEdge `json:"edges"`
}

func (AppConnection) IsConnection() {}

// The total count of items in the connection.
func (this AppConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this AppConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this AppConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

// App cost type.
type AppCost struct {
	// The name of the application.
	App string `json:"app"`
	// The sum of all cost entries for the application in euros.
	Sum float64 `json:"sum"`
	// A list of cost entries for the application.
	Cost []*CostEntry `json:"cost"`
}

type AppEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *App   `json:"node"`
}

func (AppEdge) IsEdge() {}

// A cursor for use in pagination.
func (this AppEdge) GetCursor() Cursor { return this.Cursor }

type AppState struct {
	State  State        `json:"state"`
	Errors []StateError `json:"errors"`
}

type AutoScaling struct {
	Disabled bool `json:"disabled"`
	// CPU threshold in percent
	CPUThreshold int `json:"cpuThreshold"`
	Max          int `json:"max"`
	Min          int `json:"min"`
}

type AzureAd struct {
	Application *AzureApplication `json:"application,omitempty"`
	Sidecar     *Sidecar          `json:"sidecar,omitempty"`
}

func (AzureAd) IsAuthz() {}

type AzureApplication struct {
	AllowAllUsers         bool     `json:"allowAllUsers"`
	Claims                *Claims  `json:"claims"`
	ReplyURLs             []string `json:"replyURLs"`
	SinglePageApplication bool     `json:"singlePageApplication"`
	Tenant                string   `json:"tenant"`
}

type BigQueryDataset struct {
	CascadingDelete bool   `json:"cascadingDelete"`
	Description     string `json:"description"`
	Name            string `json:"name"`
	Permission      string `json:"permission"`
}

func (BigQueryDataset) IsStorage()           {}
func (this BigQueryDataset) GetName() string { return this.Name }

type Bucket struct {
	CascadingDelete          bool   `json:"cascadingDelete"`
	Name                     string `json:"name"`
	PublicAccessPrevention   bool   `json:"publicAccessPrevention"`
	RetentionPeriodDays      int    `json:"retentionPeriodDays"`
	UniformBucketLevelAccess bool   `json:"uniformBucketLevelAccess"`
}

func (Bucket) IsStorage()           {}
func (this Bucket) GetName() string { return this.Name }

type Claims struct {
	Extra  []string `json:"extra"`
	Groups []*Group `json:"groups"`
}

type Consume struct {
	Name string `json:"name"`
}

type Consumer struct {
	Name  string `json:"name"`
	Orgno string `json:"orgno"`
}

// Cost entry type.
type CostEntry struct {
	// The date for the entry.
	Date Date `json:"date"`
	// The cost in euros.
	Cost float64 `json:"cost"`
}

// Cost series type.
type CostSeries struct {
	// The type of cost.
	CostType string `json:"costType"`
	// The sum of all daily costs in the series for this cost type in euros.
	Sum float64 `json:"sum"`
	// The cost data.
	Data []*CostEntry `json:"data"`
}

// Daily cost type.
type DailyCost struct {
	// The sum of all costs in the cost series in euros.
	Sum float64 `json:"sum"`
	// The cost series.
	Series []*CostSeries `json:"series"`
}

type Database struct {
	EnvVarPrefix string          `json:"envVarPrefix"`
	Name         string          `json:"name"`
	Users        []*DatabaseUser `json:"users"`
}

type DatabaseUser struct {
	Name string `json:"name"`
}

type DeployInfo struct {
	Deployer  string             `json:"deployer"`
	Timestamp *time.Time         `json:"timestamp,omitempty"`
	CommitSha string             `json:"commitSha"`
	URL       string             `json:"url"`
	History   DeploymentResponse `json:"history"`
	GQLVars   DeployInfoGQLVars  `json:"-"`
}

type Deployment struct {
	ID         Ident                 `json:"id"`
	Team       *Team                 `json:"team"`
	Resources  []*DeploymentResource `json:"resources"`
	Env        string                `json:"env"`
	Statuses   []*DeploymentStatus   `json:"statuses"`
	Created    time.Time             `json:"created"`
	Repository string                `json:"repository"`
}

type DeploymentConnection struct {
	TotalCount int               `json:"totalCount"`
	PageInfo   *PageInfo         `json:"pageInfo"`
	Edges      []*DeploymentEdge `json:"edges"`
}

func (DeploymentConnection) IsConnection() {}

// The total count of items in the connection.
func (this DeploymentConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this DeploymentConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this DeploymentConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (DeploymentConnection) IsDeploymentResponse() {}

type DeploymentEdge struct {
	Cursor Cursor      `json:"cursor"`
	Node   *Deployment `json:"node"`
}

func (DeploymentEdge) IsEdge() {}

// A cursor for use in pagination.
func (this DeploymentEdge) GetCursor() Cursor { return this.Cursor }

// Deployment key type.
type DeploymentKey struct {
	// The unique identifier of the deployment key.
	ID Ident `json:"id"`
	// The actual key.
	Key string `json:"key"`
	// The date the deployment key was created.
	Created time.Time `json:"created"`
	// The date the deployment key expires.
	Expires time.Time `json:"expires"`
}

func (DeploymentKey) IsNode() {}

// The unique ID of an object.
func (this DeploymentKey) GetID() Ident { return this.ID }

type DeploymentResource struct {
	ID        Ident  `json:"id"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
}

type DeploymentStatus struct {
	ID      Ident     `json:"id"`
	Status  string    `json:"status"`
	Message *string   `json:"message,omitempty"`
	Created time.Time `json:"created"`
}

type DeprecatedIngressError struct {
	Revision string     `json:"revision"`
	Level    ErrorLevel `json:"level"`
	Ingress  string     `json:"ingress"`
}

func (DeprecatedIngressError) IsStateError()             {}
func (this DeprecatedIngressError) GetRevision() string  { return this.Revision }
func (this DeprecatedIngressError) GetLevel() ErrorLevel { return this.Level }

type DeprecatedRegistryError struct {
	Revision   string     `json:"revision"`
	Level      ErrorLevel `json:"level"`
	Registry   string     `json:"registry"`
	Repository string     `json:"repository"`
	Name       string     `json:"name"`
	Tag        string     `json:"tag"`
}

func (DeprecatedRegistryError) IsStateError()             {}
func (this DeprecatedRegistryError) GetRevision() string  { return this.Revision }
func (this DeprecatedRegistryError) GetLevel() ErrorLevel { return this.Level }

type Env struct {
	ID   Ident  `json:"id"`
	Name string `json:"name"`
}

func (Env) IsNode() {}

// The unique ID of an object.
func (this Env) GetID() Ident { return this.ID }

// Env cost type.
type EnvCost struct {
	// The name of the environment.
	Env string `json:"env"`
	// The sum of all app costs for the environment in euros.
	Sum float64 `json:"sum"`
	// A list of app costs in the environment.
	Apps []*AppCost `json:"apps"`
}

// Env cost filter input type.
type EnvCostFilter struct {
	// Start date for the cost series, inclusive.
	From Date `json:"from"`
	// End date for cost series, inclusive.
	To Date `json:"to"`
	// The name of the team to get costs for.
	Team string `json:"team"`
}

type Error struct {
	Message string `json:"message"`
}

func (Error) IsDeploymentResponse() {}

type Expose struct {
	AllowedIntegrations []string    `json:"allowedIntegrations"`
	AtMaxAge            int         `json:"atMaxAge"`
	Consumers           []*Consumer `json:"consumers"`
	Enabled             bool        `json:"enabled"`
	Name                string      `json:"name"`
	Product             string      `json:"product"`
}

type External struct {
	Host  string  `json:"host"`
	Ports []*Port `json:"ports"`
}

type FailedRunError struct {
	Revision   string     `json:"revision"`
	Level      ErrorLevel `json:"level"`
	RunMessage string     `json:"runMessage"`
	RunName    string     `json:"runName"`
}

func (FailedRunError) IsStateError()             {}
func (this FailedRunError) GetRevision() string  { return this.Revision }
func (this FailedRunError) GetLevel() ErrorLevel { return this.Level }

type Flag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GCP project type.
type GcpProject struct {
	// The unique identifier of the GCP project.
	ID string `json:"id"`
	// The name of the GCP project.
	Name string `json:"name"`
	// The environment for the GCP project.
	Environment string `json:"environment"`
}

// GitHub repository type.
type GithubRepository struct {
	// The name of the GitHub repository.
	Name string `json:"name"`
}

// GitHub repository connection type.
type GithubRepositoryConnection struct {
	// The total count of available GitHub repositories.
	TotalCount int `json:"totalCount"`
	// Pagination information.
	PageInfo *PageInfo `json:"pageInfo"`
	// A list of GitHub repository edges.
	Edges []*GithubRepositoryEdge `json:"edges"`
}

func (GithubRepositoryConnection) IsConnection() {}

// The total count of items in the connection.
func (this GithubRepositoryConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this GithubRepositoryConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this GithubRepositoryConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

// GitHub repository edge type.
type GithubRepositoryEdge struct {
	// A cursor for use in pagination.
	Cursor Cursor `json:"cursor"`
	// The GitHub repository at the end of the edge.
	Node *GithubRepository `json:"node"`
}

func (GithubRepositoryEdge) IsEdge() {}

// A cursor for use in pagination.
func (this GithubRepositoryEdge) GetCursor() Cursor { return this.Cursor }

type Group struct {
	ID string `json:"id"`
}

type IDPorten struct {
	AccessTokenLifetime    *int             `json:"accessTokenLifetime,omitempty"`
	ClientURI              *string          `json:"clientURI,omitempty"`
	FrontchannelLogoutPath *string          `json:"frontchannelLogoutPath,omitempty"`
	IntegrationType        *string          `json:"integrationType,omitempty"`
	PostLogoutRedirectURIs []*string        `json:"postLogoutRedirectURIs,omitempty"`
	RedirectPath           *string          `json:"redirectPath,omitempty"`
	Scopes                 []*string        `json:"scopes,omitempty"`
	SessionLifetime        *int             `json:"sessionLifetime,omitempty"`
	Sidecar                *IDPortenSidecar `json:"sidecar,omitempty"`
}

func (IDPorten) IsAuthz() {}

type IDPortenSidecar struct {
	AutoLogin            *bool      `json:"autoLogin,omitempty"`
	AutoLoginIgnorePaths []*string  `json:"autoLoginIgnorePaths,omitempty"`
	Enabled              *bool      `json:"enabled,omitempty"`
	Level                *string    `json:"level,omitempty"`
	Locale               *string    `json:"locale,omitempty"`
	Resources            *Resources `json:"resources,omitempty"`
}

type Inbound struct {
	Rules []*Rule `json:"rules"`
}

type InboundAccessError struct {
	Revision string     `json:"revision"`
	Level    ErrorLevel `json:"level"`
	Rule     *Rule      `json:"rule"`
}

func (InboundAccessError) IsStateError()             {}
func (this InboundAccessError) GetRevision() string  { return this.Revision }
func (this InboundAccessError) GetLevel() ErrorLevel { return this.Level }

type InfluxDb struct {
	Name string `json:"name"`
}

func (InfluxDb) IsStorage()           {}
func (this InfluxDb) GetName() string { return this.Name }

type Insights struct {
	Enabled               bool `json:"enabled"`
	QueryStringLength     int  `json:"queryStringLength"`
	RecordApplicationTags bool `json:"recordApplicationTags"`
	RecordClientAddress   bool `json:"recordClientAddress"`
}

type InvalidNaisYamlError struct {
	Revision string     `json:"revision"`
	Level    ErrorLevel `json:"level"`
	Detail   string     `json:"detail"`
}

func (InvalidNaisYamlError) IsStateError()             {}
func (this InvalidNaisYamlError) GetRevision() string  { return this.Revision }
func (this InvalidNaisYamlError) GetLevel() ErrorLevel { return this.Level }

type JobState struct {
	State  State        `json:"state"`
	Errors []StateError `json:"errors"`
}

type Kafka struct {
	// The kafka pool name
	Name    string   `json:"name"`
	Streams bool     `json:"streams"`
	Topics  []*Topic `json:"topics"`
}

func (Kafka) IsStorage()           {}
func (this Kafka) GetName() string { return this.Name }

type Limits struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type LogLine struct {
	Time     time.Time `json:"time"`
	Message  string    `json:"message"`
	Instance string    `json:"instance"`
}

type LogSubscriptionInput struct {
	App       *string  `json:"app,omitempty"`
	Job       *string  `json:"job,omitempty"`
	Env       string   `json:"env"`
	Team      string   `json:"team"`
	Instances []string `json:"instances,omitempty"`
}

type Maintenance struct {
	Day  int `json:"day"`
	Hour int `json:"hour"`
}

type Maskinporten struct {
	Scopes  *MaskinportenScope `json:"scopes"`
	Enabled bool               `json:"enabled"`
}

func (Maskinporten) IsAuthz() {}

type MaskinportenScope struct {
	Consumes []*Consume `json:"consumes"`
	Exposes  []*Expose  `json:"exposes"`
}

// Montly cost type.
type MonthlyCost struct {
	// Sum for all months in the series in euros.
	Sum float64 `json:"sum"`
	// A list of monthly cost entries.
	Cost []*CostEntry `json:"cost"`
}

// Monthly cost filter input type.
type MonthlyCostFilter struct {
	// The name of the team to get costs for.
	Team string `json:"team"`
	// The name of the application to get costs for.
	App string `json:"app"`
	// The name of the environment to get costs for.
	Env string `json:"env"`
}

type NaisJobConnection struct {
	TotalCount int            `json:"totalCount"`
	PageInfo   *PageInfo      `json:"pageInfo"`
	Edges      []*NaisJobEdge `json:"edges"`
}

func (NaisJobConnection) IsConnection() {}

// The total count of items in the connection.
func (this NaisJobConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this NaisJobConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this NaisJobConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

type NaisJobEdge struct {
	Cursor Cursor   `json:"cursor"`
	Node   *NaisJob `json:"node"`
}

func (NaisJobEdge) IsEdge() {}

// A cursor for use in pagination.
func (this NaisJobEdge) GetCursor() Cursor { return this.Cursor }

type NewInstancesFailingError struct {
	Revision         string     `json:"revision"`
	Level            ErrorLevel `json:"level"`
	FailingInstances []string   `json:"failingInstances"`
}

func (NewInstancesFailingError) IsStateError()             {}
func (this NewInstancesFailingError) GetRevision() string  { return this.Revision }
func (this NewInstancesFailingError) GetLevel() ErrorLevel { return this.Level }

type NoRunningInstancesError struct {
	Revision string     `json:"revision"`
	Level    ErrorLevel `json:"level"`
}

func (NoRunningInstancesError) IsStateError()             {}
func (this NoRunningInstancesError) GetRevision() string  { return this.Revision }
func (this NoRunningInstancesError) GetLevel() ErrorLevel { return this.Level }

type OpenSearch struct {
	// The opensearch instance name
	Name   string `json:"name"`
	Access string `json:"access"`
}

func (OpenSearch) IsStorage()           {}
func (this OpenSearch) GetName() string { return this.Name }

type Outbound struct {
	Rules    []*Rule     `json:"rules"`
	External []*External `json:"external"`
}

type OutboundAccessError struct {
	Revision string     `json:"revision"`
	Level    ErrorLevel `json:"level"`
	Rule     *Rule      `json:"rule"`
}

func (OutboundAccessError) IsStateError()             {}
func (this OutboundAccessError) GetRevision() string  { return this.Revision }
func (this OutboundAccessError) GetLevel() ErrorLevel { return this.Level }

// PageInfo is a type that contains pagination information in a Relay style.
type PageInfo struct {
	// When paginating forwards, are there more items?
	HasNextPage bool `json:"hasNextPage"`
	// When paginating backwards, are there more items?
	HasPreviousPage bool `json:"hasPreviousPage"`
	// A cursor corresponding to the first node in the connection.
	StartCursor *Cursor `json:"startCursor,omitempty"`
	// A cursor corresponding to the last node in the connection.
	EndCursor *Cursor `json:"endCursor,omitempty"`
	From      int     `json:"from"`
	To        int     `json:"to"`
}

type Port struct {
	Port int `json:"port"`
}

type Redis struct {
	Name   string `json:"name"`
	Access string `json:"access"`
}

func (Redis) IsStorage()           {}
func (this Redis) GetName() string { return this.Name }

type Requests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Resources struct {
	Limits   *Limits   `json:"limits"`
	Requests *Requests `json:"requests"`
}

type Rule struct {
	Application       string `json:"application"`
	Namespace         string `json:"namespace"`
	Cluster           string `json:"cluster"`
	Mutual            bool   `json:"mutual"`
	MutualExplanation string `json:"mutualExplanation"`
}

type SearchConnection struct {
	TotalCount int           `json:"totalCount"`
	PageInfo   *PageInfo     `json:"pageInfo"`
	Edges      []*SearchEdge `json:"edges"`
}

func (SearchConnection) IsConnection() {}

// The total count of items in the connection.
func (this SearchConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this SearchConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this SearchConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

type SearchEdge struct {
	Node   SearchNode `json:"node"`
	Cursor Cursor     `json:"cursor"`
	Rank   int        `json:"-"`
}

func (SearchEdge) IsEdge() {}

// A cursor for use in pagination.
func (this SearchEdge) GetCursor() Cursor { return this.Cursor }

type SearchFilter struct {
	Type *SearchType `json:"type,omitempty"`
}

type Sidecar struct {
	AutoLogin            bool       `json:"autoLogin"`
	AutoLoginIgnorePaths []string   `json:"autoLoginIgnorePaths"`
	Resources            *Resources `json:"resources"`
}

// Slack alerts channel type.
type SlackAlertsChannel struct {
	// The name of the Slack alerts channel.
	Name string `json:"name"`
	// The environment for the Slack alerts channel.
	Env string `json:"env"`
}

type SQLInstance struct {
	AutoBackupHour      int          `json:"autoBackupHour"`
	CascadingDelete     bool         `json:"cascadingDelete"`
	Collation           string       `json:"collation"`
	Databases           []*Database  `json:"databases"`
	DiskAutoresize      bool         `json:"diskAutoresize"`
	DiskSize            int          `json:"diskSize"`
	DiskType            string       `json:"diskType"`
	Flags               []*Flag      `json:"flags"`
	HighAvailability    bool         `json:"highAvailability"`
	Insights            *Insights    `json:"insights"`
	Maintenance         *Maintenance `json:"maintenance"`
	Name                string       `json:"name"`
	PointInTimeRecovery bool         `json:"pointInTimeRecovery"`
	RetainedBackups     int          `json:"retainedBackups"`
	Tier                string       `json:"tier"`
	Type                string       `json:"type"`
}

func (SQLInstance) IsStorage()           {}
func (this SQLInstance) GetName() string { return this.Name }

// Team type.
type Team struct {
	// The unique identifier of the team.
	ID Ident `json:"id"`
	// The name of the team.
	Name string `json:"name"`
	// The description of the team.
	Description string `json:"description"`
	// Team members.
	Members *TeamMemberConnection `json:"members"`
	// The NAIS applications owned by the team.
	Apps *AppConnection `json:"apps"`
	// The NAIS jobs owned by the team.
	Naisjobs *NaisJobConnection `json:"naisjobs"`
	// The GitHub repositories that the team has access to.
	GithubRepositories *GithubRepositoryConnection `json:"githubRepositories"`
	// The main Slack channel for the team.
	SlackChannel string `json:"slackChannel"`
	// Slack alerts channels for the team.
	SlackAlertsChannels []*SlackAlertsChannel `json:"slackAlertsChannels"`
	GcpProjects         []*GcpProject         `json:"gcpProjects"`
	// The deployments of the team's applications.
	Deployments *DeploymentConnection `json:"deployments"`
	// The deploy key of the team.
	DeployKey *DeploymentKey `json:"deployKey"`
	// Whether or not the viewer is a member of the team.
	ViewerIsMember bool `json:"viewerIsMember"`
	// Whether or not the viewer is an administrator of the team.
	ViewerIsAdmin bool `json:"viewerIsAdmin"`
}

func (Team) IsSearchNode() {}

func (Team) IsNode() {}

// The unique ID of an object.
func (this Team) GetID() Ident { return this.ID }

// Team connection type.
type TeamConnection struct {
	// The total count of available teams.
	TotalCount int `json:"totalCount"`
	// Pagination information.
	PageInfo *PageInfo `json:"pageInfo"`
	// A list of team edges.
	Edges []*TeamEdge `json:"edges"`
}

func (TeamConnection) IsConnection() {}

// The total count of items in the connection.
func (this TeamConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this TeamConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this TeamConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

// Team edge type.
type TeamEdge struct {
	// A cursor for use in pagination.
	Cursor Cursor `json:"cursor"`
	// The team at the end of the edge.
	Node *Team `json:"node"`
}

func (TeamEdge) IsEdge() {}

// A cursor for use in pagination.
func (this TeamEdge) GetCursor() Cursor { return this.Cursor }

// Team member type.
type TeamMember struct {
	// The unique identifier of the team member.
	ID Ident `json:"id"`
	// The name of the team member.
	Name string `json:"name"`
	// The email of the team member.
	Email string `json:"email"`
	// The role of the team member.
	Role TeamRole `json:"role"`
}

func (TeamMember) IsNode() {}

// The unique ID of an object.
func (this TeamMember) GetID() Ident { return this.ID }

// Team member connection type.
type TeamMemberConnection struct {
	// The total count of available team members.
	TotalCount int `json:"totalCount"`
	// Pagination information.
	PageInfo *PageInfo `json:"pageInfo"`
	// A list of team member edges.
	Edges []*TeamMemberEdge `json:"edges"`
}

func (TeamMemberConnection) IsConnection() {}

// The total count of items in the connection.
func (this TeamMemberConnection) GetTotalCount() int { return this.TotalCount }

// Pagination information.
func (this TeamMemberConnection) GetPageInfo() *PageInfo { return this.PageInfo }

// A list of edges.
func (this TeamMemberConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

// Team member edge type.
type TeamMemberEdge struct {
	// A cursor for use in pagination.
	Cursor Cursor `json:"cursor"`
	// The team member at the end of the edge.
	Node *TeamMember `json:"node"`
}

func (TeamMemberEdge) IsEdge() {}

// A cursor for use in pagination.
func (this TeamMemberEdge) GetCursor() Cursor { return this.Cursor }

type TokenX struct {
	MountSecretsAsFilesOnly bool `json:"mountSecretsAsFilesOnly"`
}

func (TokenX) IsAuthz() {}

type Topic struct {
	Name string `json:"name"`
	ACL  []*ACL `json:"acl"`
}

type User struct {
	// The unique identifier for the user.
	ID Ident `json:"id"`
	// The user's full name.
	Name string `json:"name"`
	// The user's email address.
	Email string `json:"email"`
	// Teams that the user is a member and/or owner of.
	Teams *TeamConnection `json:"teams"`
}

func (User) IsNode() {}

// The unique ID of an object.
func (this User) GetID() Ident { return this.ID }

type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ErrorLevel string

const (
	ErrorLevelInfo    ErrorLevel = "INFO"
	ErrorLevelWarning ErrorLevel = "WARNING"
	ErrorLevelError   ErrorLevel = "ERROR"
)

var AllErrorLevel = []ErrorLevel{
	ErrorLevelInfo,
	ErrorLevelWarning,
	ErrorLevelError,
}

func (e ErrorLevel) IsValid() bool {
	switch e {
	case ErrorLevelInfo, ErrorLevelWarning, ErrorLevelError:
		return true
	}
	return false
}

func (e ErrorLevel) String() string {
	return string(e)
}

func (e *ErrorLevel) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ErrorLevel(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ErrorLevel", str)
	}
	return nil
}

func (e ErrorLevel) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type SearchType string

const (
	SearchTypeApp     SearchType = "APP"
	SearchTypeTeam    SearchType = "TEAM"
	SearchTypeNaisjob SearchType = "NAISJOB"
)

var AllSearchType = []SearchType{
	SearchTypeApp,
	SearchTypeTeam,
	SearchTypeNaisjob,
}

func (e SearchType) IsValid() bool {
	switch e {
	case SearchTypeApp, SearchTypeTeam, SearchTypeNaisjob:
		return true
	}
	return false
}

func (e SearchType) String() string {
	return string(e)
}

func (e *SearchType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SearchType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SearchType", str)
	}
	return nil
}

func (e SearchType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type State string

const (
	StateNais    State = "NAIS"
	StateNotnais State = "NOTNAIS"
	StateFailing State = "FAILING"
	StateUnknown State = "UNKNOWN"
)

var AllState = []State{
	StateNais,
	StateNotnais,
	StateFailing,
	StateUnknown,
}

func (e State) IsValid() bool {
	switch e {
	case StateNais, StateNotnais, StateFailing, StateUnknown:
		return true
	}
	return false
}

func (e State) String() string {
	return string(e)
}

func (e *State) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = State(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid State", str)
	}
	return nil
}

func (e State) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

// Team member roles.
type TeamRole string

const (
	// A regular team member.
	TeamRoleMember TeamRole = "MEMBER"
	// A team owner/administrator.
	TeamRoleOwner TeamRole = "OWNER"
)

var AllTeamRole = []TeamRole{
	TeamRoleMember,
	TeamRoleOwner,
}

func (e TeamRole) IsValid() bool {
	switch e {
	case TeamRoleMember, TeamRoleOwner:
		return true
	}
	return false
}

func (e TeamRole) String() string {
	return string(e)
}

func (e *TeamRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TeamRole(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TeamRole", str)
	}
	return nil
}

func (e TeamRole) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
