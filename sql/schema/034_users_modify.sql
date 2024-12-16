-- +goose Up
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ADD COLUMN google_id TEXT UNIQUE;
ALTER TABLE users DROP COLUMN api_key;


-- +goose Down
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE users DROP COLUMN google_id;
ALTER TABLE users ADD COLUMN api_key VARCHAR(64) UNIQUE;

