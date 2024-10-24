-- +goose Up
ALTER TABLE activity_logs
ADD COLUMN start_time TIMESTAMP WITH TIME ZONE NOT NULL,
ADD COLUMN end_time TIMESTAMP WITH TIME ZONE NOT NULL;

-- +goose Down

ALTER TABLE activity_logs
DROP COLUMN start_time,
DROP COLUMN end_time;
