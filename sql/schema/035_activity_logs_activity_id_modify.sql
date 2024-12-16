-- +goose Up
ALTER TABLE user_activity_logs
ALTER COLUMN  activity_id DROP NOT NULL;

