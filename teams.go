package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
	"time"
)

func (apiCfg *apiConfig) CreateTeam(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name         string `json:"team_name"`
		TeamIndustry string `json:"team_industry"`
		TeamSize     int32  `json:"team_size"`
		IsPrivate    bool   `json:"is_private"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}
	team, err := apiCfg.DB.CreateTeam(r.Context(), database.CreateTeamParams{
		ID:           uuid.New(),
		Name:         params.Name,
		TeamIndustry: params.TeamIndustry,
		TeamSize:     params.TeamSize,
		IsPrivate:    params.IsPrivate,
		CreatedBy:    user.ID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in creating team: %s", err))
		return
	}
	teamrole, err := apiCfg.DB.SetTeamRole(r.Context(), database.SetTeamRoleParams{
		ID:       uuid.New(),
		RoleName: "owner",
		TeamID:   team.ID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in creating team roles: %s", err))
		return
	}
	teamMembership, err := apiCfg.DB.CreateTeamMembership(r.Context(), database.CreateTeamMembershipParams{
		ID:     uuid.New(),
		TeamID: team.ID,
		UserID: user.ID,
	})
	err = apiCfg.DB.CreateTeamUserRoles(r.Context(), database.CreateTeamUserRolesParams{
		ID:               uuid.New(),
		TeamMembershipID: teamMembership.ID,
		RoleID:           teamrole.ID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in creating team roles: %s", err))
		return
	}
	respondWithJson(w, 200, databaseTeamToTeam(team))
}

func (apiCfg *apiConfig) GetUserTeams(w http.ResponseWriter, r *http.Request, user database.User) {
	userteams, err := apiCfg.DB.GetUserTeams(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in getting user teams: %s", err))
		return
	}
	respondWithJson(w, 200, databaseUserTeamsToUserTeams(userteams))
}

func (apiCfg *apiConfig) GetTeamInfo(w http.ResponseWriter, r *http.Request) {
	teamid := r.PathValue("teamid")
	teamUUID, err := uuid.Parse(teamid)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	teaminfo, err := apiCfg.DB.GetTeamInFo(r.Context(), teamUUID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in getting team info : %s", err))
		return
	}
	respondWithJson(w, 200, databaseTeamInfoToTeamInfo(teaminfo))
}

func (apiCfg *apiConfig) GetTeamActivities(w http.ResponseWriter, r *http.Request) {
	teamId := r.PathValue("teamid")
	parsedTeamUUID, err := uuid.Parse(teamId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	teamactivities, err := apiCfg.DB.GetTeamActivities(r.Context(), parsedTeamUUID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in getting  team activites: %s", err))
		return
	}
	respondWithJson(w, 200, databaseTeamActivityToTeamActivity(teamactivities))
}

func (apiCfg *apiConfig) SetTeamRole(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		RoleName string `json:"role_name"`
		TeamID   string `json:"team_id"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}
	parsedTeamUUID, err := uuid.Parse(params.TeamID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	_, err = apiCfg.DB.SetTeamRole(r.Context(), database.SetTeamRoleParams{
		ID:       uuid.New(),
		RoleName: params.RoleName,
		TeamID:   parsedTeamUUID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in creating team roles: %s", err))
		return
	}
}

func (apiCfg *apiConfig) GetTeamRoles(w http.ResponseWriter, r *http.Request) {
	teamId := r.PathValue("teamid")
	allRoles := r.URL.Query().Get("allroles")
	teamUUID, err := uuid.Parse(teamId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	if allRoles == "true" {
		allTeamRoles, err := apiCfg.DB.GetAllTeamRoles(r.Context(), teamUUID)
		if err != nil {
			respondWithError(w, 500, fmt.Sprintf("Error in getting all team roles: %s", err))
			return
		}
		respondWithJson(w, 200, databaseAllTeamRolesToALlTeamRoles(allTeamRoles))
		return
	}
	teamRoles, err := apiCfg.DB.GetTeamRoles(r.Context(), teamUUID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in getting team roles: %s", err))
	}
	respondWithJson(w, 200, databaseTeamRolesToTeamRoles(teamRoles))
}

func (apiCfg *apiConfig) SetTeamActivity(w http.ResponseWriter, r *http.Request) {
	teamId := r.PathValue("teamid")
	parsedTeamUUID, err := uuid.Parse(teamId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	type parameters struct {
		ActivityName   string   `json:"activity_name"`
		ActivityPoints int32    `json:"activity_points"`
		ActivityRoles  []string `json:"activity_roles"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}
	err = apiCfg.DB.SetTeamActivity(r.Context(), database.SetTeamActivityParams{
		ID:            uuid.New(),
		TeamID:        parsedTeamUUID,
		ActivityName:  params.ActivityName,
		Points:        params.ActivityPoints,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		ActivityRoles: params.ActivityRoles,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in creating team roles: %s", err))
		return
	}
}

func (apiCfg *apiConfig) IsUserTeamOwner(w http.ResponseWriter, r *http.Request, user database.User) {
	teamid := r.PathValue("teamid")
	parsedTeamUUID, err := uuid.Parse(teamid)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	isowner, err := apiCfg.DB.IsUserTeamOwner(r.Context(), database.IsUserTeamOwnerParams{
		UserID: user.ID,
		TeamID: parsedTeamUUID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in getting user ownership: %s", err))
		return
	}
	respondWithJson(w, 200, isowner)
}

func (apiCfg *apiConfig) GetUserTeamActivities(w http.ResponseWriter, r *http.Request, user database.User) {
	teamid := r.PathValue("teamid")
	parsedTeamUUID, err := uuid.Parse(teamid)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing uuid: %s", err))
		return
	}
	userTeamActivities, err := apiCfg.DB.GetUserTeamActivities(r.Context(), database.GetUserTeamActivitiesParams{
		TeamID: parsedTeamUUID,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in getting user team activities: %s", err))
		return
	}
	respondWithJson(w, 200, databaseUserTeamActivityToUserTeamActivity(userTeamActivities))
}
