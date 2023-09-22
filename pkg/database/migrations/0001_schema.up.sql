-- +goose Up
CREATE TABLE cost (
    env text,
    team text,
    app text,
    service_description text NOT NULL,
    dato DATE NOT NULL,
    cost real NOT NULL
);
