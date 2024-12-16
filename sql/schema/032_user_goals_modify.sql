-- +goose Up

ALTER TABLE user_goals ADD COLUMN IF NOT EXISTS status TEXT  DEFAULT 'not completed' NOT NULL;

-- +goose Down

ALTER TABLE user_goals DROP COLUMN IF EXISTS completed;
