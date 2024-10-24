-- name: CreateTeam :one
INSERT INTO teams (id , name,  team_industry , team_size , is_private , created_by) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: CreateTeamMembership :one
INSERT INTO team_memberships (id, team_id, user_id ) VALUES ($1, $2, $3 ) 
RETURNING *;

-- name: CreateTeamUserRoles :exec
INSERT INTO team_user_roles (id, team_membership_id, role_id) VALUES ($1, $2, $3 );

-- name: GetUserTeams :many
SELECT 
    t.id AS team_id,
    t.name,
    MAX(CASE 
        WHEN tr.role_name = 'owner' THEN 1 
        ELSE 0 
    END) AS is_owner
FROM 
    team_memberships tm
JOIN 
    teams t ON tm.team_id = t.id
LEFT JOIN 
    team_user_roles tur ON tm.id = tur.team_membership_id
LEFT JOIN 
    team_roles tr ON tur.role_id = tr.id
WHERE 
    tm.user_id = $1  
GROUP BY 
    t.id, t.name;

-- name: GetTeamInFo :one

SELECT t.id ,  t.name , t.team_industry , t.team_size , t.is_private  , t.created_by  
FROM teams t
WHERE t.id = $1;

-- name: GetTeamActivities :many
SELECT activity_name, points , activity_roles FROM team_activities WHERE team_id = $1;

-- name: SetTeamRole :one
INSERT INTO team_roles (id, role_name, team_id) VALUES ($1, $2, $3) RETURNING *;

-- name: GetTeamRoles :many
SELECT id, role_name FROM team_roles
WHERE team_id = $1 AND role_name <> 'owner';
-- name: GetAllTeamRoles :many
SELECT id, role_name FROM team_roles
WHERE team_id = $1;

-- name: SetTeamActivity :exec
INSERT INTO team_activities (id,team_id ,  activity_name, points, created_at , updated_at , activity_roles) VALUES ($1, $2, $3, $4 , $5 , $6 , $7);

-- name: IsUserTeamOwner :one
SELECT 
    COALESCE(tr.role_name = 'owner', false) AS is_owner
FROM 
    team_memberships tm
LEFT JOIN 
    team_user_roles tur ON tm.id = tur.team_membership_id
LEFT JOIN 
    team_roles tr ON tur.role_id = tr.id
WHERE 
    tm.user_id = $1
AND 
    tm.team_id = $2
LIMIT 1;
-- name: GetUserTeamActivities :many
WITH user_roles AS (
    SELECT array_agg(tr.role_name) AS roles
    FROM team_memberships tm
    JOIN team_user_roles tur ON tur.team_membership_id = tm.id
    JOIN team_roles tr ON tr.id = tur.role_id
    WHERE tm.user_id = $1
      AND tm.team_id =  $2
),
filtered_activities AS (
    SELECT ta.id, ta.team_id, ta.activity_name, ta.points, ta.created_at, ta.updated_at
    FROM team_activities ta, user_roles ur
    WHERE ta.team_id =  $2   AND (
        'owner' = ANY(ur.roles)  
        OR EXISTS (
            SELECT 1
            FROM unnest(ta.activity_roles) ar
            WHERE ar = ANY(ur.roles)
        )
    )
)
SELECT *
FROM filtered_activities
ORDER BY created_at DESC;


