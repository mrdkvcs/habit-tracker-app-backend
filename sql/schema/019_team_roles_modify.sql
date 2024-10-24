-- +goose Up
ALTER TABLE team_roles
ALTER COLUMN team_id SET NOT NULL;
