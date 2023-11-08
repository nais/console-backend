package dtrack

import (
	"context"
	"fmt"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	dependencytrack "github.com/nais/dependencytrack/pkg/client"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	api "go.opentelemetry.io/otel/metric"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	client      dependencytrack.Client
	frontendUrl string
	log         logrus.FieldLogger
	errors      api.Int64Counter
	cache       *cache.Cache
}

func New(cfg config.DTrack, errors api.Int64Counter, log *logrus.Entry) *Client {
	c := dependencytrack.New(
		cfg.Endpoint,
		cfg.Username,
		cfg.Password,
		dependencytrack.WithApiKeySource("Administrators"),
		dependencytrack.WithLogger(log),
	)

	ch := cache.New(5*time.Minute, 10*time.Minute)

	return &Client{
		client:      c,
		frontendUrl: cfg.Frontend,
		log:         log,
		errors:      errors,
		cache:       ch,
	}
}

func (c *Client) Init(ctx context.Context) error {
	_, err := c.client.Headers(ctx)
	if err != nil {
		return fmt.Errorf("initializing DependencyTrack client: %w", err)
	}
	return nil
}

func (c *Client) WithClient(client dependencytrack.Client) *Client {
	c.client = client
	return c
}

func (c *Client) VulnerabilitySummary(ctx context.Context, env, app, image string) (*model.DependencyTrack, error) {
	p, err := c.projectByImageAndCluster(ctx, env, app, image)
	if err != nil {
		return nil, fmt.Errorf("getting project by env %q app %q image %q: %w", env, app, image, err)
	}

	if p == nil {
		c.log.Infof("no project found in DependencyTrack for env %q app %q image %q:", env, app, image)
		return nil, nil
	}

	return c.findings(ctx, p.Uuid, p.Name)
}

func (c *Client) AddFindings(ctx context.Context, nodes []*model.VulnerabilitiesNode) error {
	var wg sync.WaitGroup
	now := time.Now()

	dts := make([]*model.DependencyTrack, 0)
	for _, n := range nodes {
		wg.Add(1)
		go func(env, team, app, image string) {
			defer wg.Done()
			p, err := c.projectByImage(ctx, image, env, team, app)
			if err != nil {
				return
			}
			if p == nil {
				return
			}

			d, err := c.findings(ctx, p.Uuid, p.Name)
			if err != nil {
				c.log.Errorf("getting findings for project %s: %v", p.Name, err)
				return
			}
			if d == nil {
				c.log.Infof("no findings found in DependencyTrack for project %s", p.Name)
				return
			}
			dts = append(dts, d)
		}(n.Env, n.Team, n.AppName, n.Image)
	}
	wg.Wait()
	fmt.Printf("DependencyTrack fetch: %v\n", time.Since(now))

	for _, node := range nodes {
		fmt.Printf("App: %s, Env: %s, Image:%s\n", node.AppName, node.Env, node.Image)
		for _, d := range dts {
			if d.ProjectName == fmt.Sprintf("%s:%s:%s", node.Env, node.Team, node.AppName) {
				node.Project = d
				break
			}
		}
	}
	return nil
}

func (c *Client) findings(ctx context.Context, uuid, name string) (*model.DependencyTrack, error) {
	url := strings.TrimSuffix(c.frontendUrl, "/")
	findingsLink := fmt.Sprintf("%s/projects/%s/findings", url, uuid)

	findings, err := c.client.GetFindings(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("getting findings from DependencyTrack: %w", err)
	}

	var low, medium, high, critical, unassigned int

	v := make([]model.Vulnerability, 0)
	for _, finding := range findings {
		switch finding.Vulnerability.Severity {
		case "LOW":
			low += 1
		case "MEDIUM":
			medium += 1
		case "HIGH":
			high += 1
		case "CRITICAL":
			critical += 1
		case "UNASSIGNED":
			unassigned += 1
		}

		v = append(v, model.Vulnerability{
			ID:            finding.Vulnerability.UUID,
			Severity:      finding.Vulnerability.Severity,
			SeverityRank:  finding.Vulnerability.SeverityRank,
			Name:          finding.Vulnerability.Name,
			ComponentPurl: finding.Component.PURL,
		})
	}

	summary := model.VulnerabilitySummary{
		Total:      len(findings),
		Critical:   critical,
		High:       high,
		Medium:     medium,
		Low:        low,
		Unassigned: unassigned,
	}

	d := &model.DependencyTrack{
		ID:              scalar.DependencyTrackIdent(uuid),
		ProjectUUID:     uuid,
		ProjectName:     name,
		FindingsLink:    findingsLink,
		Vulnerabilities: v,
		Summary:         &summary,
	}
	return d, nil
}

func (c *Client) projectByImage(ctx context.Context, image string, tags ...string) (*dependencytrack.Project, error) {
	tag := url.QueryEscape(image)
	projects, err := c.client.GetProjectsByTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("getting projects from DependencyTrack: %w", err)
	}

	if len(projects) == 0 {
		return nil, nil
	}
	var p *dependencytrack.Project
	for _, project := range projects {
		if containsAllTags(project.Tags, tags...) {
			p = project
			break
		}
	}
	return p, nil
}

func (c *Client) projectByImageAndCluster(ctx context.Context, env, app, image string) (*dependencytrack.Project, error) {
	tag := url.QueryEscape(image)
	projects, err := c.client.GetProjectsByTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("getting projects from DependencyTrack: %w", err)
	}

	if len(projects) == 0 {
		return nil, nil
	}

	var p *dependencytrack.Project
	for _, project := range projects {
		if containsAllTags(project.Tags, env, app, image) {
			p = project
			break
		}
	}
	return p, nil
}

func containsAllTags(tags []dependencytrack.Tag, s ...string) bool {
	found := 0
	for _, tag := range tags {
		for _, t := range s {
			if tag.Name == t {
				found += 1
				break
			}
		}
	}
	return found == len(s)
}

func (c *Client) getProjects(ctx context.Context, apps []*model.VulnerabilitiesNode) []*dependencytrack.Project {
	var wg sync.WaitGroup
	projects := make([]*dependencytrack.Project, 0)
	for _, app := range apps {
		wg.Add(1)
		go func(env, app, image string) {
			defer wg.Done()

			p, err := c.projectByImageAndCluster(ctx, env, app, image)
			if err != nil {
				c.log.Errorf("getting project by image %s and cluster %s: %v", image, env, err)
				//c.errors.Add(ctx, 1)
				return
			}
			if p == nil {
				c.log.Infof("no project found in DependencyTrack for image %s and cluster %s", image, env)
				return
			}
			projects = append(projects, p)
		}(app.Env, app.AppName, app.Image)
	}
	wg.Wait()
	return projects
}
