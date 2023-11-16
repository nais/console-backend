package vulnerabilities

import "github.com/nais/console-backend/internal/graph/model"

func Sort(v []*model.VulnerabilitiesNode, field model.OrderByField, direction model.SortOrder) {
	switch field {
	case model.OrderByFieldName:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			return model.Compare(a.AppName, b.AppName, direction)
		})
	case model.OrderByFieldEnv:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			return model.Compare(a.Env, b.Env, direction)
		})
	case model.OrderByFieldRiskScore:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Summary.RiskScore, b.Summary.RiskScore, direction)
		})
	case model.OrderByFieldSeverityCritical:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Summary.Critical, b.Summary.Critical, direction)
		})
	case model.OrderByFieldSeverityHigh:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Summary.High, b.Summary.High, direction)
		})
	case model.OrderByFieldSeverityMedium:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Summary.Medium, b.Summary.Medium, direction)
		})
	case model.OrderByFieldSeverityLow:
		model.SortWith(v, func(a, b *model.VulnerabilitiesNode) bool {
			isNil, returnValue := summaryIsNil(a, b, direction)
			if isNil {
				return returnValue
			}
			return model.Compare(a.Summary.Low, b.Summary.Low, direction)
		})
	}
}

func summaryIsNil(a *model.VulnerabilitiesNode, b *model.VulnerabilitiesNode, direction model.SortOrder) (isNil bool, returnValue bool) {
	if a.Summary == nil {
		isNil = true
		returnValue = direction == model.SortOrderAsc
	}
	if b.Summary == nil {
		isNil = true
		returnValue = direction == model.SortOrderDesc
	}
	return isNil, returnValue
}
