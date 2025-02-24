-- +goose Up

ALTER TABLE user_activities 
ALTER COLUMN activity_type TYPE TEXT USING activity_type::TEXT;


