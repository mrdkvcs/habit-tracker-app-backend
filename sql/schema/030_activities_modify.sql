-- +goose Up
UPDATE activities SET points = 10 WHERE name = 'Learning';
UPDATE activities SET points = 8 WHERE name = 'Exercise';
UPDATE activities SET points = 6 WHERE name = 'Meditation';
UPDATE activities SET points = 4 WHERE name = 'Reading';
UPDATE activities SET points = 2 WHERE name = 'Household Chores';
UPDATE activities SET points = -2 WHERE name = 'Watching Series';
UPDATE activities SET points = -4 WHERE name = 'Watching TV';
UPDATE activities SET points = -6 WHERE name = 'Gaming';
UPDATE activities SET points = -8 WHERE name = 'Social Media Scrolling';
UPDATE activities SET points = -10 WHERE activity_id = '8fd79d20-20a7-47f2-97b1-eadf441a9314';



-- +goose Down
UPDATE activities SET points = 60 WHERE name = 'Learning';
UPDATE activities SET points = 50 WHERE name = 'Exercise';
UPDATE activities SET points = 40 WHERE name = 'Meditation';
UPDATE activities SET points = 30 WHERE name = 'Reading';
UPDATE activities SET points = 20 WHERE name = 'Household Chores';
UPDATE activities SET points = -10 WHERE name = 'Watching Series';
UPDATE activities SET points = -20 WHERE name = 'Watching TV';
UPDATE activities SET points = -25 WHERE name = 'Gaming';
UPDATE activities SET points = -30 WHERE name = 'Social Media Scrolling';
UPDATE activities SET points = -40 WHERE activity_id = '8fd79d20-20a7-47f2-97b1-eadf441a9314';
ALTER TABLE activities DROP COLUMN edited;
