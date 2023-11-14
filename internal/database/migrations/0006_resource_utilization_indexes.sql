-- +goose Up
CREATE INDEX ON resource_utilization_metrics (app);
CREATE INDEX ON resource_utilization_metrics (env);
CREATE INDEX ON resource_utilization_metrics (resource_type);
CREATE INDEX ON resource_utilization_metrics (team);
CREATE INDEX ON resource_utilization_metrics (timestamp);

-- +goose Down
DROP INDEX resource_utilization_metrics_app_idx;
DROP INDEX resource_utilization_metrics_env_idx;
DROP INDEX resource_utilization_metrics_resource_type_idx;
DROP INDEX resource_utilization_metrics_team_idx;
DROP INDEX resource_utilization_metrics_timestamp_idx;
