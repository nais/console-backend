-- +goose Up
ALTER TABLE cost RENAME COLUMN cost TO daily_cost;

-- +goose Down
ALTER TABLE cost RENAME COLUMN daily_cost TO cost;