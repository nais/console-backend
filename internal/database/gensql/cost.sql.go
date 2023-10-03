// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: cost.sql

package gensql

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const costForApp = `-- name: CostForApp :many
SELECT id, env, team, app, cost_type, date, cost FROM cost
WHERE
    date >= $4::date
    AND date <= $5::date
    AND env = $1
    AND team = $2
    AND app = $3
GROUP by id, team, app, cost_type, date
ORDER BY date ASC
`

type CostForAppParams struct {
	Env      *string
	Team     *string
	App      *string
	FromDate pgtype.Date
	ToDate   pgtype.Date
}

func (q *Queries) CostForApp(ctx context.Context, arg CostForAppParams) ([]*Cost, error) {
	rows, err := q.db.Query(ctx, costForApp,
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
			&i.Cost,
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

const costLastDate = `-- name: CostLastDate :one
SELECT MAX(date)::date AS "date"
FROM cost
`

func (q *Queries) CostLastDate(ctx context.Context) (pgtype.Date, error) {
	row := q.db.QueryRow(ctx, costLastDate)
	var date pgtype.Date
	err := row.Scan(&date)
	return date, err
}

const getCost = `-- name: GetCost :many
SELECT id, env, team, app, cost_type, date, cost FROM cost
`

func (q *Queries) GetCost(ctx context.Context) ([]*Cost, error) {
	rows, err := q.db.Query(ctx, getCost)
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
			&i.Cost,
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
SELECT team, app, env, date_trunc('month', date)::date AS month, SUM(cost)::real AS cost FROM cost
WHERE team = $1
AND app = $2
AND env = $3
GROUP BY team, app, env, month
ORDER BY month DESC
LIMIT 12
`

type MonthlyCostForTeamParams struct {
	Team *string
	App  *string
	Env  *string
}

type MonthlyCostForTeamRow struct {
	Team  *string
	App   *string
	Env   *string
	Month pgtype.Date
	Cost  float32
}

func (q *Queries) MonthlyCostForTeam(ctx context.Context, arg MonthlyCostForTeamParams) ([]*MonthlyCostForTeamRow, error) {
	rows, err := q.db.Query(ctx, monthlyCostForTeam, arg.Team, arg.App, arg.Env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*MonthlyCostForTeamRow
	for rows.Next() {
		var i MonthlyCostForTeamRow
		if err := rows.Scan(
			&i.Team,
			&i.App,
			&i.Env,
			&i.Month,
			&i.Cost,
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
