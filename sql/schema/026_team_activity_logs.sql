-- +goose Up
CREATE TABLE team_activity_logs (
  id UUID PRIMARY KEY , 
  team_id UUID NOT NULL REFERENCES teams(id),
  user_id UUID NOT NULL REFERENCES users(id),
  activity_name  TEXT NOT NULL,
  points INTEGER NOT NULL, 
  logged_at TIMESTAMP WITH TIME ZONE NOT NULL
);
-- +goose Down
DROP TABLE team_activity_logs;
