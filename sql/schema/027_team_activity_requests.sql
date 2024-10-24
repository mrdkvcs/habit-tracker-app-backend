-- +goose Up
CREATE TABLE team_activity_requests (
  id UUID PRIMARY KEY NOT NULL ,
  team_id UUID NOT NULL REFERENCES teams(id),
  user_id UUID NOT NULL REFERENCES users(id),
  activity_name TEXT NOT NULL,
  points INTEGER NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
-- +goose Down
DROP TABLE team_activity_requests;
