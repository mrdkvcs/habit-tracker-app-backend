-- +goose Up
ALTER TABLE team_memberships DROP COLUMN role;
ALTER TABLE team_memberships ADD COLUMN role_id UUID REFERENCES team_roles(id) ON DELETE CASCADE NOT NULL;

-- +goose Down
ALTER TABLE team_memberships DROP COLUMN role_id;

