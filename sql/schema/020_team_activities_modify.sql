-- +goose Up
ALTER TABLE team_activities DROP COLUMN IF EXISTS activity_role;
ALTER TABLE team_activities ADD COLUMN activity_roles TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE team_activities DROP COLUMN IF EXISTS activity_roles;
