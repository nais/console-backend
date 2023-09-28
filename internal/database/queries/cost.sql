-- name: GetCost :many
SELECT * FROM cost;

-- name: CostLastDate :one
SELECT MAX(date)::DATE AS "date"
FROM cost;

-- name: CostUpsert :batchexec
INSERT INTO cost (env, team, app, cost_type, date, cost)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (env, team, app, cost_type, date) DO
    UPDATE SET cost = EXCLUDED.cost;

-- name: TotalCostForApp :many
SELECT app, env, team, cost_type, SUM(cost)::REAL as cost FROM cost
WHERE date >= sqlc.arg('from_date')::DATE AND date <= sqlc.arg('to_date')::DATE AND env = $1 AND team = $2 AND app = $3
GROUP by team, app, cost_type;

-- name: CostForApp :many
SELECT * FROM cost
WHERE date >= sqlc.arg('from_date')::DATE AND date <= sqlc.arg('to_date')::DATE AND env = $1 AND team = $2 AND app = $3
GROUP by id, team, app, cost_type, date
ORDER BY date ASC;

-- name: TotalCostForTeam :many
SELECT app, env, team, cost_type, SUM(cost)::REAL as cost FROM cost
WHERE date >= sqlc.arg('from_date')::DATE AND date <= sqlc.arg('to_date')::DATE AND env = $1 AND team = $2
GROUP by team, app, cost_type;

