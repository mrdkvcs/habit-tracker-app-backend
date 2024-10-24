-- +goose Up
CREATE TABLE all_activities (
  id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
  type TEXT NOT NULL,
  id_seged_1 SERIAL
);

INSERT INTO all_activities (type)
VALUES
    ('default'),
    ('default'),
    ('default'),
    ('default'),
    ('default'),
    ('default'),
    ('default'),
    ('default'),
    ('default'),
    ('default');
-- +goose Down
DROP  TABLE all_activities;

