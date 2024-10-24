-- +goose Up
ALTER TABLE activities
DROP CONSTRAINT fk_activities_all_activities_id;

