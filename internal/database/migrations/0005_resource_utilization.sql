-- +goose Up
CREATE TYPE resource_type AS ENUM (
    'cpu',
    'memory'
);

CREATE TABLE utilization_lowres (
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

CREATE TABLE utilization_highres (
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
DROP TABLE resource_utilization_lowres, resource_utilization_highres;
DROP TYPE resource_type;