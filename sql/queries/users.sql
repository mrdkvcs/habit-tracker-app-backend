-- name: CreateUser :one
INSERT INTO users (id , created_at , updated_at , username , email , password_hash , google_id) VALUES ($1, $2, $3, $4 , $5 , $6 , $7) RETURNING *;


-- name: GetUserByGoogleId :one
SELECT * FROM users WHERE google_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

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

-- name: CreatePasswordReset :exec

INSERT INTO password_reset (id , user_id , token , expires_at) VALUES ($1 , $2 , $3  ,$4);

-- name: GetPasswordReset :one

SELECT u.email , pr.expires_at FROM password_reset pr  JOIN users u  ON u.id = pr.user_id  WHERE token = $1;

-- name: DeletePasswordReset :exec
DELETE FROM password_reset WHERE token = $1;

-- name: SetNewPassword :exec

UPDATE users SET password_hash = $1 WHERE email = $2;


