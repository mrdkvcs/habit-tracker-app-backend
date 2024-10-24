-- +goose Up

CREATE TABLE team_activities (
  id UUID PRIMARY KEY NOT NULL,
  team_id UUID REFERENCES teams(id) ON DELETE CASCADE NOT NULL,
  activity_name VARCHAR(255) NOT NULL,
  activity_role VARCHAR(150) NOT NULL,
  points INTEGER NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +goose Down
DROP TABLE team_activities;
