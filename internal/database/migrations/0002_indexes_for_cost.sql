-- +goose Up
CREATE INDEX cost_env_idx ON cost (env);
CREATE INDEX cost_team_idx ON cost (team);
CREATE INDEX cost_app_idx ON cost (app);
CREATE INDEX cost_date_idx ON cost (date);

-- +goose Down
DROP INDEX cost_env_idx;
DROP INDEX cost_team_idx;
DROP INDEX cost_app_idx;
DROP INDEX cost_date_idx;
