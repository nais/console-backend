-- ResourceUtilizationDateRangeForTeam will return the min and max timestamps for a specific team.
-- name: ResourceUtilizationDateRangeForTeam :one
SELECT
    MIN(timestamp)::timestamptz AS "from",
    MAX(timestamp)::timestamptz AS "to"
FROM
    resource_utilization_metrics
WHERE
    team = $1;

-- ResourceUtilizationDateRange will return the min and max timestamps for an app/job.
-- name: ResourceUtilizationDateRange :one
SELECT
    MIN(timestamp)::timestamptz AS "from",
    MAX(timestamp)::timestamptz AS "to"
FROM
    resource_utilization_metrics
WHERE
    env = $1
    AND team = $2
    AND name = $3
    AND kind = $4;

-- ResourceUtilizationOverageCostForTeam will return overage records for a given team.
-- name: ResourceUtilizationOverageCostForTeam :many
SELECT
    SUM(usage)::double precision AS usage,
    SUM(request)::double precision AS request,
    env,
    name,
    kind,
    resource_type
FROM
    resource_utilization_metrics
WHERE
    team = $1
    AND timestamp >= sqlc.arg('start')::timestamptz
    AND timestamp < sqlc.arg('end')::timestamptz
GROUP BY
    env, name, kind, resource_type
HAVING
    SUM(request) > SUM(usage);


-- ResourceUtilizationUpsert will insert or update resource utilization records.
-- name: ResourceUtilizationUpsert :batchexec
INSERT INTO resource_utilization_metrics (timestamp, env, team, name, kind, resource_type, usage, request)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT ON CONSTRAINT resource_utilization_metric DO NOTHING;

-- MaxResourceUtilizationDate will return the max date for resource utilization records.
-- name: MaxResourceUtilizationDate :one
SELECT MAX(timestamp)::timestamptz FROM resource_utilization_metrics;

-- ResourceUtilizationForApp will return resource utilization records for a given app.
-- name: ResourceUtilizationForApp :many
SELECT
    *
FROM
    resource_utilization_metrics
WHERE
    env = $1
    AND team = $2
    AND name = $3
    AND kind = $4
    AND resource_type = $5
    AND timestamp >= sqlc.arg('start')::timestamptz
    AND timestamp < sqlc.arg('end')::timestamptz
ORDER BY
    timestamp ASC;

-- ResourceUtilizationForTeam will return resource utilization records for a given team.
-- name: ResourceUtilizationForTeam :many
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
    AND timestamp >= sqlc.arg('start')::timestamptz
    AND timestamp < sqlc.arg('end')::timestamptz
GROUP BY
    timestamp
ORDER BY
    timestamp ASC;
