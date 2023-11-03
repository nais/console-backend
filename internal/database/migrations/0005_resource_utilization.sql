-- +goose Up
CREATE TYPE resource_type AS ENUM (
    'cpu',
    'memory'
);

CREATE TABLE lowres_resource_utilization_metrics (
    id serial PRIMARY KEY,
    date date NOT NULL,
    env text,
    team text,
    app text,
    resource_type resource_type NOT NULL,
    usage real NOT NULL,
    request real NOT NULL,
    CONSTRAINT lowres_metric UNIQUE (date, env, team, app, resource_type)
);

CREATE TABLE highres_resource_utilization_metrics (
    id serial PRIMARY KEY,
    date date NOT NULL,
    env text,
    team text,
    app text,
    resource_type resource_type NOT NULL,
    usage real NOT NULL,
    request real NOT NULL,
    CONSTRAINT highres_metric UNIQUE (date, env, team, app, resource_type)
);

-- +goose Down
DROP TABLE lowres_resource_utilization_metrics, highres_resource_utilization_metrics;
DROP TYPE resource_type;