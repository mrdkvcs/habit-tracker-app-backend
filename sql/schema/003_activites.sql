-- +goose Up
CREATE TABLE activities (
  activity_id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  points INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  id_seged_2 SERIAL
);

INSERT INTO activities (name, points) VALUES
('Exercise', 50),
('Reading', 30),
('Meditation', 40),
('Learning', 60),
('Household Chores', 20),
('Watching Series' , -10),
('Watching TV', -20),
('Gaming', -25),
('Social Media Scrolling', -30),
('Watching adult websites ', -40);

UPDATE activities AS a
SET activity_id = aa.id
FROM all_activities AS aa
WHERE a.id_seged_2 = aa.id_seged_1;

ALTER TABLE activities
ADD CONSTRAINT fk_activities_all_activities_id
FOREIGN KEY (activity_id) REFERENCES all_activities(id);

ALTER TABLE activities DROP COLUMN id_seged_2;
ALTER TABLE all_activities DROP COLUMN id_seged_1;

-- +goose Down
DROP TABLE activities;

