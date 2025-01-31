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

-- name: SetActivity :one
INSERT INTO user_activities (id , user_id , name , points , activity_type ) VALUES ($1 , $2 , $3 , $4 , $5 ) RETURNING *;

-- name: SetActivityLog :exec
INSERT INTO user_activity_logs (id , user_id , activity_id , duration , points , logged_at , activity_description  ) VALUES ($1 , $2 , $3 , $4 , $5 , $6 , $7  ); 

-- name: GetDailyActivityLogs :many
SELECT activity_id , ua.name ,  duration , user_activity_logs.points , activity_description FROM user_activity_logs LEFT  JOIN user_activities ua ON ua.id = user_activity_logs.activity_id WHERE user_activity_logs.user_id = $1 AND DATE(logged_at) = CURRENT_DATE;

-- name: GetDailyPoints :one
--
SELECT 
  CAST(COALESCE((SELECT SUM(ual.points) 
            FROM user_activity_logs ual 
            WHERE ual.user_id = $1 
              AND DATE(ual.logged_at) = CURRENT_DATE), 0) AS INTEGER) AS total_points,
  CAST(COALESCE((SELECT g.goal_points 
            FROM user_goals g 
            WHERE g.user_id = $1 
              AND DATE(g.created_at) = CURRENT_DATE), 0) AS INTEGER) AS goal_points;


-- name: GetDailyProductiveTime :one 
--
SELECT 
    CAST(COALESCE(SUM(CASE WHEN points > 0 THEN duration ELSE 0 END) , 0) AS INTEGER) AS productive_time,
    CAST(COALESCE(SUM(CASE WHEN points < 0 THEN duration ELSE 0 END), 0) AS INTEGER) AS unproductive_time
FROM 
    user_activity_logs
WHERE 
    DATE(logged_at) = CURRENT_DATE 
    AND user_id = $1;

-- name: GetRecentActivities :many

SELECT duration , user_activity_logs.points , activity_description, ua.name   FROM user_activity_logs LEFT JOIN user_activities ua ON ua.id = user_activity_logs.activity_id WHERE user_activity_logs.user_id = $1 AND DATE(logged_at) = CURRENT_DATE ORDER BY logged_at DESC LIMIT 3;

-- name: GetDailyActivityLogsCount :one
--
SELECT COUNT(*) as daily_activity_count
FROM user_activity_logs
WHERE user_id = $1 
AND DATE(logged_at) = CURRENT_DATE;

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
