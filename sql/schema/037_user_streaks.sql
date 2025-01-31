-- +goose Up
CREATE TABLE user_streaks (
  user_id UUID PRIMARY KEY NOT NULL  REFERENCES users(id) ON DELETE CASCADE,
  current_streak INT DEFAULT 0 NOT NULL,
  longest_streak INT DEFAULT 0 NOT NULL,
  last_logged_date DATE   DEFAULT NULL 
);

-- +goose Down

DROP TABLE user_streaks;
