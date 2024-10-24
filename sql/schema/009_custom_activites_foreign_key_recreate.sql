-- +goose Up
ALTER TABLE custom_activities
ADD CONSTRAINT fk_custom_activities_all_activities_id
FOREIGN KEY (activity_id) REFERENCES all_activities(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE custom_activities DROP CONSTRAINT fk_custom_activities_all_activities_id;
