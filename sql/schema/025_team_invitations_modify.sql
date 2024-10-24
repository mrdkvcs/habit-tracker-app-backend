-- +goose Up
ALTER TABLE team_invitations ADD column seen BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE team_invitations DROP column seen;
