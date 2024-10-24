-- name: SetProductivityGoal :exec

INSERT INTO user_goals (user_id , goal_date , goal_points) VALUES ($1, $2, $3);
