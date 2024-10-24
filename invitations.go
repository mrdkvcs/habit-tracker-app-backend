package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"log"
	"net/http"
	"time"
)

func (apiCfg *apiConfig) CreateTeamInvitation(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		RecipientID string `json:"recipient_id"`
		SenderID    string `json:"sender_id"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	teamId := r.PathValue("teamid")
	parsedTeamUUID, err := uuid.Parse(teamId)
	if err != nil {
		respondWithJson(w, 500, "Error parsing team id")
	}
	parsedRecipientUUID, err := uuid.Parse(params.RecipientID)
	if err != nil {
		respondWithJson(w, 500, "Error parsing recipient id")
	}
	parsedSenderUUID, err := uuid.Parse(params.SenderID)
	if err != nil {
		respondWithJson(w, 500, "Error parsing sender id")
	}
	err = apiCfg.DB.CreateTeamInvitation(r.Context(), database.CreateTeamInvitationParams{
		ID:          uuid.New(),
		TeamID:      parsedTeamUUID,
		SenderID:    parsedSenderUUID,
		RecipientID: parsedRecipientUUID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		respondWithJson(w, 500, "Error creating team invitation")
		return
	}
	respondWithJson(w, 200, "Team invitation sent successfully ")
	broadcast <- params.RecipientID
}
func (apiCfg *apiConfig) GetTeamInvitations(w http.ResponseWriter, r *http.Request, user database.User) {
	teamInvites, err := apiCfg.DB.GetTeamInvitations(r.Context(), user.ID)
	if err != nil {
		respondWithJson(w, 500, "Error getting team invitations")
		return
	}
	respondWithJson(w, 200, databaseTeamInvitationsToTeamInvitations(teamInvites))
}
func (apiCfg *apiConfig) GetInvitationsCount(w http.ResponseWriter, r *http.Request, user database.User) {
	inviteCount, err := apiCfg.DB.GetInvitationsCount(r.Context(), user.ID)
	if err != nil {
		respondWithJson(w, 500, "Error getting  invitations count")
		return
	}
	respondWithJson(w, 200, inviteCount)
}

func (apiCfg *apiConfig) SetInvitationsAsSeen(w http.ResponseWriter, r *http.Request, user database.User) {
	err := apiCfg.DB.SetInvitationAsSeen(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error setting invitation as seen: %s", err)
		return
	}
}
func (apiCfg *apiConfig) AcceptTeamInvite(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		TeamID   string `json:"team_id"`
		InviteID string `json:"invitation_id"`
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
		respondWithError(w, 400, fmt.Sprintf("Error in parsing team uuid: %s", err))
		return
	}
	parsedInviteUUID, err := uuid.Parse(params.InviteID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing invite uuid: %s", err))
		return
	}
	_, err = apiCfg.DB.CreateTeamMembership(r.Context(), database.CreateTeamMembershipParams{
		ID:     uuid.New(),
		TeamID: parsedTeamUUID,
		UserID: user.ID,
	})
	err = apiCfg.DB.DeleteTeamInvitation(r.Context(), parsedInviteUUID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in deleting team invitation: %s", err))
		return
	}
}
func (apiCfg *apiConfig) DeclineTeamInvite(w http.ResponseWriter, r *http.Request) {
	invitationId := r.PathValue("invitationid")
	parsedInviteUUID, err := uuid.Parse(invitationId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing invite uuid: %s", err))
		return
	}
	err = apiCfg.DB.DeleteTeamInvitation(r.Context(), parsedInviteUUID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in deleting team invitation: %s", err))
		return
	}
}
