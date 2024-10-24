-- +goose Up
CREATE TABLE users (
  ID UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  username TEXT NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  api_key VARCHAR(64) NOT NULL UNIQUE
);
-- +goose Down
DROP TABLE users;
