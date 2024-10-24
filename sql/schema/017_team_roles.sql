-- +goose Up
CREATE TABLE team_roles (
id UUID PRIMARY KEY,
role_name TEXT NOT NULL,
team_id UUID REFERENCES teams (id),
UNIQUE (role_name, team_id)
);

-- +goose Down
DROP TABLE team_roles;
