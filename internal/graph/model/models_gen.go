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

type DeploymentResponse interface {
	IsDeploymentResponse()
}

type Node interface {
	IsNode()
	GetID() Ident
}

type SearchNode interface {
	IsSearchNode()
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

type AppEdge struct {
	Cursor Cursor `json:"cursor"`
	Node   *App   `json:"node"`
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

type Database struct {
	EnvVarPrefix string          `json:"envVarPrefix"`
	Name         string          `json:"name"`
	Users        []*DatabaseUser `json:"users"`
}

type DatabaseUser struct {
	Name string `json:"name"`
}

type DeploymentKey struct {
	ID      Ident     `json:"id"`
	Key     string    `json:"key"`
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"`
}

func (DeploymentKey) IsNode()           {}
func (this DeploymentKey) GetID() Ident { return this.ID }

type Env struct {
	ID   Ident  `json:"id"`
	Name string `json:"name"`
}

func (Env) IsNode()           {}
func (this Env) GetID() Ident { return this.ID }

type Expose struct {
	AllowedIntegrations []string    `json:"allowedIntegrations"`
	AtMaxAge            int         `json:"atMaxAge"`
	Consumers           []*Consumer `json:"consumers"`
	Enabled             bool        `json:"enabled"`
	Name                string      `json:"name"`
	Product             string      `json:"product"`
}

type Flag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GithubRepository struct {
	Name string `json:"name"`
}

type GithubRepositoryConnection struct {
	TotalCount int                     `json:"totalCount"`
	PageInfo   *PageInfo               `json:"pageInfo"`
	Edges      []*GithubRepositoryEdge `json:"edges"`
}

type GithubRepositoryEdge struct {
	Cursor Cursor            `json:"cursor"`
	Node   *GithubRepository `json:"node"`
}

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

type Insights struct {
	Enabled               bool `json:"enabled"`
	QueryStringLength     int  `json:"queryStringLength"`
	RecordApplicationTags bool `json:"recordApplicationTags"`
	RecordClientAddress   bool `json:"recordClientAddress"`
}

type Kafka struct {
	// The kafka pool name
	Name    string   `json:"name"`
	Streams bool     `json:"streams"`
	Topics  []*Topic `json:"topics"`
}

func (Kafka) IsStorage()           {}
func (this Kafka) GetName() string { return this.Name }

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

type OpenSearch struct {
	// The opensearch instance name
	Name   string `json:"name"`
	Access string `json:"access"`
}

func (OpenSearch) IsStorage()           {}
func (this OpenSearch) GetName() string { return this.Name }

type Requests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type Resources struct {
	Limits   *Limits   `json:"limits"`
	Requests *Requests `json:"requests"`
}

type SearchConnection struct {
	Edges      []*SearchEdge `json:"edges"`
	PageInfo   *PageInfo     `json:"pageInfo"`
	TotalCount int           `json:"totalCount"`
}

type SearchEdge struct {
	Node   SearchNode `json:"node"`
	Cursor Cursor     `json:"cursor"`
	Rank   int        `json:"-"`
}

type SearchFilter struct {
	Type *SearchType `json:"type,omitempty"`
}

type Sidecar struct {
	AutoLogin            bool       `json:"autoLogin"`
	AutoLoginIgnorePaths []string   `json:"autoLoginIgnorePaths"`
	Resources            *Resources `json:"resources"`
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

type TokenX struct {
	MountSecretsAsFilesOnly bool `json:"mountSecretsAsFilesOnly"`
}

func (TokenX) IsAuthz() {}

type Topic struct {
	Name string `json:"name"`
	ACL  []*ACL `json:"acl"`
}

type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AppState string

const (
	AppStateNais    AppState = "NAIS"
	AppStateNotnais AppState = "NOTNAIS"
	AppStateFailing AppState = "FAILING"
	AppStateUnknown AppState = "UNKNOWN"
)

var AllAppState = []AppState{
	AppStateNais,
	AppStateNotnais,
	AppStateFailing,
	AppStateUnknown,
}

func (e AppState) IsValid() bool {
	switch e {
	case AppStateNais, AppStateNotnais, AppStateFailing, AppStateUnknown:
		return true
	}
	return false
}

func (e AppState) String() string {
	return string(e)
}

func (e *AppState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AppState(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AppState", str)
	}
	return nil
}

func (e AppState) MarshalGQL(w io.Writer) {
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

type TeamRole string

const (
	TeamRoleMember TeamRole = "MEMBER"
	TeamRoleOwner  TeamRole = "OWNER"
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
