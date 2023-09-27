package cost

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nais/console-backend/internal/database/gensql"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

const (
	gcpProject = "nais-io"
	query      = `
SELECT *
FROM
  ` + "`nais-io.console_data.console_nav`" + `
WHERE
  date >= TIMESTAMP_SUB(CURRENT_DATE(), INTERVAL 3 DAY)`
)

type CostUpdater struct {
	log     logrus.FieldLogger
	queries *gensql.Queries
	client  *bigquery.Client
}

func NewCostUpdater(ctx context.Context, queries *gensql.Queries, log logrus.FieldLogger) (*CostUpdater, error) {
	client, err := bigquery.NewClient(ctx, gcpProject)
	if err != nil {
		return nil, err
	}
	client.Location = "EU"
	it := client.Datasets(ctx)
	for {
		_, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return &CostUpdater{
		queries: queries,
		client:  client,
		log:     log,
	}, nil
}

func (c *CostUpdater) Run(ctx context.Context, schedule time.Duration) {
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

// updateCosts will insert the latest cost data into the database
// it will only do so if the current date is newer than the latest date in the database +1 day
// and the time is after 05:00
func (c *CostUpdater) updateCosts(ctx context.Context) error {
	lastDate, err := c.queries.CostLastDate(ctx)
	if err != nil {
		return err
	}

	// Only run if the latest date recorded isn't today, and the time is after 05:00
	if lastDate.Time.Day()+1 == time.Now().Day() && time.Now().Hour() < 5 {
		return nil
	}

	q := c.client.Query(query)
	it, err := q.Read(ctx)
	if err != nil {
		return err
	}

	type Row struct {
		Env      bigquery.NullString `bigquery:"env"`
		Team     bigquery.NullString `bigquery:"team"`
		App      bigquery.NullString `bigquery:"app"`
		CostType string              `bigquery:"cost_type"`
		Date     civil.Date          `bigquery:"date"`
		Cost     float32             `bigquery:"total"`
	}

	start := time.Now()
	rows := make([]gensql.CostUpsertParams, 0)
	for {

		var r Row
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			if err == context.Canceled {
				return err
			}

			c.log.WithError(err).Error("failed to read row")
			continue
		}

		rows = append(rows, gensql.CostUpsertParams{
			Env:      nullToPointerString(r.Env),
			Team:     nullToPointerString(r.Team),
			App:      nullToPointerString(r.App),
			Date:     pgtype.Date{Time: r.Date.In(time.UTC), Valid: true},
			Cost:     r.Cost,
			CostType: r.CostType,
		})
	}

	c.queries.CostUpsert(ctx, rows).Exec(func(i int, err error) {
		if err != nil {
			c.log.WithError(err).Errorf("failed to upsert cost: index %v", i)
		}
	})

	c.log.WithField("duration", time.Since(start).String()).WithField("num_inserts", len(rows)).Info("inserted costs")

	return nil
}

func nullToPointerString(s bigquery.NullString) *string {
	if s.Valid {
		return &s.StringVal
	}
	return nil
}
