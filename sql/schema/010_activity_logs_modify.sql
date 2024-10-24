-- +goose Up
ALTER TABLE activity_logs DROP COLUMN activity_type;


-- +goose Down
ALTER TABLE activity_logs ADD COLUMN activity_type VARCHAR(255);
