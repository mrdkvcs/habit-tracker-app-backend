-- +goose Up
ALTER TABLE custom_activities
DROP CONSTRAINT IF EXISTS custom_activities_activity_id_fkey ;

