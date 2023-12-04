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
    AND timestamp < $6::timestamptz
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
    AND timestamp < $5::timestamptz
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

const resourceUtilizationOverageForTeam = `-- name: ResourceUtilizationOverageForTeam :many
SELECT
    usage,
    request,
    app,
    env,
    (request-usage)::double precision AS overage
FROM
    resource_utilization_metrics
WHERE
    team = $1
    AND timestamp = $2
    AND resource_type = $3
GROUP BY
    app, env, usage, request, timestamp
ORDER BY
    overage DESC
`

type ResourceUtilizationOverageForTeamParams struct {
	Team         string
	Timestamp    pgtype.Timestamptz
	ResourceType ResourceType
}

type ResourceUtilizationOverageForTeamRow struct {
	Usage   float64
	Request float64
	App     string
	Env     string
	Overage float64
}

// ResourceUtilizationOverageForTeam will return overage records for a given team, ordered by overage descending.
func (q *Queries) ResourceUtilizationOverageForTeam(ctx context.Context, arg ResourceUtilizationOverageForTeamParams) ([]*ResourceUtilizationOverageForTeamRow, error) {
	rows, err := q.db.Query(ctx, resourceUtilizationOverageForTeam, arg.Team, arg.Timestamp, arg.ResourceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ResourceUtilizationOverageForTeamRow
	for rows.Next() {
		var i ResourceUtilizationOverageForTeamRow
		if err := rows.Scan(
			&i.Usage,
			&i.Request,
			&i.App,
			&i.Env,
			&i.Overage,
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

const resourceUtilizationRangeForApp = `-- name: ResourceUtilizationRangeForApp :one
SELECT
    MIN(timestamp)::timestamptz AS "from",
    MAX(timestamp)::timestamptz AS "to"
FROM
    resource_utilization_metrics
WHERE
    env = $1
    AND team = $2
    AND app = $3
`

type ResourceUtilizationRangeForAppParams struct {
	Env  string
	Team string
	App  string
}

type ResourceUtilizationRangeForAppRow struct {
	From pgtype.Timestamptz
	To   pgtype.Timestamptz
}

// ResourceUtilizationRangeForApp will return the min and max timestamps for a specific app.
func (q *Queries) ResourceUtilizationRangeForApp(ctx context.Context, arg ResourceUtilizationRangeForAppParams) (*ResourceUtilizationRangeForAppRow, error) {
	row := q.db.QueryRow(ctx, resourceUtilizationRangeForApp, arg.Env, arg.Team, arg.App)
	var i ResourceUtilizationRangeForAppRow
	err := row.Scan(&i.From, &i.To)
	return &i, err
}

const resourceUtilizationRangeForTeam = `-- name: ResourceUtilizationRangeForTeam :one
SELECT
    MIN(timestamp)::timestamptz AS "from",
    MAX(timestamp)::timestamptz AS "to"
FROM
    resource_utilization_metrics
WHERE
    team = $1
`

type ResourceUtilizationRangeForTeamRow struct {
	From pgtype.Timestamptz
	To   pgtype.Timestamptz
}

// ResourceUtilizationRangeForTeam will return the min and max timestamps for a specific team.
func (q *Queries) ResourceUtilizationRangeForTeam(ctx context.Context, team string) (*ResourceUtilizationRangeForTeamRow, error) {
	row := q.db.QueryRow(ctx, resourceUtilizationRangeForTeam, team)
	var i ResourceUtilizationRangeForTeamRow
	err := row.Scan(&i.From, &i.To)
	return &i, err
}

const specificResourceUtilizationForApp = `-- name: SpecificResourceUtilizationForApp :one
SELECT
    usage,
    request,
    timestamp
FROM
    resource_utilization_metrics
WHERE
    env = $1
    AND team = $2
    AND app = $3
    AND resource_type = $4
    AND timestamp >= $5::timestamptz
    AND timestamp < $6::timestamptz
ORDER BY
    timestamp DESC
LIMIT
    1
`

type SpecificResourceUtilizationForAppParams struct {
	Env          string
	Team         string
	App          string
	ResourceType ResourceType
	Start        pgtype.Timestamptz
	End          pgtype.Timestamptz
}

type SpecificResourceUtilizationForAppRow struct {
	Usage     float64
	Request   float64
	Timestamp pgtype.Timestamptz
}

// SpecificResourceUtilizationForApp will return resource utilization for an app at a specific timestamp.
func (q *Queries) SpecificResourceUtilizationForApp(ctx context.Context, arg SpecificResourceUtilizationForAppParams) (*SpecificResourceUtilizationForAppRow, error) {
	row := q.db.QueryRow(ctx, specificResourceUtilizationForApp,
		arg.Env,
		arg.Team,
		arg.App,
		arg.ResourceType,
		arg.Start,
		arg.End,
	)
	var i SpecificResourceUtilizationForAppRow
	err := row.Scan(&i.Usage, &i.Request, &i.Timestamp)
	return &i, err
}

const specificResourceUtilizationForTeam = `-- name: SpecificResourceUtilizationForTeam :one
SELECT
    SUM(usage)::double precision AS usage,
    SUM(request)::double precision AS request,
    timestamp
FROM
    resource_utilization_metrics
WHERE
    team = $1
    AND resource_type = $2
    AND timestamp >= $3::timestamptz
    AND timestamp < $4::timestamptz
    AND request > usage
GROUP BY
    timestamp
ORDER BY
    timestamp DESC
LIMIT
    1
`

type SpecificResourceUtilizationForTeamParams struct {
	Team         string
	ResourceType ResourceType
	Start        pgtype.Timestamptz
	End          pgtype.Timestamptz
}

type SpecificResourceUtilizationForTeamRow struct {
	Usage     float64
	Request   float64
	Timestamp pgtype.Timestamptz
}

// SpecificResourceUtilizationForTeam will return resource utilization for a team at a specific timestamp. Applications
// with a usage greater than request will be ignored.
func (q *Queries) SpecificResourceUtilizationForTeam(ctx context.Context, arg SpecificResourceUtilizationForTeamParams) (*SpecificResourceUtilizationForTeamRow, error) {
	row := q.db.QueryRow(ctx, specificResourceUtilizationForTeam,
		arg.Team,
		arg.ResourceType,
		arg.Start,
		arg.End,
	)
	var i SpecificResourceUtilizationForTeamRow
	err := row.Scan(&i.Usage, &i.Request, &i.Timestamp)
	return &i, err
}
