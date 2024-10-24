-- +goose Up
ALTER TABLE activity_logs
DROP COLUMN start_time,
DROP COLUMN end_time;



