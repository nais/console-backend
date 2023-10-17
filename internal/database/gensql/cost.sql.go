// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: cost.sql

package gensql

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const dailyCostForApp = `-- name: DailyCostForApp :many
SELECT
    id, env, team, app, cost_type, date, daily_cost
FROM
    cost
WHERE
    date >= $4::date
    AND date <= $5::date
    AND env = $1
    AND team = $2
    AND app = $3
ORDER BY
    date, cost_type ASC
`

type DailyCostForAppParams struct {
	Env      *string
	Team     *string
	App      string
	FromDate pgtype.Date
	ToDate   pgtype.Date
}

// DailyCostForApp will fetch the daily cost for a specific team app in a specific environment, across all cost types
// in a date range.
func (q *Queries) DailyCostForApp(ctx context.Context, arg DailyCostForAppParams) ([]*Cost, error) {
	rows, err := q.db.Query(ctx, dailyCostForApp,
		arg.Env,
		arg.Team,
		arg.App,
		arg.FromDate,
		arg.ToDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Cost
	for rows.Next() {
		var i Cost
		if err := rows.Scan(
			&i.ID,
			&i.Env,
			&i.Team,
			&i.App,
			&i.CostType,
			&i.Date,
			&i.DailyCost,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const dailyCostForTeam = `-- name: DailyCostForTeam :many
SELECT
    id, env, team, app, cost_type, date, daily_cost
FROM
    cost
WHERE
    date >= $2::date
    AND date <= $3::date
    AND team = $1
ORDER BY
    date, env, app, cost_type ASC
`

type DailyCostForTeamParams struct {
	Team     *string
	FromDate pgtype.Date
	ToDate   pgtype.Date
}

// DailyCostForTeam will fetch the daily cost for a specific team across all apps and envs in a date range.
func (q *Queries) DailyCostForTeam(ctx context.Context, arg DailyCostForTeamParams) ([]*Cost, error) {
	rows, err := q.db.Query(ctx, dailyCostForTeam, arg.Team, arg.FromDate, arg.ToDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Cost
	for rows.Next() {
		var i Cost
		if err := rows.Scan(
			&i.ID,
			&i.Env,
			&i.Team,
			&i.App,
			&i.CostType,
			&i.Date,
			&i.DailyCost,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const dailyEnvCostForTeam = `-- name: DailyEnvCostForTeam :many
SELECT
    team,
    app,
    date,
    SUM(daily_cost)::real AS daily_cost
FROM
    cost
WHERE
    date >= $3::date
    AND date <= $4::date
    AND env = $1
    AND team = $2
GROUP BY
    team, app, date
ORDER BY
    date, app ASC
`

type DailyEnvCostForTeamParams struct {
	Env      *string
	Team     *string
	FromDate pgtype.Date
	ToDate   pgtype.Date
}

type DailyEnvCostForTeamRow struct {
	Team      *string
	App       string
	Date      pgtype.Date
	DailyCost float32
}

// DailyEnvCostForTeam will fetch the daily cost for a specific team and env across all apps in a date range.
func (q *Queries) DailyEnvCostForTeam(ctx context.Context, arg DailyEnvCostForTeamParams) ([]*DailyEnvCostForTeamRow, error) {
	rows, err := q.db.Query(ctx, dailyEnvCostForTeam,
		arg.Env,
		arg.Team,
		arg.FromDate,
		arg.ToDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*DailyEnvCostForTeamRow
	for rows.Next() {
		var i DailyEnvCostForTeamRow
		if err := rows.Scan(
			&i.Team,
			&i.App,
			&i.Date,
			&i.DailyCost,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const lastCostDate = `-- name: LastCostDate :one
SELECT
    MAX(date)::date AS date
FROM
    cost
`

// LastCostDate will return the last date that has a cost.
func (q *Queries) LastCostDate(ctx context.Context) (pgtype.Date, error) {
	row := q.db.QueryRow(ctx, lastCostDate)
	var date pgtype.Date
	err := row.Scan(&date)
	return date, err
}

const monthlyCostForApp = `-- name: MonthlyCostForApp :many
WITH last_run AS (
    SELECT MAX(date)::date AS "last_run"
    FROM cost
)
SELECT 
    team, 
    app, 
    env, 
    date_trunc('month', date)::date AS month,
    -- Extract last day of known cost samples for the month, or the last recorded date
    -- This helps with estimation etc
    MAX(CASE 
        WHEN date_trunc('month', date) < date_trunc('month', last_run) THEN date_trunc('month', date) + interval '1 month' - interval '1 day'
        ELSE date_trunc('day', last_run)
    END)::date AS last_recorded_date,
    SUM(daily_cost)::real AS daily_cost
FROM cost c 
LEFT JOIN last_run ON true
WHERE c.team = $1
AND c.app = $2
AND c.env = $3
GROUP BY team, app, env, month
ORDER BY month DESC
LIMIT 12
`

type MonthlyCostForAppParams struct {
	Team *string
	App  string
	Env  *string
}

type MonthlyCostForAppRow struct {
	Team             *string
	App              string
	Env              *string
	Month            pgtype.Date
	LastRecordedDate pgtype.Date
	DailyCost        float32
}

func (q *Queries) MonthlyCostForApp(ctx context.Context, arg MonthlyCostForAppParams) ([]*MonthlyCostForAppRow, error) {
	rows, err := q.db.Query(ctx, monthlyCostForApp, arg.Team, arg.App, arg.Env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*MonthlyCostForAppRow
	for rows.Next() {
		var i MonthlyCostForAppRow
		if err := rows.Scan(
			&i.Team,
			&i.App,
			&i.Env,
			&i.Month,
			&i.LastRecordedDate,
			&i.DailyCost,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const monthlyCostForTeam = `-- name: MonthlyCostForTeam :many
WITH last_run AS (
    SELECT MAX(date)::date AS "last_run"
    FROM cost
)
SELECT 
    team, 
    date_trunc('month', date)::date AS month,
    -- Extract last day of known cost samples for the month, or the last recorded date
    -- This helps with estimation etc
    MAX(CASE 
        WHEN date_trunc('month', date) < date_trunc('month', last_run) THEN date_trunc('month', date) + interval '1 month' - interval '1 day'
        ELSE date_trunc('day', last_run)
    END)::date AS last_recorded_date,
    SUM(daily_cost)::real AS daily_cost
FROM cost c
LEFT JOIN last_run ON true
WHERE c.team = $1
GROUP BY team, month
ORDER BY month DESC
LIMIT 12
`

type MonthlyCostForTeamRow struct {
	Team             *string
	Month            pgtype.Date
	LastRecordedDate pgtype.Date
	DailyCost        float32
}

func (q *Queries) MonthlyCostForTeam(ctx context.Context, team *string) ([]*MonthlyCostForTeamRow, error) {
	rows, err := q.db.Query(ctx, monthlyCostForTeam, team)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*MonthlyCostForTeamRow
	for rows.Next() {
		var i MonthlyCostForTeamRow
		if err := rows.Scan(
			&i.Team,
			&i.Month,
			&i.LastRecordedDate,
			&i.DailyCost,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const truncateCostTable = `-- name: TruncateCostTable :exec
TRUNCATE TABLE cost
`

// TruncateCostTable will truncate the cost table before doing a complete reimport.
func (q *Queries) TruncateCostTable(ctx context.Context) error {
	_, err := q.db.Exec(ctx, truncateCostTable)
	return err
}
