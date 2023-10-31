package dtrack

import (
	"context"
	"fmt"
	"github.com/nais/console-backend/internal/graph/model"
	dependencytrack "github.com/nais/dependencytrack/pkg/client"
	"net/url"
)

func VulnerabilitySummary(ctx context.Context, dtrackClient dependencytrack.Client, cluster, image string) (*model.DependencyTrack, error) {
	tag := url.QueryEscape(image)
	projects, err := dtrackClient.GetProjectsByTag(ctx, tag)
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
	if p == nil {
		return nil, nil
	}

	// TODO: use salsa frontend url, not client baseurl
	findingsLink := fmt.Sprintf("%s/projects/%s/findings", "baseurl", p.Uuid)

	// https://salsa.nav.cloud.nais.io/projects/4381f963-e53b-4804-8084-7ede767f9006/findings

	findings, err := dtrackClient.GetFindings(ctx, p.Uuid)
	if err != nil {
		return nil, fmt.Errorf("getting findings from DependencyTrack: %w", err)
	}
	//fmt.Printf("findings: %+v\n", findings)

	low := make([]*dependencytrack.Finding, 0)
	medium := make([]*dependencytrack.Finding, 0)
	high := make([]*dependencytrack.Finding, 0)
	critical := make([]*dependencytrack.Finding, 0)
	unassigned := make([]*dependencytrack.Finding, 0)
	v := make([]model.Vulnerability, 0)
	for _, finding := range findings {
		switch finding.Vulnerability.Severity {
		case "LOW":
			low = append(low, finding)
		case "MEDIUM":
			medium = append(medium, finding)
		case "HIGH":
			high = append(high, finding)
		case "CRITICAL":
			critical = append(critical, finding)
		case "UNASSIGNED":
			unassigned = append(unassigned, finding)
		}
		v = append(v, model.Vulnerability{
			ID:            finding.Vulnerability.UUID,
			Severity:      finding.Vulnerability.Severity,
			SeverityRank:  finding.Vulnerability.SeverityRank,
			Name:          finding.Vulnerability.Name,
			ComponentPurl: finding.Component.PURL,
		})
		fmt.Printf("finding: %+v\n", finding)
	}

	summary := model.VulnerabilitySummary{
		Total:      len(low) + len(medium) + len(high) + len(critical) + len(unassigned),
		Critical:   len(critical),
		High:       len(high),
		Medium:     len(medium),
		Low:        len(low),
		Unassigned: len(unassigned),
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
