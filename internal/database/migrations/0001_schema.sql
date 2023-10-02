-- +goose Up
CREATE TABLE cost (
    id serial PRIMARY KEY,
    env text,
    team text,
    app text,
    cost_type text NOT NULL,
    date date NOT NULL,
    cost real NOT NULL,
    UNIQUE NULLS NOT DISTINCT (env, team, app, cost_type, date)
);

-- +goose Down
DROP TABLE cost;