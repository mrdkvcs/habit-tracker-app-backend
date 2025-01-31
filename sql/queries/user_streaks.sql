-- name: GetStreakData :one
SELECT current_streak, longest_streak , last_logged_date FROM user_streaks WHERE user_id = $1;

-- name: UpdateStreakData :exec

INSERT INTO user_streaks (user_id, current_streak, longest_streak, last_logged_date) VALUES ($1 , $2 , $3 , $4) ON CONFLICT (user_id) DO UPDATE SET current_streak = $2, longest_streak = $3, last_logged_date = $4;
