-- +goose Up
CREATE TABLE cost (
    PRIMARY KEY (tenant, env, app, cost_type, date),
    env text,
    team text,
    app text,
    cost_type text NOT NULL,
    date DATE NOT NULL,
    cost real NOT NULL
);
