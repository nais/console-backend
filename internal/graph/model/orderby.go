package model

import "sort"

type SortableVulnerabilities struct {
	orderBy *VulnerabilitiesOrderBy
	nodes   []*VulnerabilitiesNode
}

func NewSortableVulnerabilities(nodes []*VulnerabilitiesNode, orderBy *VulnerabilitiesOrderBy) *SortableVulnerabilities {
	return &SortableVulnerabilities{orderBy: orderBy, nodes: nodes}
}

func (s *SortableVulnerabilities) Sort() []*VulnerabilitiesNode {
	switch s.orderBy.Field {
	case VulnerabilitiesOrderByFieldAppName:
		sort.SliceStable(s.nodes, func(i, j int) bool {
			if s.orderBy.Direction == SortAsc {
				return s.nodes[i].AppName < s.nodes[j].AppName
			} else {
				return s.nodes[i].AppName > s.nodes[j].AppName
			}
		})
	case VulnerabilitiesOrderByFieldEnvName:
		sort.SliceStable(s.nodes, func(i, j int) bool {
			if s.orderBy.Direction == SortAsc {
				return s.nodes[i].Env < s.nodes[j].Env
			} else {
				return s.nodes[i].Env > s.nodes[j].Env
			}
		})
	case VulnerabilitiesOrderByFieldSeverityCritical, VulnerabilitiesOrderByFieldSeverityHigh, VulnerabilitiesOrderByFieldSeverityMedium, VulnerabilitiesOrderByFieldSeverityLow:
		sort.SliceStable(s.nodes, func(i, j int) bool {
			return s.compareVulnerabilitySummary(s.nodes[i].Project, s.nodes[j].Project)
		})
	}
	return s.nodes
}

func (s *SortableVulnerabilities) compareVulnerabilitySummary(first, second *DependencyTrack) bool {
	if s.orderBy.Direction == SortAsc {
		if first == nil {
			return true
		}
		if second == nil {
			return false
		}
		switch s.orderBy.Field {
		case VulnerabilitiesOrderByFieldSeverityCritical:
			return first.Summary.Critical < second.Summary.Critical
		case VulnerabilitiesOrderByFieldSeverityHigh:
			return first.Summary.High < second.Summary.High
		case VulnerabilitiesOrderByFieldSeverityMedium:
			return first.Summary.Medium < second.Summary.Medium
		case VulnerabilitiesOrderByFieldSeverityLow:
			return first.Summary.Low < second.Summary.Low
		}
		return true
	} else {
		if first == nil {
			return false
		}
		if second == nil {
			return true
		}
		switch s.orderBy.Field {
		case VulnerabilitiesOrderByFieldSeverityCritical:
			return first.Summary.Critical > second.Summary.Critical
		case VulnerabilitiesOrderByFieldSeverityHigh:
			return first.Summary.High > second.Summary.High
		case VulnerabilitiesOrderByFieldSeverityMedium:
			return first.Summary.Medium > second.Summary.Medium
		case VulnerabilitiesOrderByFieldSeverityLow:
			return first.Summary.Low > second.Summary.Low
		}
		return false
	}
}
