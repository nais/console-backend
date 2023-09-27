-- name: GetCost :many
SELECT * FROM cost;

-- name: CostLastDate :one
SELECT MAX(date)::DATE AS "date"
FROM cost
;

-- name: CostUpsert :batchexec
INSERT INTO cost (
  env,
  team,
  app,
  cost_type,
  date,
  cost
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
) ON CONFLICT (env, team, app, cost_type, date) DO UPDATE SET
  cost = EXCLUDED.cost
;