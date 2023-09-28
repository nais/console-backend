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
WHERE date >= $4::DATE AND date <= $5::DATE AND env = $1 AND team = $2 AND app = $3
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
SELECT MAX(date)::DATE AS "date"
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

const totalCostForApp = `-- name: TotalCostForApp :many
SELECT app, env, team, cost_type, SUM(cost)::REAL as cost FROM cost
WHERE date >= $4::DATE AND date <= $5::DATE AND env = $1 AND team = $2 AND app = $3
GROUP by team, app, cost_type
`

type TotalCostForAppParams struct {
	Env      *string
	Team     *string
	App      *string
	FromDate pgtype.Date
	ToDate   pgtype.Date
}

type TotalCostForAppRow struct {
	App      *string
	Env      *string
	Team     *string
	CostType string
	Cost     float32
}

func (q *Queries) TotalCostForApp(ctx context.Context, arg TotalCostForAppParams) ([]*TotalCostForAppRow, error) {
	rows, err := q.db.Query(ctx, totalCostForApp,
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
	var items []*TotalCostForAppRow
	for rows.Next() {
		var i TotalCostForAppRow
		if err := rows.Scan(
			&i.App,
			&i.Env,
			&i.Team,
			&i.CostType,
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

const totalCostForTeam = `-- name: TotalCostForTeam :many
SELECT app, env, team, cost_type, SUM(cost)::REAL as cost FROM cost
WHERE date >= $3::DATE AND date <= $4::DATE AND env = $1 AND team = $2
GROUP by team, app, cost_type
`

type TotalCostForTeamParams struct {
	Env      *string
	Team     *string
	FromDate pgtype.Date
	ToDate   pgtype.Date
}

type TotalCostForTeamRow struct {
	App      *string
	Env      *string
	Team     *string
	CostType string
	Cost     float32
}

func (q *Queries) TotalCostForTeam(ctx context.Context, arg TotalCostForTeamParams) ([]*TotalCostForTeamRow, error) {
	rows, err := q.db.Query(ctx, totalCostForTeam,
		arg.Env,
		arg.Team,
		arg.FromDate,
		arg.ToDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*TotalCostForTeamRow
	for rows.Next() {
		var i TotalCostForTeamRow
		if err := rows.Scan(
			&i.App,
			&i.Env,
			&i.Team,
			&i.CostType,
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
