-- +goose Up
ALTER TABLE activity_logs ADD COLUMN activity_description TEXT NOT NULL;

-- +goose Down
ALTER TABLE activity_logs DROP COLUMN activity_description;
