-- name: GetCost :many
SELECT * FROM cost;

-- name: CostLastDate :one
SELECT MAX(date)::date AS "date"
FROM cost;

-- name: MonthlyCostForTeam :many
SELECT team, app, env, date_trunc('month', date)::date AS month, SUM(cost)::real AS cost FROM cost
WHERE team = $1
AND app = $2
AND env = $3
GROUP BY team, app, env, month
ORDER BY month DESC
LIMIT 12;

-- name: CostUpsert :batchexec
INSERT INTO cost (env, team, app, cost_type, date, cost)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (env, team, app, cost_type, date) DO
    UPDATE SET cost = EXCLUDED.cost;

-- name: CostForApp :many
SELECT * FROM cost
WHERE
    date >= sqlc.arg('from_date')::date
    AND date <= sqlc.arg('to_date')::date
    AND env = $1
    AND team = $2
    AND app = $3
GROUP by id, team, app, cost_type, date
ORDER BY date ASC;