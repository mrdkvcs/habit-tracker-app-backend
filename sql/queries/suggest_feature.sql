-- name: CreateSuggestFeature :exec
INSERT INTO suggest_feature (id , title, description , username) VALUES ($1, $2 , $3 , $4) RETURNING *;

-- name: GetSuggestFeature :many
SELECT * FROM suggest_feature ORDER BY upvote DESC;

-- name: SuggestFeatureUpvote :exec
UPDATE suggest_feature SET upvote = upvote + 1 WHERE id = $1 RETURNING *;
-- name: SuggestFeatureDownvote :exec
UPDATE suggest_feature SET upvote = upvote - 1 WHERE id = $1 RETURNING *;


