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

type AppInstance struct {
	Env, Team, App, Image string
}

func (d *AppInstance) ID() string {
	return fmt.Sprintf("%s:%s:%s:%s", d.Env, d.Team, d.App, d.Image)
}

func (d *AppInstance) ProjectName() string {
	return fmt.Sprintf("%s:%s:%s", d.Env, d.Team, d.App)
}

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

func (c *Client) VulnerabilitySummary(ctx context.Context, app *AppInstance) (*model.DependencyTrack, error) {
	return c.findingsForApp(ctx, app)
}

func (c *Client) AddFindings(ctx context.Context, nodes []*model.VulnerabilitiesNode) error {
	var wg sync.WaitGroup
	now := time.Now()

	dts := make([]*model.DependencyTrack, 0)
	for _, n := range nodes {
		wg.Add(1)
		go func(app *AppInstance) {
			defer wg.Done()
			d, err := c.findingsForApp(ctx, app)
			if err != nil {
				c.log.Errorf("retrieveFindings for app %q: %v", app.ID(), err)
				return
			}
			if d == nil {
				c.log.Debugf("no findings found in DependencyTrack for app %q", app.ID())
				return
			}
			dts = append(dts, d)
		}(&AppInstance{n.Env, n.Team, n.AppName, n.Image})
	}
	wg.Wait()

	for _, node := range nodes {
		for _, d := range dts {
			if d.ProjectName == fmt.Sprintf("%s:%s:%s", node.Env, node.Team, node.AppName) {
				node.Project = d
				break
			}
		}
	}
	c.log.Debugf("DependencyTrack fetch: %v\n", time.Since(now))
	return nil
}

func (c *Client) findingsForApp(ctx context.Context, app *AppInstance) (*model.DependencyTrack, error) {
	if d, ok := c.cache.Get(app.ID()); ok {
		return d.(*model.DependencyTrack), nil
	}
	p, err := c.retrieveProject(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("getting project by app %s: %w", app.ID(), err)
	}
	if p == nil {
		return nil, nil
	}

	u := strings.TrimSuffix(c.frontendUrl, "/")
	findingsLink := fmt.Sprintf("%s/projects/%s/findings", u, p.Uuid)

	d := &model.DependencyTrack{
		ID:           scalar.DependencyTrackIdent(p.Uuid),
		ProjectUUID:  p.Uuid,
		ProjectName:  p.Name,
		FindingsLink: findingsLink,
		HasBom:       p.LastBomImportFormat != "",
	}

	if !d.HasBom {
		c.log.Debugf("no bom found in DependencyTrack for project %s", p.Name)
		d.Summary = c.createSummary([]*dependencytrack.Finding{}, p.LastInheritedRiskScore)
		c.cache.Set(app.ID(), d, cache.DefaultExpiration)
		return d, nil
	}

	f, err := c.retrieveFindings(ctx, p.Uuid)
	if err != nil {
		return nil, err
	}

	d.Summary = c.createSummary(f, p.LastInheritedRiskScore)

	if d == nil {
		c.log.Debugf("no findings found in DependencyTrack for project %s", p.Name)
		return nil, nil
	}

	c.cache.Set(app.ID(), d, cache.DefaultExpiration)
	return d, nil
}

func (c *Client) retrieveFindings(ctx context.Context, uuid string) ([]*dependencytrack.Finding, error) {
	findings, err := c.client.GetFindings(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("retrieveFindings from DependencyTrack: %w", err)
	}

	return findings, nil
}

func (c *Client) createSummary(findings []*dependencytrack.Finding, riskScore float64) *model.VulnerabilitySummary {
	var low, medium, high, critical, unassigned int
	if len(findings) == 0 {
		return &model.VulnerabilitySummary{
			RiskScore:  riskScore,
			Total:      -1,
			Critical:   -1,
			High:       -1,
			Medium:     -1,
			Low:        -1,
			Unassigned: -1,
		}
	}

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
	}

	return &model.VulnerabilitySummary{
		Total:      len(findings),
		RiskScore:  riskScore,
		Critical:   critical,
		High:       high,
		Medium:     medium,
		Low:        low,
		Unassigned: unassigned,
	}
}
func (c *Client) retrieveProject(ctx context.Context, app *AppInstance) (*dependencytrack.Project, error) {
	tag := url.QueryEscape(app.Image)
	projects, err := c.client.GetProjectsByTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("getting projects from DependencyTrack: %w", err)
	}

	if len(projects) == 0 {
		return nil, nil
	}
	var p *dependencytrack.Project
	for _, project := range projects {
		if containsAllTags(project.Tags, app.Env, app.Team, app.App) {
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
