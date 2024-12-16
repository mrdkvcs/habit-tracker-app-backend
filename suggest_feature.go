package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
)

func (apiCfg *apiConfig) createSuggestFeature(w http.ResponseWriter, r *http.Request, user database.User) {

	type parameters struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error decoding request: %v", err))
		return
	}

	err = apiCfg.DB.CreateSuggestFeature(r.Context(), database.CreateSuggestFeatureParams{
		ID:          uuid.New(),
		Title:       params.Title,
		Description: params.Description,
		Username:    user.Username,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error creating suggestion: %v", err))
		return
	}

	respondWithJson(w, 200, "Suggestion created successfully , thanks for your feedback!")
}
func (apiCfg *apiConfig) GetSuggestFeature(w http.ResponseWriter, r *http.Request) {
	suggestFeatures, err := apiCfg.DB.GetSuggestFeature(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting suggest features: %v", err))
		return
	}
	respondWithJson(w, 200, databaseSuggestFeaturesToSuggestFeatures(suggestFeatures))
}

func (apiCfg *apiConfig) SetSuggestFeatureUpVote(w http.ResponseWriter, r *http.Request) {
	suggestFeatureID := r.PathValue("id")
	parsedSuggestFeatureUUID, err := uuid.Parse(suggestFeatureID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing suggest feature uuid: %s", err))
		return
	}
	err = apiCfg.DB.SuggestFeatureUpvote(r.Context(), parsedSuggestFeatureUUID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in setting suggest feature upvote: %s", err))
		return
	}
}
func (apiCfg *apiConfig) SetSuggestFeatureDownVote(w http.ResponseWriter, r *http.Request) {
	suggestFeatureID := r.PathValue("id")
	parsedSuggestFeatureUUID, err := uuid.Parse(suggestFeatureID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing suggest feature uuid: %s", err))
		return
	}
	err = apiCfg.DB.SuggestFeatureDownvote(r.Context(), parsedSuggestFeatureUUID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in setting suggest feature upvote: %s", err))
		return
	}
}
