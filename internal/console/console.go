package console

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
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
}

func New(token, endpoint string) *Client {
	return &Client{endpoint: endpoint, httpClient: Transport{Token: token}.Client()}
}

func (c *Client) GetTeam(ctx context.Context, name string) (*Team, error) {
	q := `query team($slug: Slug!) {
	team(slug: $slug) {
	  slug
	  purpose
	  slackChannel
	  gitHubRepositories{
		name
	  }
	  slackAlertsChannels{
		environment
		channelName
	  }
}}`

	vars := map[string]string{
		"slug": name,
	}

	respBody := struct {
		Data struct {
			Team *Team `json:"team"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := c.consoleQuery(ctx, q, vars, &respBody); err != nil {
		return nil, fmt.Errorf("querying console: %w", err)
	}

	return respBody.Data.Team, nil
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

	if err := c.consoleQuery(ctx, q, vars, &respBody); err != nil {
		return nil, err
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

	if err := c.consoleQuery(ctx, q, vars, &respBody); err != nil {
		return nil, fmt.Errorf("querying console: %w", err)
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

	if err := c.consoleQuery(ctx, q, nil, &respBody); err != nil {
		return nil, fmt.Errorf("querying console: %w", err)
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

	if err := c.consoleQuery(ctx, q, vars, &respBody); err != nil {
		return nil, fmt.Errorf("querying console: %w", err)
	}

	return respBody.Data.UserByEmail.Teams, nil
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
	if err := c.consoleQuery(ctx, q, vars, &respBody); err != nil {
		return nil, fmt.Errorf("querying console: %w", err)
	}
	if respBody.Data.UserByEmail == nil {
		return nil, fmt.Errorf("user %s not found", email)
	}

	return respBody.Data.UserByEmail, nil
}

func (c *Client) consoleQuery(ctx context.Context, query string, vars map[string]string, respBody interface{}) error {
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
		return fmt.Errorf("console: %v", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	return nil
}
