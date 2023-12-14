package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	"github.com/nais/console-backend/internal/search"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const teamsCacheTTL = 15 * time.Minute

type User struct {
	Name  string           `json:"name"`
	ID    uuid.UUID        `json:"id"`
	Teams []TeamMembership `json:"teams"`
}

type TeamMembership struct {
	Team Team `json:"team"`
}

type ReconcilerState struct {
	GcpProjects []GcpProject `json:"gcpProjects"`
}

type GcpProject struct {
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	Environment string `json:"environment"`
}

type Team struct {
	Slug                string               `json:"slug"`
	Purpose             string               `json:"purpose"`
	SlackChannel        string               `json:"slackChannel"`
	GitHubRepositories  []GitHubRepository   `json:"gitHubRepositories"`
	SlackAlertsChannels []SlackAlertsChannel `json:"slackAlertsChannels"`
	Members             []Member             `json:"members"`
	ReconcilerState     ReconcilerState      `json:"reconcilerState"`
}

type GitHubRepository struct {
	Name string `json:"name"`
}

type SlackAlertsChannel struct {
	Environment string `json:"environment"`
	ChannelName string `json:"channelName"`
}

type Member struct {
	Role string   `json:"role"`
	User TeamUser `json:"user"`
}

type TeamUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Client interface {
	Search(ctx context.Context, query string, filter *model.SearchFilter) []*search.Result
	GetTeam(ctx context.Context, teamSlug string) (*model.Team, error)
	GetGithubRepositories(ctx context.Context, teamSlug string) ([]GitHubRepository, error)
	GetTeamMembers(ctx context.Context, teamSlug string) ([]Member, error)
	GetTeams(ctx context.Context) ([]Team, error)
	GetTeamsForUser(ctx context.Context, email string) ([]TeamMembership, error)
	GetUser(ctx context.Context, email string) (*User, error)
	GetSelf(ctx context.Context) (*User, error)
	TeamExists(ctx context.Context, teamSlug string) bool
}

type client struct {
	endpoint   string
	httpClient *httpClient
	lock       sync.RWMutex
	teams      []*model.Team
	updated    time.Time
	log        logrus.FieldLogger
	errors     metric.Int64Counter
}

func New(cfg config.Teams, onBehalfOf bool, errors metric.Int64Counter, log logrus.FieldLogger) Client {
	return &client{
		endpoint: cfg.Endpoint,
		httpClient: &httpClient{
			client:     &http.Client{},
			apiToken:   cfg.Token,
			onBehalfOf: onBehalfOf,
		},
		log:    log,
		errors: errors,
	}
}

// TeamExists checks if a team exists on the backend or not
func (c *client) TeamExists(ctx context.Context, teamSlug string) bool {
	c.updateTeams(ctx)
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, team := range c.teams {
		if team.Slug == teamSlug {
			return true
		}
	}
	return false
}

// Search searches for teams matching the query
func (c *client) Search(ctx context.Context, query string, filter *model.SearchFilter) []*search.Result {
	if !isTeamFilterOrNoFilter(filter) {
		return nil
	}

	c.updateTeams(ctx)
	c.lock.RLock()
	defer c.lock.RUnlock()

	edges := make([]*search.Result, 0)
	for _, team := range c.teams {
		rank := search.Match(query, team.Slug)
		if rank == -1 {
			continue
		}
		edges = append(edges, &search.Result{
			Rank: rank,
			Node: team,
		})
	}
	return edges
}


// GetTeam get a team by the team slug
func (c *client) GetTeam(ctx context.Context, teamSlug string) (*model.Team, error) {
	c.updateTeams(ctx)
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, team := range c.teams {
		if team.Slug == teamSlug {
			return team, nil
		}
	}
	return nil, fmt.Errorf("team not found: %s", teamSlug)
}

