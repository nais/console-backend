package vulnerabilities

import "github.com/nais/console-backend/internal/graph/model"

func Sort(v []*model.VulnerabilitiesNode, field model.VulnerabilitiesOrderByField, direction model.SortOrder) {
	switch field {
	case model.VulnerabilitiesOrderByFieldAppName:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			return model.Compare(a.AppName, b.AppName, direction)
		})
	case model.VulnerabilitiesOrderByFieldEnvName:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			return model.Compare(a.Env, b.Env, direction)
		})
	case model.VulnerabilitiesOrderByFieldRiskScore:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Project.Summary.RiskScore, b.Project.Summary.RiskScore, direction)
		})
	case model.VulnerabilitiesOrderByFieldSeverityCritical:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Project.Summary.Critical, b.Project.Summary.Critical, direction)
		})
	case model.VulnerabilitiesOrderByFieldSeverityHigh:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Project.Summary.High, b.Project.Summary.High, direction)
		})
	case model.VulnerabilitiesOrderByFieldSeverityMedium:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Project.Summary.Medium, b.Project.Summary.Medium, direction)
		})
	case model.VulnerabilitiesOrderByFieldSeverityLow:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Project.Summary.Low, b.Project.Summary.Low, direction)
		})
	}
}

func summaryIsNil(a *model.VulnerabilitiesNode, b *model.VulnerabilitiesNode, direction model.SortOrder) (isNil bool, returnValue bool) {
	if a.Project == nil || a.Project.Summary == nil {
		isNil = true
		returnValue = direction == model.SortOrderAsc
	}
	if b.Project == nil || b.Project.Summary == nil {
		isNil = true
		returnValue = direction == model.SortOrderDesc
	}
	return isNil, returnValue
}
