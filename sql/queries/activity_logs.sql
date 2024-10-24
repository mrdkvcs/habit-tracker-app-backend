-- name: SetDefaultActivities :exec
INSERT INTO user_activities (id , user_id , name , points , activity_type ) 
VALUES
(uuid_generate_v4(), $1, 'Learning', 10, 'default'),
(uuid_generate_v4(), $1, 'Exercise', 8, 'default'),
(uuid_generate_v4(), $1, 'Meditation', 6, 'default'),
(uuid_generate_v4(), $1, 'Reading', 4, 'default'),
(uuid_generate_v4(), $1, 'Household Chores', 2, 'default'),
(uuid_generate_v4(), $1, 'Watching Series', -2, 'default'),
(uuid_generate_v4(), $1, 'Watching TV', -4, 'default'),
(uuid_generate_v4(), $1, 'Gaming', -6, 'default'),
(uuid_generate_v4(), $1, 'Social Media Scrolling', -8, 'default'),
(uuid_generate_v4(), $1, 'Watching adult websites', -10, 'default');
-- name: GetActivities :many
SELECT id , name , points , activity_type FROM user_activities WHERE user_id = $1 ORDER BY points DESC;

-- name: SetActivity :exec
INSERT INTO user_activities (id , user_id , name , points , activity_type ) VALUES ($1 , $2 , $3 , $4 , $5 );

-- name: SetActivityLog :exec
INSERT INTO user_activity_logs (id , user_id , activity_id , duration , points , logged_at , activity_description  ) VALUES ($1 , $2 , $3 , $4 , $5 , $6 , $7  ); 

-- name: GetDailyActivityLogs :many
SELECT activity_id , ua.name ,  duration , user_activity_logs.points , activity_description FROM user_activity_logs JOIN user_activities ua ON ua.id = user_activity_logs.activity_id WHERE user_activity_logs.user_id = $1 AND DATE(logged_at) = CURRENT_DATE;

-- name: GetDailyPoints :one
--
SELECT 
  COALESCE((SELECT SUM(ual.points) 
            FROM user_activity_logs ual 
            WHERE ual.user_id = $1 
              AND DATE(ual.logged_at) = CURRENT_DATE), 0) AS total_points,
  COALESCE((SELECT g.goal_points 
            FROM user_goals g 
            WHERE g.user_id = $1 
              AND DATE(g.goal_date) = CURRENT_DATE), 0) AS goal_points;

-- name: EditActivity :exec
--
UPDATE user_activities SET name = $1 , points = $2, activity_type = 'custom' WHERE id = $3;

-- name: CheckIfActivityLogExists :one

SELECT EXISTS (
  SELECT 1 FROM user_activity_logs WHERE user_id = $1 AND activity_id = $2
);

-- name: DeleteActivity :exec
DELETE FROM user_activities WHERE id = $1;

-- name: GetDailyMinutes :one

SELECT COALESCE(SUM(duration), 0)::BIGINT AS total_hours
FROM user_activity_logs
WHERE user_id = $1
AND DATE(logged_at) = CURRENT_DATE;
