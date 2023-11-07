package model

import (
	"sort"
)

type SortableApps struct {
	orderBy *AppsOrderBy
	apps    []*App
}

func NewSortableApps(apps []*App, orderBy *AppsOrderBy) *SortableApps {
	return &SortableApps{orderBy: orderBy, apps: apps}
}

func (a *SortableApps) Sort() []*App {
	switch *a.orderBy.Field {
	case AppsOrderByFieldName:
		sort.SliceStable(a.apps, func(i, j int) bool {
			if *a.orderBy.Direction == SortAsc {
				return a.apps[i].Name < a.apps[j].Name
			} else {
				return a.apps[i].Name > a.apps[j].Name
			}
		})
	case AppsOrderByFieldSeverityCritical, AppsOrderByFieldSeverityHigh, AppsOrderByFieldSeverityMedium, AppsOrderByFieldSeverityLow:
		sort.SliceStable(a.apps, func(i, j int) bool {
			return a.compareVulnerabilitySummary(a.apps[i].DependencyTrack, a.apps[j].DependencyTrack)
		})
	}
	return a.apps
}

func (a *SortableApps) compareVulnerabilitySummary(first, second *DependencyTrack) bool {
	if *a.orderBy.Direction == SortAsc {
		if first == nil {
			return true
		}
		if second == nil {
			return false
		}
		switch *a.orderBy.Field {
		case AppsOrderByFieldSeverityCritical:
			return first.Summary.Critical < second.Summary.Critical
		case AppsOrderByFieldSeverityHigh:
			return first.Summary.High < second.Summary.High
		case AppsOrderByFieldSeverityMedium:
			return first.Summary.Medium < second.Summary.Medium
		case AppsOrderByFieldSeverityLow:
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
		switch *a.orderBy.Field {
		case AppsOrderByFieldSeverityCritical:
			return first.Summary.Critical > second.Summary.Critical
		case AppsOrderByFieldSeverityHigh:
			return first.Summary.High > second.Summary.High
		case AppsOrderByFieldSeverityMedium:
			return first.Summary.Medium > second.Summary.Medium
		case AppsOrderByFieldSeverityLow:
			return first.Summary.Low > second.Summary.Low
		}
		return false
	}
}
