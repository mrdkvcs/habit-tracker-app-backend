-- +goose Up
ALTER TABLE activities
ADD CONSTRAINT fk_activities_all_activities_id
FOREIGN KEY (activity_id) REFERENCES all_activities(id) ON DELETE CASCADE;

-- +goose Down

ALTER TABLE activities
DROP CONSTRAINT fk_activities_all_activities_id;

