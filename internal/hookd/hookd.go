package hookd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

type Client struct {
	endpoint   string
	httpClient *httpClient
	log        logrus.FieldLogger
	errors     api.Int64Counter
}

func New(psk, endpoint string, errors api.Int64Counter, log logrus.FieldLogger) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &httpClient{
			client: &http.Client{},
			psk:    psk,
		},
		log:    log,
		errors: errors,
	}
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

type RequestOptions func(*http.Request)

func WithTeam(team string) RequestOptions {
	return func(req *http.Request) {
		q := req.URL.Query()
		q.Set("team", team)
		req.URL.RawQuery = q.Encode()
	}
}

func WithCluster(cluster string) RequestOptions {
	return func(req *http.Request) {
		q := req.URL.Query()
		q.Set("cluster", cluster)
		req.URL.RawQuery = q.Encode()
	}
}

func WithLimit(limit int) RequestOptions {
	return func(req *http.Request) {
		q := req.URL.Query()
		q.Set("limit", strconv.Itoa(limit))
		req.URL.RawQuery = q.Encode()
	}
}

func WithIgnoreTeams(teams ...string) RequestOptions {
	return func(req *http.Request) {
		q := req.URL.Query()
		q.Set("ignoreTeam", strings.Join(teams, ","))
		req.URL.RawQuery = q.Encode()
	}
}

func (c *Client) Deployments(ctx context.Context, opts ...RequestOptions) ([]Deploy, error) {
	url := c.endpoint + "/internal/api/v1/console/deployments"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, c.error(ctx, err, "create request for hookd")
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.error(ctx, err, "calling hookd")
	}

	var deploymentsResponse DeploymentsResponse

	if err := json.NewDecoder(resp.Body).Decode(&deploymentsResponse); err != nil {
		return nil, c.error(ctx, err, "decoding response from hookd")
	}

	ret := deploymentsResponse.Deployments

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].DeploymentInfo.Created.After(ret[j].DeploymentInfo.Created)
	})

	return ret, nil
}

func (c *Client) DeploymentsByKind(ctx context.Context, name, team, env, kind string) ([]Deploy, error) {
	deploys, err := c.Deployments(ctx, WithTeam(team), WithCluster(env))
	if err != nil {
		return nil, c.error(ctx, err, "getting deploys from hookd")
	}

	return filterByKind(deploys, name, kind), nil
}

func filterByKind(deploys []Deploy, name, kind string) []Deploy {
	ret := []Deploy{}
	for _, deploy := range deploys {
		for _, resource := range deploy.Resources {
			if resource.Name == name && resource.Kind == kind {
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
		return nil, c.error(ctx, err, "create request for deploy key API")
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, c.error(ctx, err, "calling hookd")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.error(ctx, fmt.Errorf("deploy key API returned %s", resp.Status), "deploy key API returned non-200")
	}

	return c.DeployKey(ctx, team)
}

func (c *Client) DeployKey(ctx context.Context, team string) (*DeployKey, error) {
	url := fmt.Sprintf("%s/internal/api/v1/console/apikey/%s", c.endpoint, team)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, c.error(ctx, err, "create request for deploy key API")
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.error(ctx, fmt.Errorf("deploy key API returned %s", resp.Status), "deploy key API returned non-200")
	}

	data, _ := io.ReadAll(resp.Body)
	ret := &DeployKey{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		return nil, c.error(ctx, err, "unmarshal reply from deploy API")
	}

	return ret, nil
}

func (c *Client) error(ctx context.Context, err error, msg string) error {
	c.errors.Add(ctx, 1, api.WithAttributes(attribute.String("component", "hookd-client")))
	c.log.WithError(err).Error(msg)
	return fmt.Errorf("%s: %w", msg, err)
}
