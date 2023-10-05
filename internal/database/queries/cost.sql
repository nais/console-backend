-- name: CostLastDate :one
SELECT MAX(date)::date AS "date"
FROM cost;

-- name: MonthlyCostForApp :many
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
LIMIT 12;

-- name: EnvCostForTeam :many
SELECT
    team,
    app,
    date,
    SUM(daily_cost)::real AS daily_cost
FROM cost
WHERE
    date >= sqlc.arg('from_date')::date
    AND date <= sqlc.arg('to_date')::date
    AND env = $1
    AND team = $2
GROUP by team, app, date
ORDER BY date, app ASC;

-- name: MonthlyCostForTeam :many
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
LIMIT 12;

-- name: CostUpsert :batchexec
INSERT INTO cost (env, team, app, cost_type, date, daily_cost)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (env, team, app, cost_type, date) DO
    UPDATE SET daily_cost = EXCLUDED.daily_cost;

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

-- DailyCostForTeam will fetch the daily cost for a specific team across all apps and envs in a date range.
-- name: DailyCostForTeam :many
SELECT
    *
FROM
    cost
WHERE
    date >= sqlc.arg('from_date')::date
    AND date <= sqlc.arg('to_date')::date
    AND team = $1
ORDER BY
    date, app, env ASC;