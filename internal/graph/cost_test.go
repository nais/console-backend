package graph

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/nais/console-backend/internal/graph/model"
	"github.com/stretchr/testify/assert"
)

func TestDailyCostsFromDatabaseRows(t *testing.T) {
	fromTime := time.Date(2020, time.April, 28, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2020, time.May, 2, 0, 0, 0, 0, time.UTC)
	from := model.NewDate(fromTime)
	to := model.NewDate(endTime)

	t.Run("no cost types present", func(t *testing.T) {
		existingCosts := make([]*gensql.Cost, 0)
		costs := DailyCostsFromDatabaseRows(from, to, existingCosts)
		assert.Len(t, costs, 0)
	})

	t.Run("add missing and sort dates", func(t *testing.T) {
		const allowedDelta = 0.000001

		existingCosts2 := []*gensql.Cost{
			{
				CostType: "type",
				Date:     pgtype.Date{Time: fromTime, Valid: true},
				Cost:     float32(1.1),
			},
			{
				CostType: "type",
				Date:     pgtype.Date{Time: fromTime.AddDate(0, 0, 1), Valid: true},
				Cost:     float32(2.1),
			},
			{
				CostType: "type",
				Date:     pgtype.Date{Time: fromTime.AddDate(0, 0, 2), Valid: true},
				Cost:     float32(3.1),
			},
			{
				CostType: "type",
				Date:     pgtype.Date{Time: fromTime.AddDate(0, 0, 4), Valid: true},
				Cost:     float32(5.1),
			},

			{
				CostType: "type2",
				Date:     pgtype.Date{Time: fromTime.AddDate(0, 0, 4), Valid: true},
				Cost:     float32(5.2),
			},
			{
				CostType: "type2",
				Date:     pgtype.Date{Time: fromTime.AddDate(0, 0, 2), Valid: true},
				Cost:     float32(3.2),
			},
		}
		costs := DailyCostsFromDatabaseRows(from, to, existingCosts2)
		assert.Len(t, costs, 2)

		assert.Len(t, costs["type"], 5)
		assert.InDelta(t, 1.1, costs["type"][0].Cost, allowedDelta)
		assert.InDelta(t, 2.1, costs["type"][1].Cost, allowedDelta)
		assert.InDelta(t, 3.1, costs["type"][2].Cost, allowedDelta)
		assert.InDelta(t, 0.0, costs["type"][3].Cost, allowedDelta)
		assert.InDelta(t, 5.1, costs["type"][4].Cost, allowedDelta)
		assert.Equal(t, "2020-04-28", string(costs["type"][0].Date))
		assert.Equal(t, "2020-04-29", string(costs["type"][1].Date))
		assert.Equal(t, "2020-04-30", string(costs["type"][2].Date))
		assert.Equal(t, "2020-05-01", string(costs["type"][3].Date))
		assert.Equal(t, "2020-05-02", string(costs["type"][4].Date))

		assert.Len(t, costs["type2"], 5)
		assert.InDelta(t, 0.0, costs["type2"][0].Cost, allowedDelta)
		assert.InDelta(t, 0.0, costs["type2"][1].Cost, allowedDelta)
		assert.InDelta(t, 3.2, costs["type2"][2].Cost, allowedDelta)
		assert.InDelta(t, 0.0, costs["type2"][3].Cost, allowedDelta)
		assert.InDelta(t, 5.2, costs["type2"][4].Cost, allowedDelta)
		assert.Equal(t, "2020-04-28", string(costs["type2"][0].Date))
		assert.Equal(t, "2020-04-29", string(costs["type2"][1].Date))
		assert.Equal(t, "2020-04-30", string(costs["type2"][2].Date))
		assert.Equal(t, "2020-05-01", string(costs["type2"][3].Date))
		assert.Equal(t, "2020-05-02", string(costs["type2"][4].Date))
	})
}
