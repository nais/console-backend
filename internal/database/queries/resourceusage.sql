-- ResourceUtilizationUpsert will insert or update resource utilization records.
-- name: ResourceUtilizationUpsert :batchexec
INSERT INTO resource_utilization_metrics (timestamp, env, team, app, resource_type, usage, request)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
    AND app = $3
    AND resource_type = $4
    AND timestamp >= sqlc.arg('start')::timestamptz
    AND timestamp <= sqlc.arg('end')::timestamptz
ORDER BY
    timestamp ASC;
