-- +goose Up
CREATE TABLE suggest_feature (
  ID UUID PRIMARY KEY NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  username TEXT NOT NULL,
  upvote INTEGER DEFAULT 0 NOT NULL 
);

-- +goose Down
DROP TABLE suggest_feature;

