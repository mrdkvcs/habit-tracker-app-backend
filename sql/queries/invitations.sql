-- name: CreateTeamInvitation :exec
INSERT INTO team_invitations (id, team_id, sender_id, recipient_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTeamInvitations :many
SELECT 
    ti.id AS invitation_id,
    t.id AS team_id,
    t.name AS team_name,
    t.team_industry AS team_industry,
    t.team_size AS team_size
FROM 
    team_invitations ti
JOIN 
    teams t 
ON 
    ti.team_id = t.id
WHERE 
    ti.recipient_id = $1
ORDER BY 
    ti.created_at DESC; 

-- name: GetInvitationsCount :one

SELECT COUNT(recipient_id) AS invite_count FROM team_invitations ti
WHERE ti.recipient_id = $1 AND ti.seen = false;

-- name: SetInvitationAsSeen :exec
UPDATE team_invitations
SET seen = true
WHERE recipient_id = $1 AND seen = false;

-- name: DeleteTeamInvitation :exec
DELETE FROM team_invitations
WHERE id = $1;
