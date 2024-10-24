package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
)

func (apiCfg *apiConfig) GetTeamMembers(w http.ResponseWriter, r *http.Request, user database.User) {
	teamId := r.PathValue("teamid")
	parsedteamUUID, err := uuid.Parse(teamId)
	if err != nil {
		respondWithJson(w, 400, fmt.Sprintf("Error parsing UUID: %s", err))
		return
	}
	members, err := apiCfg.DB.GetTeamMembers(r.Context(), database.GetTeamMembersParams{
		TeamID: parsedteamUUID,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error getting team members: %s", err))
		return
	}
	respondWithJson(w, 200, databaseMembersToMembers(members))
}
func (apiCfg *apiConfig) SetMemberRoles(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		SelectedRoles []string `json:"selected_roles"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}
	membershipId := r.PathValue("membership_id")
	teamMembershipUUID, err := uuid.Parse(membershipId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing team membership uuid: %s", err))
		return
	}
	for _, roleId := range params.SelectedRoles {
		roleUUID, err := uuid.Parse(roleId)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error in parsing role uuid: %s", err))
			return
		}
		err = apiCfg.DB.SetMemberRoles(r.Context(), database.SetMemberRolesParams{
			ID:               uuid.New(),
			TeamMembershipID: teamMembershipUUID,
			RoleID:           roleUUID,
		})
		if err != nil {
			respondWithError(w, 500, fmt.Sprintf("Error in setting member roles: %s", err))
			return
		}
	}
}
func (apiCfg *apiConfig) GetNotAssignedRoles(w http.ResponseWriter, r *http.Request) {
	memberShipId := r.PathValue("membership_id")
	parsedmembershipUUID, err := uuid.Parse(memberShipId)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in parsing team membership uuid: %s", err))
		return
	}
	notAssignedRoles, err := apiCfg.DB.GetNotAssignedRoles(r.Context(), parsedmembershipUUID)
	respondWithJson(w, 200, databaseNotAssignedRolesToNotAssignedRoles(notAssignedRoles))
}
