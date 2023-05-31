package hookd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
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
	ID               string    `json:"id"`
	Team             string    `json:"team"`
	Cluster          string    `json:"cluster"`
	Created          time.Time `json:"created"`
	GithubRepository string    `json:"githubRepository"`
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

func (c *Client) Deployments(ctx context.Context, team, cluster *string) ([]Deploy, error) {
	url := fmt.Sprintf("%s/internal/api/v1/console/deployments?", c.endpoint)
	if team != nil {
		url = fmt.Sprintf("%s&team=%s", url, *team)
	}
	if cluster != nil {
		url = fmt.Sprintf("%s&cluster=%s", url, *cluster)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	ret := deploymentsResponse.Deployments

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].DeploymentInfo.Created.After(ret[j].DeploymentInfo.Created)
	})

	return ret, nil
}

func (c *Client) DeploymentsByApp(ctx context.Context, app, team, env string) ([]Deploy, error) {
	deploys, err := c.Deployments(ctx, &team, &env)
	if err != nil {
		return nil, fmt.Errorf("getting deployments: %w", err)
	}

	return filterByApplication(deploys, app), nil
}

func filterByApplication(deploys []Deploy, app string) []Deploy {
	ret := []Deploy{}
	for _, deploy := range deploys {
		for _, resource := range deploy.Resources {
			if resource.Name == app && resource.Kind == "Application" {
				ret = append(ret, deploy)
				continue
			}
		}
	}
	return ret
}

type DeployKey struct {
	Team    string    `json:"team"`
	Key     string    `json:"key"`
	Expires time.Time `json:"expires"`
	Created time.Time `json:"created"`
}

func (c *Client) ChangeDeployKey(ctx context.Context, team string) (*DeployKey, error) {
	url := fmt.Sprintf("%s/internal/api/v1/console/apikey/%s", c.endpoint, team)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for deploy key API: %w", err)
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deploy key API returned %s", resp.Status)
	}

	return c.DeployKey(ctx, team)
}

func (c *Client) DeployKey(ctx context.Context, team string) (*DeployKey, error) {
	url := fmt.Sprintf("%s/internal/api/v1/console/apikey/%s", c.endpoint, team)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for deploy key API: %w", err)
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deploy key API returned %s", resp.Status)
	}

	data, _ := io.ReadAll(resp.Body)
	ret := &DeployKey{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal reply from deploy API: %v", err)
	}

	return ret, nil
}
