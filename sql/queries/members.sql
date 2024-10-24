-- name: GetTeamMembers :many
SELECT 
    tm.id,
    u.username,
    COALESCE(CAST(STRING_AGG(tr.role_name, ', ') AS TEXT), '') AS roles
FROM 
    team_memberships tm
JOIN 
    users u ON tm.user_id = u.id
LEFT JOIN 
    team_user_roles tur ON tm.id = tur.team_membership_id
LEFT JOIN 
    team_roles tr ON tur.role_id = tr.id
WHERE 
    tm.team_id = $1
    AND tm.user_id <> $2
GROUP BY 
    u.username,
    tm.id
ORDER BY 
    u.username;

-- name: SetMemberRoles :exec
INSERT INTO team_user_roles (id, team_membership_id, role_id) VALUES ($1, $2, $3);

-- name: GetNotAssignedRoles :many
SELECT 
    tr.id, 
    tr.role_name
FROM 
    team_roles tr
LEFT JOIN 
    team_user_roles tur ON tr.id = tur.role_id 
    AND tur.team_membership_id = $1
WHERE 
    tur.role_id IS NULL
ORDER BY 
    tr.role_name;

