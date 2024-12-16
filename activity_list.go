package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
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
	_, err = apiCfg.DB.SetActivity(r.Context(), database.SetActivityParams{
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
	parsedActivityUUID, err := uuid.Parse(activity_id)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing activity uuid: %s", err))
		return
	}
	activitylogExists, err := apiCfg.DB.CheckIfActivityLogExists(r.Context(), database.CheckIfActivityLogExistsParams{
		UserID:     user.ID,
		ActivityID: uuid.NullUUID{UUID: parsedActivityUUID, Valid: true},
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error checking if activity log exists: %s", err))
		return
	}
	respondWithJson(w, 200, activitylogExists)
}

func (apiCfg *apiConfig) SetNewActivity(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		ActivityName        string `json:"activity_name"`
		ActivityPoints      int32  `json:"activity_points"`
		ActivityDuration    int32  `json:"activity_duration"`
		ActivityDescription string `json:"activity_description"`
		OneTime             string `json:"one_time"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error decoding request body %s", err))
		return
	}
	if params.OneTime == "true" {
		err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
			ID:                  uuid.New(),
			UserID:              user.ID,
			ActivityID:          uuid.NullUUID{Valid: false},
			Duration:            params.ActivityDuration,
			Points:              params.ActivityPoints,
			LoggedAt:            time.Now(),
			ActivityDescription: params.ActivityDescription,
		})
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error setting activity log %s", err))
			return
		}
		return
	}
	activity, err := apiCfg.DB.SetActivity(r.Context(), database.SetActivityParams{
		ID:           uuid.New(),
		UserID:       user.ID,
		Name:         params.ActivityName,
		Points:       params.ActivityPoints,
		ActivityType: "custom",
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error setting activity %s", err))
		return
	}
	err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ActivityID:          uuid.NullUUID{UUID: activity.ID, Valid: true},
		Duration:            params.ActivityDuration,
		Points:              params.ActivityPoints,
		LoggedAt:            time.Now(),
		ActivityDescription: params.ActivityDescription,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error setting activity log %s", err))
		return
	}
}
