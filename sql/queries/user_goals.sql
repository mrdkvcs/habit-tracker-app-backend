-- name: SetProductivityGoal :exec

INSERT INTO user_goals (user_id , goal_date , goal_points) VALUES ($1, $2, $3);


-- name: SetGoalCompleted :exec

UPDATE user_goals SET status = 'completed' WHERE user_id = $1 AND goal_date = CURRENT_DATE
RETURNING *;
-- name: SetGoalUnCompleted :exec

UPDATE user_goals SET status = 'not completed' WHERE user_id = $1 AND goal_date = CURRENT_DATE
RETURNING *;

