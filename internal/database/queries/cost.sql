-- LastCostDate will return the last date that has a cost.
-- name: LastCostDate :one
SELECT
    MAX(date)::date AS date
FROM
    cost;

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

-- CostUpsert will insert or update a cost record. If there is a conflict on the daily_cost_key constrant, the
-- daily_cost column will be updated.
-- name: CostUpsert :batchexec
INSERT INTO cost (env, team, app, cost_type, date, daily_cost)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT ON CONSTRAINT daily_cost_key DO
    UPDATE SET daily_cost = EXCLUDED.daily_cost;

-- DailyCostForApp will fetch the daily cost for a specific team app in a specific environment, across all cost types
-- in a date range.
-- name: DailyCostForApp :many
SELECT
    *
FROM
    cost
WHERE
    date >= sqlc.arg('from_date')::date
    AND date <= sqlc.arg('to_date')::date
    AND env = $1
    AND team = $2
    AND app = $3
ORDER BY
    date, cost_type ASC;

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
    date, env, app, cost_type ASC;

-- DailyEnvCostForTeam will fetch the daily cost for a specific team and env across all apps in a date range.
-- name: DailyEnvCostForTeam :many
SELECT
    team,
    app,
    date,
    SUM(daily_cost)::real AS daily_cost
FROM
    cost
WHERE
    date >= sqlc.arg('from_date')::date
    AND date <= sqlc.arg('to_date')::date
    AND env = $1
    AND team = $2
GROUP BY
    team, app, date
ORDER BY
    date, app ASC;

-- TruncateCostTable will truncate the cost table before doing a complete reimport.
-- name: TruncateCostTable :exec
TRUNCATE TABLE cost;