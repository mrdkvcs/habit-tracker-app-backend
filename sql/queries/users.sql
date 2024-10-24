-- name: GetUserByApikey :one
SELECT * FROM users WHERE api_key = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id , created_at , updated_at , username , email , password_hash , api_key)
VALUES ($1, $2, $3, $4 , $5 , $6 , encode(sha256(random()::text::bytea) , 'hex'))
RETURNING *;

-- name: GetUsers :many
SELECT 
    u.id,
    u.username,
    CASE 
        WHEN ti.recipient_id IS NOT NULL THEN TRUE 
        ELSE FALSE 
    END AS has_been_invited
FROM 
    users u
LEFT JOIN 
    team_invitations ti 
ON 
    u.id = ti.recipient_id 
    AND ti.team_id = $1
WHERE 
    u.id != $2
    AND u.username ILIKE $3
ORDER BY 
    u.username
LIMIT 10;  

