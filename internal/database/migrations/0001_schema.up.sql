-- +goose Up
CREATE TABLE cost (
    id serial PRIMARY KEY,
    env text,
    team text,
    app text,
    cost_type text NOT NULL,
    date DATE NOT NULL,
    cost real NOT NULL,
    UNIQUE NULLS NOT DISTINCT (env, team, app, cost_type, date)
);

