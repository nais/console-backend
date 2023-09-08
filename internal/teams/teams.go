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
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/search"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type User struct {
	Name  string           `json:"name"`
	ID    uuid.UUID        `json:"id"`
	Teams []TeamMembership `json:"teams"`
}

type TeamMembership struct {
	Team Team `json:"team"`
}

type Team struct {
	Slug                string               `json:"slug"`
	Purpose             string               `json:"purpose"`
	SlackChannel        string               `json:"slackChannel"`
	GitHubRepositories  []GitHubRepository   `json:"gitHubRepositories"`
	SlackAlertsChannels []SlackAlertsChannel `json:"slackAlertsChannels"`
	Members             []Member             `json:"members"`
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

type Client struct {
	endpoint   string
	httpClient *http.Client
	lock       sync.RWMutex
	teams      []*model.Team
	updated    time.Time
	log        *logrus.Entry
	errors     metric.Int64Counter
}

func New(token, endpoint string, errors metric.Int64Counter, log *logrus.Entry) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: Transport{Token: token}.Client(),
		log:        log,
		errors:     errors,
	}
}

func (c *Client) Search(ctx context.Context, q string, filter *model.SearchFilter) []*search.SearchResult {
	// early exit if we're not searching for teams
	if filter != nil && filter.Type != nil && *filter.Type != model.SearchTypeTeam {
		return nil
	}

	c.updateTeams(ctx)
	c.lock.RLock()
	defer c.lock.RUnlock()

	edges := []*search.SearchResult{}
	for _, team := range c.teams {
		team := team
		rank := search.Match(q, team.Name)
		if rank == -1 {
			continue
		}
		edges = append(edges, &search.SearchResult{
			Rank: rank,
			Node: team,
		})
	}
	return edges
}

func (c *Client) updateTeams(ctx context.Context) {
	c.lock.RLock()
	if time.Since(c.updated) < 15*time.Minute {
		c.lock.RUnlock()
		return
	}
	c.lock.RUnlock()
	c.lock.Lock()
	defer c.lock.Unlock()

	teams, err := c.GetTeams(ctx)
	if err != nil {
		c.error(ctx, err, "getting teams from teams")
		return
	}

	c.teams = toModelTeams(teams)
	c.updated = time.Now()
}

func toModelTeams(teams []Team) []*model.Team {
	ret := []*model.Team{}
	for _, team := range teams {
		team := team
		ret = append(ret, &model.Team{
			ID:           model.Ident{ID: team.Slug, Type: "team"},
			Name:         team.Slug,
			Description:  &team.Purpose,
			SlackChannel: team.SlackChannel,
			SlackAlertsChannels: func(t []SlackAlertsChannel) []model.SlackAlertsChannel {
				ret := []model.SlackAlertsChannel{}
				for _, v := range t {
					ret = append(ret, model.SlackAlertsChannel{
						Env:  v.Environment,
						Name: v.ChannelName,
					})
				}
				return ret
			}(team.SlackAlertsChannels),
		})
	}
	return ret
}

func (c *Client) GetTeam(ctx context.Context, name string) (*model.Team, error) {
	c.updateTeams(ctx)
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, team := range c.teams {
		team := team
		if team.Name == name {
			return team, nil
		}
	}
	return nil, fmt.Errorf("team not found: %s", name)
}

func (c *Client) GetGithubRepositories(ctx context.Context, name string) ([]GitHubRepository, error) {
	q := `
	query githubRepositories($slug: Slug!) {
	  team(slug: $slug) {
		gitHubRepositories{
			name
		}
	  }
	}`

	vars := map[string]string{
		"slug": name,
	}

	respBody := struct {
		Data struct {
			Team *Team `json:"team"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, q, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for github repositories")
	}

	return respBody.Data.Team.GitHubRepositories, nil
}

func (c *Client) GetMembers(ctx context.Context, name string) ([]Member, error) {
	q := `query teamMembers($slug: Slug!) {
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
	vars := map[string]string{
		"slug": name,
	}

	respBody := struct {
		Data struct {
			Team *Team `json:"team"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, q, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for members")
	}

	return respBody.Data.Team.Members, nil
}

func (c *Client) GetTeams(ctx context.Context) ([]Team, error) {
	q := `query {
	teams {
	  slug
	  purpose
}}`

	respBody := struct {
		Data struct {
			Teams []Team `json:"teams"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, q, nil, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for teams")
	}

	return respBody.Data.Teams, nil
}

func (c *Client) GetTeamsForUser(ctx context.Context, email string) ([]TeamMembership, error) {
	q := `query userByEmail($email: String!) {
	userByEmail(email: $email) {
	  teams {
		team {
		  slug
		  purpose
		}
	  }
	}
  }`
	vars := map[string]string{
		"email": email,
	}

	respBody := struct {
		Data struct {
			UserByEmail *User `json:"userByEmail"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.teamsQuery(ctx, q, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for user teams")
	}

	return respBody.Data.UserByEmail.Teams, nil
}

func (c *Client) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	q := `query GetUser($id: UUID!) {
	user(id: $id) {
		name
		email
	}
}`
	vars := map[string]string{"id": id}
	respBody := struct {
		Data struct {
			UserByID *struct{ Name, Email string } `json:"user"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}
	if err := c.teamsQuery(ctx, q, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for user")
	}
	if respBody.Data.UserByID == nil {
		return nil, fmt.Errorf("user %s not found", id)
	}
	user := &model.User{
		ID:    model.Ident{ID: id, Type: "user"},
		Name:  respBody.Data.UserByID.Name,
		Email: respBody.Data.UserByID.Email,
	}

	return user, nil
}

func (c *Client) GetUser(ctx context.Context, email string) (*User, error) {
	q := `query GetUser($email: String!) {
	userByEmail(email: $email) {
		name
		id
	}
}`
	vars := map[string]string{
		"email": email,
	}
	respBody := struct {
		Data struct {
			UserByEmail *User `json:"userByEmail"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}
	if err := c.teamsQuery(ctx, q, vars, &respBody); err != nil {
		return nil, c.error(ctx, err, "querying teams for user")
	}
	if respBody.Data.UserByEmail == nil {
		return nil, fmt.Errorf("user %s not found", email)
	}

	return respBody.Data.UserByEmail, nil
}

func (c *Client) teamsQuery(ctx context.Context, query string, vars map[string]string, respBody interface{}) error {
	q := struct {
		Query     string            `json:"query"`
		Variables map[string]string `json:"variables"`
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

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	return nil
}

func (c *Client) error(ctx context.Context, err error, msg string) error {
	c.errors.Add(context.Background(), 1, metric.WithAttributes(attribute.String("component", "teams-client")))
	c.log.WithError(err).Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}
