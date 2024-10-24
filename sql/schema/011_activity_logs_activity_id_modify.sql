-- +goose Up
ALTER TABLE activity_logs
ALTER COLUMN  activity_id SET NOT NULL;
