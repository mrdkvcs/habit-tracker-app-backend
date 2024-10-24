-- +goose Up
CREATE TABLE team_user_roles (
    id UUID PRIMARY KEY NOT NULL,
    team_membership_id UUID REFERENCES team_memberships(id) ON DELETE CASCADE NOT NULL,
    role_id UUID REFERENCES team_roles(id) ON DELETE CASCADE NOT NULL,
    UNIQUE(team_membership_id, role_id)
);

-- +goose Down

DROP TABLE team_user_roles;
