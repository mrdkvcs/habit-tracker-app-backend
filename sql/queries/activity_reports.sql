-- name: GetTotalAndAverageProductivityPoints :one
WITH points_per_day AS ( SELECT DATE(logged_at) AS date, 
        COALESCE(SUM(points), 0) AS total_points
    FROM user_activity_logs
    WHERE user_id = $1 AND logged_at >= $2 AND logged_at < $3
    GROUP BY DATE(logged_at)
)

SELECT 
    ROUND(COALESCE(SUM(total_points), 0) , 0) AS total_points,
    ROUND(COALESCE(AVG(total_points), 0), 2) AS average_points_per_day
FROM points_per_day;

-- name: GetBestProductivityDay :one

SELECT DATE(logged_at) AS date, COALESCE(SUM(points ) , 0) AS total_points
FROM user_activity_logs
WHERE user_id = $1 AND logged_at >= $2 AND logged_at < $3
GROUP BY DATE(logged_at)
ORDER BY total_points DESC
LIMIT 1;

-- name: GetProductivityDays :many

SELECT DATE(logged_at) AS date , COALESCE(SUM(points ) , 0) AS total_points
FROM user_activity_logs
WHERE user_id = $1 AND logged_at >= $2 AND logged_at < $3
GROUP BY DATE(logged_at)
ORDER BY DATE(logged_at);
