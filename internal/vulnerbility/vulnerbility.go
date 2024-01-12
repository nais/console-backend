package vulnerbility

import "github.com/nais/console-backend/internal/graph/model"

type Vuln struct {
	*model.VulnerabilitySummary
}

func New(total int) *Vuln {
	return &Vuln{
		VulnerabilitySummary: &model.VulnerabilitySummary{
			Total: total,
		},
	}
}

func (v *Vuln) RiskScore() *model.VulnerabilitySummary {
	// algorithm: https://github.com/DependencyTrack/dependency-track/blob/41e2ba8afb15477ff2b7b53bd9c19130ba1053c0/src/main/java/org/dependencytrack/metrics/Metrics.java#L31-L33
	v.VulnerabilitySummary.RiskScore = (v.Critical * 10) + (v.High * 5) + (v.Medium * 3) + (v.Low * 1) + (v.Unassigned * 5)
	return v.VulnerabilitySummary
}

func (v *Vuln) Count(severity string) {
	switch severity {
	case "LOW":
		v.Low += 1
	case "MEDIUM":
		v.Medium += 1
	case "HIGH":
		v.High += 1
	case "CRITICAL":
		v.Critical += 1
	case "UNASSIGNED":
		v.Unassigned += 1
	}
}
