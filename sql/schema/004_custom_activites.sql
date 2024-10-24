-- +goose Up
CREATE TABLE custom_activities (
  activity_id UUID REFERENCES all_activities(id) NOT NULL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  name VARCHAR(255) NOT NULL,
  points INTEGER NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE custom_activities;

