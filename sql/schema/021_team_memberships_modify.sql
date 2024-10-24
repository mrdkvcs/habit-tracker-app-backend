-- +goose Up
ALTER TABLE team_memberships DROP COLUMN IF EXISTS role_id;

