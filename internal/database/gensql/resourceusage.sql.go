// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: resourceusage.sql

package gensql

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const maxResourceUtilizationDate = `-- name: MaxResourceUtilizationDate :one
SELECT MAX(timestamp)::timestamptz FROM resource_utilization_metrics
`

// MaxResourceUtilizationDate will return the max date for resource utilization records.
func (q *Queries) MaxResourceUtilizationDate(ctx context.Context) (pgtype.Timestamptz, error) {
	row := q.db.QueryRow(ctx, maxResourceUtilizationDate)
	var column_1 pgtype.Timestamptz
	err := row.Scan(&column_1)
	return column_1, err
}

const resourceUtilizationForApp = `-- name: ResourceUtilizationForApp :many
SELECT
    id, timestamp, env, team, app, resource_type, usage, request
FROM
    resource_utilization_metrics
WHERE
    env = $1
    AND team = $2
    AND app = $3
    AND resource_type = $4
    AND timestamp >= $5::timestamptz
    AND timestamp <= $6::timestamptz
ORDER BY
    timestamp ASC
`

type ResourceUtilizationForAppParams struct {
	Env          string
	Team         string
	App          string
	ResourceType ResourceType
	Start        pgtype.Timestamptz
	End          pgtype.Timestamptz
}

// ResourceUtilizationForApp will return resource utilization records for a given app.
func (q *Queries) ResourceUtilizationForApp(ctx context.Context, arg ResourceUtilizationForAppParams) ([]*ResourceUtilizationMetric, error) {
	rows, err := q.db.Query(ctx, resourceUtilizationForApp,
		arg.Env,
		arg.Team,
		arg.App,
		arg.ResourceType,
		arg.Start,
		arg.End,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ResourceUtilizationMetric
	for rows.Next() {
		var i ResourceUtilizationMetric
		if err := rows.Scan(
			&i.ID,
			&i.Timestamp,
			&i.Env,
			&i.Team,
			&i.App,
			&i.ResourceType,
			&i.Usage,
			&i.Request,
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

const resourceUtilizationForTeam = `-- name: ResourceUtilizationForTeam :many
SELECT
    SUM(usage)::double precision AS usage,
    SUM(request)::double precision AS request,
    timestamp
FROM
    resource_utilization_metrics
WHERE
    env = $1
    AND team = $2
    AND resource_type = $3
    AND timestamp >= $4::timestamptz
    AND timestamp <= $5::timestamptz
GROUP BY
    timestamp
ORDER BY
    timestamp ASC
`

type ResourceUtilizationForTeamParams struct {
	Env          string
	Team         string
	ResourceType ResourceType
	Start        pgtype.Timestamptz
	End          pgtype.Timestamptz
}

type ResourceUtilizationForTeamRow struct {
	Usage     float64
	Request   float64
	Timestamp pgtype.Timestamptz
}

// ResourceUtilizationForTeam will return resource utilization records for a given team.
func (q *Queries) ResourceUtilizationForTeam(ctx context.Context, arg ResourceUtilizationForTeamParams) ([]*ResourceUtilizationForTeamRow, error) {
	rows, err := q.db.Query(ctx, resourceUtilizationForTeam,
		arg.Env,
		arg.Team,
		arg.ResourceType,
		arg.Start,
		arg.End,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ResourceUtilizationForTeamRow
	for rows.Next() {
		var i ResourceUtilizationForTeamRow
		if err := rows.Scan(&i.Usage, &i.Request, &i.Timestamp); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
