-- +goose Up

CREATE TABLE teams (
  id UUID PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  team_industry TEXT NOT NULL,
  team_size INT NOT NULL,
  is_private BOOLEAN NOT NULL,
  created_by UUID REFERENCES users(id) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- +goose Down
DROP TABLE teams;
