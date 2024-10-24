package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
)

func (apiCfg *apiConfig) GetActivites(w http.ResponseWriter, r *http.Request, user database.User) {
	activities, err := apiCfg.DB.GetActivities(r.Context(), user.ID)
	if err != nil {
		respondWithJson(w, 400, fmt.Sprintf("Error getting activities: %v", err))
	}
	respondWithJson(w, 200, databaseActivitiesToActivities(activities))
}

func (apiCfg *apiConfig) SetActivity(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name   string `json:"name"`
		Points int32  `json:"points"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error decoding parameters: %v", err))
		return
	}
	err = apiCfg.DB.SetActivity(r.Context(), database.SetActivityParams{
		ID:           uuid.New(),
		UserID:       user.ID,
		Name:         params.Name,
		Points:       params.Points,
		ActivityType: "custom",
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error setting activity: %v", err))
		return
	}
}

func (apiCfg *apiConfig) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	activity_id := r.PathValue("id")
	parsedActivityUUID, err := uuid.Parse(activity_id)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing activity uuid: %s", err))
		return
	}
	err = apiCfg.DB.DeleteActivity(r.Context(), parsedActivityUUID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error deleting activity: %v", err))
		return
	}
	respondWithJson(w, 200, "Activity deleted successfully")
}
func (apiCfg *apiConfig) EditActivity(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		ActivityName   string `json:"activity_name"`
		ActivityPoints int32  `json:"activity_points"`
	}

	activityId := r.PathValue("id")
	parsedActivityUUID, err := uuid.Parse(activityId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing activity uuid: %s", err))
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error decoding request: %v", err))
		return
	}

	err = apiCfg.DB.EditActivity(r.Context(), database.EditActivityParams{
		Name:   params.ActivityName,
		Points: params.ActivityPoints,
		ID:     parsedActivityUUID,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error editing activity: %v", err))
		return
	}

}

func (apiCfg *apiConfig) CheckActivityLogExists(w http.ResponseWriter, r *http.Request, user database.User) {
	activity_id := r.URL.Query().Get("activity_id")
	user_id := r.URL.Query().Get("user_id")
	parsedActivityUUID, err := uuid.Parse(activity_id)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing activity uuid: %s", err))
		return
	}
	parsedUserUUID, err := uuid.Parse(user_id)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing user uuid: %s", err))
		return
	}
	activitylogExists, err := apiCfg.DB.CheckIfActivityLogExists(r.Context(), database.CheckIfActivityLogExistsParams{
		UserID:     parsedUserUUID,
		ActivityID: parsedActivityUUID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error checking if activity log exists: %s", err))
		return
	}
	respondWithJson(w, 200, activitylogExists)
}
