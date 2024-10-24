-- +goose Up

CREATE TABLE team_memberships (
  id UUID PRIMARY KEY NOT NULL , 
  team_id UUID REFERENCES teams(id) ON DELETE CASCADE NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
  role VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
  UNIQUE(team_id, user_id)
);

-- +goose Down

DROP TABLE team_memberships;
