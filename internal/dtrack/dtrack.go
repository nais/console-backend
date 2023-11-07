package dtrack

import (
	"context"
	"fmt"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/nais/console-backend/internal/graph/scalar"
	dependencytrack "github.com/nais/dependencytrack/pkg/client"
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
}

func New(cfg config.DTrack, errors api.Int64Counter, log *logrus.Entry) *Client {
	c := dependencytrack.New(
		cfg.Endpoint,
		cfg.Username,
		cfg.Password,
		dependencytrack.WithApiKeySource("Administrators"),
		dependencytrack.WithLogger(log),
	)
	return &Client{
		client:      c,
		frontendUrl: cfg.Frontend,
		log:         log,
		errors:      errors,
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

func (c *Client) VulnerabilitySummary(ctx context.Context, cluster, image string) (*model.DependencyTrack, error) {
	p, err := c.projectByImageAndCluster(ctx, image, cluster)
	if err != nil {
		return nil, fmt.Errorf("getting project by image %s and cluster %s: %w", image, cluster, err)
	}

	if p == nil {
		c.log.Infof("no project found in DependencyTrack for image %s and cluster %s", image, cluster)
		return nil, nil
	}

	return c.findings(ctx, p.Uuid, p.Name)
}

func (c *Client) AddFindings(ctx context.Context, team string, nodes []*model.VulnerabilitiesNode) error {
	var wg sync.WaitGroup
	now := time.Now()
	projects, err := c.client.GetProjectsByTag(ctx, team)
	if err != nil {
		return fmt.Errorf("getting projects from DependencyTrack: %w", err)
	}

	dts := make([]*model.DependencyTrack, 0)
	for _, p := range projects {
		wg.Add(1)
		go func(uuid, name string) {
			defer wg.Done()
			d, err := c.findings(ctx, uuid, name)
			if err != nil {
				c.log.Errorf("getting findings for project %s: %v", name, err)
				return
			}
			if d == nil {
				c.log.Infof("no findings found in DependencyTrack for project %s", name)
				return
			}
			dts = append(dts, d)
		}(p.Uuid, p.Name)
	}
	wg.Wait()
	fmt.Printf("DependencyTrack fetch: %v\n", time.Since(now))

	for _, node := range nodes {
		for _, d := range dts {
			if strings.HasPrefix(d.ProjectName, fmt.Sprintf("%s:%s:%s", node.Env, team, node.AppName)) {
				node.Project = d
				break
			}
		}
	}
	fmt.Printf("DependencyTrack nodes loop: %v\n", time.Since(now))
	return nil
}

func projectNameStartsWith(projects []*dependencytrack.Project, s string) *dependencytrack.Project {
	for _, project := range projects {
		if strings.HasPrefix(project.Name, s) {
			return project
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

func (c *Client) projectByImageAndCluster(ctx context.Context, image, cluster string) (*dependencytrack.Project, error) {
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
		for _, t := range project.Tags {
			if t.Name == cluster {
				p = project
				break
			}
		}
	}
	return p, nil
}

func (c *Client) getProjects(ctx context.Context, apps []*model.App) []*dependencytrack.Project {
	var wg sync.WaitGroup
	projects := make([]*dependencytrack.Project, 0)
	for _, app := range apps {
		wg.Add(1)
		go func(image, env string) {
			defer wg.Done()

			p, err := c.projectByImageAndCluster(ctx, image, env)
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
		}(app.Image, app.Env.Name)
	}
	wg.Wait()
	return projects
}
