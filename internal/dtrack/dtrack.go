package dtrack

import (
	"context"
	"fmt"
	"github.com/nais/console-backend/internal/config"
	"github.com/nais/console-backend/internal/graph/model"
	dependencytrack "github.com/nais/dependencytrack/pkg/client"
	"github.com/sirupsen/logrus"
	api "go.opentelemetry.io/otel/metric"
	"net/url"
	"strings"
)

type Client struct {
	client      dependencytrack.Client
	frontendUrl string
	log         logrus.FieldLogger
	errors      api.Int64Counter
}

func New(cfg config.DTrack, errors api.Int64Counter, log *logrus.Entry) *Client {
	c := dependencytrack.New(cfg.Endpoint, cfg.Username, cfg.Password, dependencytrack.WithLogger(log))
	return &Client{
		client:      c,
		frontendUrl: cfg.Frontend,
		log:         log,
		errors:      errors,
	}
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

	url := strings.TrimSuffix(c.frontendUrl, "/")
	findingsLink := fmt.Sprintf("%s/projects/%s/findings", url, p.Uuid)

	findings, err := c.client.GetFindings(ctx, p.Uuid)
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
		ProjectUUID:     p.Uuid,
		ProjectName:     p.Name,
		FindingsLink:    findingsLink,
		Vulnerabilities: v,
		Summary:         summary,
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
