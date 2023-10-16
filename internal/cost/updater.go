package cost

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

type Updater struct {
	log           logrus.FieldLogger
	queries       gensql.Querier
	client        *bigquery.Client
	bigQueryTable string
	daysToFetch   int
	reimport      bool
}

// NewCostUpdater creates a new cost updater
func NewCostUpdater(ctx context.Context, queries gensql.Querier, tenantName string, daysToFetch int, reimport bool, log logrus.FieldLogger) (*Updater, error) {
	client, err := bigquery.NewClient(ctx, bigquery.DetectProjectID)
	if err != nil {
		return nil, err
	}

	client.Location = "EU"

	return &Updater{
		queries:       queries,
		client:        client,
		log:           log,
		bigQueryTable: "nais-io.console.cost_" + tenantName,
		daysToFetch:   daysToFetch,
		reimport:      reimport,
	}, nil
}

// Run will update the costs
func (c *Updater) Run(ctx context.Context, schedule time.Duration) {
	ticker := time.NewTicker(schedule)
	for {
		if err := c.updateCosts(ctx); err != nil {
			c.log.WithError(err).Error("failed to update costs")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// getDayIntervalForBigQuerySql returns the number of days to fetch for the SQL query. When doing a reimport, we want to
// fetch all data from 2020-01-01, which should be enough to get all historic data.
func (c *Updater) getDayIntervalForBigQuerySql() int {
	if c.reimport {
		return int(time.Now().Sub(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)
	}

	return c.daysToFetch
}

// updateCosts will insert the latest cost data into the database
// it will only do so if the current date is newer than the latest date in the database +1 day
// and the time is after 05:00
func (c *Updater) updateCosts(ctx context.Context) error {
	lastDate, err := c.queries.LastCostDate(ctx)
	if err != nil {
		return err
	}

	// Only run if the latest date recorded isn't today, and the time is after 05:00
	if lastDate.Time.Day()+1 == time.Now().Day() && time.Now().Hour() < 5 {
		return nil
	}

	sql := fmt.Sprintf(
		"SELECT * FROM `%s` WHERE `date` >= TIMESTAMP_SUB(CURRENT_DATE(), INTERVAL %d DAY)",
		c.bigQueryTable,
		c.getDayIntervalForBigQuerySql(),
	)
	c.log.WithField("query", sql).Debugf("fetch data from bigquery")
	query := c.client.Query(sql)
	it, err := query.Read(ctx)
	if err != nil {
		return err
	}

	type Row struct {
		Env      bigquery.NullString `bigquery:"env"`
		Team     bigquery.NullString `bigquery:"team"`
		App      bigquery.NullString `bigquery:"app"`
		CostType string              `bigquery:"cost_type"`
		Date     civil.Date          `bigquery:"date"`
		Cost     float32             `bigquery:"cost"`
	}

	c.log.Debugf("collect cost data")
	rows := make([]gensql.CostUpsertParams, 0)
	start := time.Now()
	for {
		var r Row
		if err := it.Next(&r); err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}

			if errors.Is(err, context.Canceled) {
				return err
			}

			c.log.WithError(err).Error("failed to read row")
			continue
		}

		rows = append(rows, gensql.CostUpsertParams{
			Env:       nullToStringPointer(r.Env),
			Team:      nullToStringPointer(r.Team),
			App:       nullToStringPointer(r.App),
			CostType:  r.CostType,
			Date:      pgtype.Date{Time: r.Date.In(time.UTC), Valid: true},
			DailyCost: r.Cost,
		})
	}
	c.log.
		WithField("duration", time.Since(start).String()).
		WithField("rows", len(rows)).
		Info("collected data")

	start = time.Now()
	c.log.Debugf("start upserting cost data")
	if c.reimport {
		c.log.Debugf("truncate existing data (reimport set to true)")
		if err := c.queries.TruncateCostTable(ctx); err != nil {
			return err
		}
	}

	c.queries.CostUpsert(ctx, rows).Exec(func(i int, err error) {
		if err != nil {
			c.log.WithError(err).Errorf("failed to upsert cost: index %v", i)
		}
	})

	c.log.
		WithField("duration", time.Since(start).String()).
		WithField("num_inserts", len(rows)).
		Info("updated cost data")

	return nil
}

// nullToStringPointer converts a bigquery.NullString to a *string
func nullToStringPointer(s bigquery.NullString) *string {
	if s.Valid {
		return &s.StringVal
	}
	return nil
}
