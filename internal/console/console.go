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
	Slug    string `json:"slug"`
	Purpose string `json:"purpose"`
}

var (
	HTTPClient           = http.DefaultClient
	ConsoleQueryEndpoint = "http://localhost:3000/query"
	Token                = "secret"
)

const teamQuery = `query userByEmail($email: String!) {
	userByEmail(email: $email) {
	  teams {
		team {
		  slug
		  purpose
		}
	  }
	}
  }`

const userQuery = `query GetUser($email: String!) {
	userByEmail(email: $email) {
		name
		id
	}
}`

func GetTeams(ctx context.Context, email string) ([]TeamMembership, error) {
	q := struct {
		Query     string            `json:"query"`
		Variables map[string]string `json:"variables"`
	}{
		Query: teamQuery,
		Variables: map[string]string{
			"email": email,
		},
	}

	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ConsoleQueryEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		return nil, fmt.Errorf("console: %v", resp.Status)
	}

	respBody := struct {
		Data struct {
			UserByEmail *User `json:"userByEmail"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	if len(respBody.Errors) > 0 {
		return nil, fmt.Errorf("console: %v", respBody.Errors)
	}

	return respBody.Data.UserByEmail.Teams, nil
}

func GetUser(ctx context.Context, email string) (*User, error) {
	q := struct {
		Query     string            `json:"query"`
		Variables map[string]string `json:"variables"`
	}{
		Query: userQuery,
		Variables: map[string]string{
			"email": email,
		},
	}

	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ConsoleQueryEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		return nil, fmt.Errorf("console: %v", resp.Status)
	}

	respBody := struct {
		Data struct {
			UserByEmail *User `json:"userByEmail"`
		} `json:"data"`
		Errors []map[string]any `json:"errors"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	if len(respBody.Errors) > 0 {
		return nil, fmt.Errorf("console: %v", respBody.Errors)
	}

	return respBody.Data.UserByEmail, nil
}
