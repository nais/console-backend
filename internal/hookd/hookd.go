package hookd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	psk        string
	endpoint   string
	httpClient *http.Client
}

func New(psk, endpoint string) *Client {
	return &Client{psk: psk, endpoint: endpoint, httpClient: Transport{PSK: psk}.Client()}
}

type DeploymentsResponse struct {
	Deployments []Deploy `json:"deployments"`
}

type Deploy struct {
	DeploymentInfo DeploymentInfo `json:"deployment"`
	Statuses       []Status       `json:"statuses"`
	Resources      []Resource     `json:"resources"`
}

type DeploymentInfo struct {
	ID      string `json:"id"`
	Team    string `json:"team"`
	Cluster string `json:"cluster"`
}

type Status struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Created time.Time `json:"created"`
}

type Resource struct {
	ID        string `json:"id"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Namespace string `json:"namespace"`
}

func (c *Client) GetDeploysForTeam(ctx context.Context, team string) ([]Deploy, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/dashboard/deployments?team=%s", c.endpoint, team), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	var deploymentsResponse DeploymentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&deploymentsResponse); err != nil {
		return nil, fmt.Errorf("decoding hookd response: %w", err)
	}

	return deploymentsResponse.Deployments, nil
}