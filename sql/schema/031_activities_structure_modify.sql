-- +goose Up

CREATE TABLE user_activities (
  id UUID PRIMARY KEY NOT NULL,
  user_id UUID REFERENCES users(id) NOT NULL,
  name VARCHAR(255) NOT NULL,
  points INT NOT NULL,
  activity_type activity_type NOT NULL,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP DEFAULT NOW() NOT NULL
);

CREATE TABLE user_activity_logs (
  id UUID NOT NULL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  activity_id UUID NOT NULL  REFERENCES user_activities(id) ON DELETE CASCADE,
  duration INTEGER NOT NULL,
  points INTEGER NOT NULL,
  activity_description TEXT NOT NULL,
  logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);


-- +goose Down
DROP TABLE user_activity_logs;
DROP TABLE user_activities;
