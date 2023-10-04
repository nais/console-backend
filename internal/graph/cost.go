package graph

import (
	"time"

	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
)

type (
	dailyCosts       map[string]map[model.Date]float64
	sortedDailyCosts map[string][]*model.CostEntry
)

// DailyCostsFromDatabaseRows will convert a slice of cost rows from the database to a sortedDailyCosts map.
func DailyCostsFromDatabaseRows(from model.Date, to model.Date, rows []*gensql.Cost) sortedDailyCosts {
	daily := dailyCosts{}
	for _, row := range rows {
		if _, exists := daily[row.CostType]; !exists {
			daily[row.CostType] = make(map[model.Date]float64)
		}
		daily[row.CostType][model.NewDate(row.Date.Time)] = float64(row.Cost)
	}

	return normalizeDailyCosts(from, to, daily)
}

func DailyCostsForTeamFromDatabaseRows(from model.Date, to model.Date, rows []*gensql.Cost) sortedDailyCosts {
	daily := dailyCosts{}
	for _, row := range rows {
		// TODO: hack to filter out unknown apps (ie aiven costs). Remove when we have a better solution for aiven costs.
		if row.App != nil && *row.App == "<unknown>" {
			continue
		}
		if _, exists := daily[row.CostType]; !exists {
			daily[row.CostType] = make(map[model.Date]float64)
		}
		date := model.NewDate(row.Date.Time)
		if _, exists := daily[row.CostType][date]; !exists {
			daily[row.CostType][date] = float64(row.Cost)
		} else {
			daily[row.CostType][date] += float64(row.Cost)
		}
	}

	return normalizeDailyCosts(from, to, daily)
}

// normalizeDailyCosts will make sure all dates in the "from -> to" range are present in the returned map for all cost
// types. The dates will also be sorted in ascending order.
func normalizeDailyCosts(from, to model.Date, costs dailyCosts) sortedDailyCosts {
	start, _ := time.Parse(model.DateFormat, string(from))
	end, _ := time.Parse(model.DateFormat, string(to))
	sortedDailyCost := make(map[string][]*model.CostEntry)
	for k, daysInSeries := range costs {
		data := make([]*model.CostEntry, 0)
		for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
			date := model.NewDate(day)
			cost := 0.0
			if c, exists := daysInSeries[date]; exists {
				cost = c
			}

			data = append(data, &model.CostEntry{
				Date: date,
				Cost: cost,
			})
		}

		sortedDailyCost[k] = data
	}

	return sortedDailyCost
}