// GetGithubRepositories get a list of GitHub repositories for a specific team
func (c *client) GetGithubRepositories(ctx context.Context, teamSlug string) ([]GitHubRepository, error) {
	query := `query ($slug: Slug!) {
		team(slug: $slug) {
			gitHubRepositories {
				name
			}
		}
	}`

	vars := map[string]any{
		"slug": teamSlug,
	}

	respBody := struct {
		Data struct {
			Team *Team `json:"team"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, query, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for github repositories")
	}

	if len(respBody.Errors) > 0 {
		return nil, fmt.Errorf("team not found: %s", teamSlug)
	}

	return respBody.Data.Team.GitHubRepositories, nil
}

// GetTeamMembers get a list of team members for a specific team
func (c *client) GetTeamMembers(ctx context.Context, teamSlug string) ([]Member, error) {
	query := `query ($slug: Slug!) {
		team(slug: $slug) {
			members {
				role
				user {
					email
					name
				}
			}
		}
	}`

	vars := map[string]any{
		"slug": teamSlug,
	}

	respBody := struct {
		Data struct {
			Team *Team `json:"team"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, query, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for members")
	}

	if len(respBody.Errors) > 0 {
		return nil, fmt.Errorf("team not found: %s", teamSlug)
	}

	return respBody.Data.Team.Members, nil
}

func (c *client) GetTeams(ctx context.Context) ([]Team, error) {
	query := `query {
		teams {
			slug
			purpose
			slackChannel
			slackAlertsChannels {
				channelName
				environment
			}
			reconcilerState {
				gcpProjects {
					projectId
					projectName
					environment
				}
			}
		}
	}`

	respBody := struct {
		Data struct {
			Teams []Team `json:"teams"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, query, nil, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for teams")
	}

	return respBody.Data.Teams, nil
}

func (c *client) GetTeamsForUser(ctx context.Context, email string) ([]TeamMembership, error) {
	query := `query ($email: String!) {
		userByEmail(email: $email) {
			teams {
				team {
					slug
					purpose
				}
			}
		}
	}`

	vars := map[string]any{
		"email": email,
	}

	respBody := struct {
		Data struct {
			UserByEmail *User `json:"userByEmail"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, query, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for user teams")
	}

	return respBody.Data.UserByEmail.Teams, nil
}


// GetSelf gets a user by token
func (c *client) GetSelf(ctx context.Context) (*User, error) {
	query := `query {
	
		me {
			... on User {
			name
			id
			}
		}
	}`

	respBody := struct {
		Data struct {
			Self *User `json:"me"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}
	vars := map[string]any{}

	if err := c.teamsQuery(ctx, query, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for self")
	}

	if respBody.Data.Self == nil {
		return nil, fmt.Errorf("user not found")
	}

	return respBody.Data.Self, nil
}

// GetUser get a user by email
func (c *client) GetUser(ctx context.Context, email string) (*User, error) {
	query := `query ($email: String!) {
		userByEmail(email: $email) {
			name
			id
		}
	}`

	vars := map[string]any{
		"email": email,
	}

	respBody := struct {
		Data struct {
			UserByEmail *User `json:"userByEmail"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, query, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for user")
	}

	if respBody.Data.UserByEmail == nil {
		return nil, fmt.Errorf("user %s not found", email)
	}

	return respBody.Data.UserByEmail, nil
}

func (c *client) teamsQuery(ctx context.Context, query string, vars map[string]any, respBody interface{}) error {
	q := struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}{
		Query:     query,
		Variables: vars,
	}

	body, err := json.Marshal(q)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		return fmt.Errorf("teams: %v", resp.Status)
	}
	r := io.TeeReader(resp.Body, os.Stdout)

	if err := json.NewDecoder(r).Decode(&respBody); err != nil {
		return err
	}

	return nil
}

func (c *client) error(ctx context.Context, err error, msg string) error {
	c.errors.Add(ctx, 1, metric.WithAttributes(attribute.String("component", "teams-client")))
	c.log.WithError(err).Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}

// updateTeams update the teams cache when necessary
func (c *client) updateTeams(ctx context.Context) error {
	c.lock.RLock()
	if time.Since(c.updated) < teamsCacheTTL {
		c.lock.RUnlock()
		return nil
	}
	c.lock.RUnlock()
	c.lock.Lock()
	defer c.lock.Unlock()

	teams, err := c.GetTeams(ctx)
	if err != nil {
		return c.error(ctx, err, "get teams from the teams-backend")
	}

	c.teams = toModelTeams(teams)
	c.updated = time.Now()
	return nil
}

// toModelTeams convert a list of teams from the backend to a list of console backend teams
func toModelTeams(teams []Team) []*model.Team {
	models := make([]*model.Team, 0)
	for _, team := range teams {
		models = append(models, &model.Team{
			ID:           scalar.TeamIdent(team.Slug),
			Slug:         team.Slug,
		})
	}
	return models
}

// isTeamFilterOrNoFilter returns true if the filter is a team filter or no filter is provided
func isTeamFilterOrNoFilter(filter *model.SearchFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Type == nil {
		return true
	}

	return *filter.Type == model.SearchTypeTeam
}
