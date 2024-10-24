-- +goose Up

CREATE TABLE activity_logs (
  id UUID NOT NULL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  activity_id UUID references all_activities(id) ON DELETE CASCADE,
  activity_type VARCHAR(10) NOT NULL,
  duration INTEGER NOT NULL,
  points INTEGER NOT NULL,
  logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);


-- +goose Down
DROP TABLE activity_logs;

