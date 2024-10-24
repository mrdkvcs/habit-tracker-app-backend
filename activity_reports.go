package main

import (
	"encoding/json"
	"fmt"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
	"time"
)

type DatabaseProductivityStats struct {
	ProductivityPoints  database.GetTotalAndAverageProductivityPointsRow
	BestProductivityDay database.GetBestProductivityDayRow
	ProductivityDays    []database.GetProductivityDaysRow
}

func (apiCfg *apiConfig) GetProductivityStats(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error decoding request body: %v", err))
		return
	}
	productivityPoints, err := apiCfg.DB.GetTotalAndAverageProductivityPoints(r.Context(), database.GetTotalAndAverageProductivityPointsParams{
		UserID:     user.ID,
		LoggedAt:   params.StartTime,
		LoggedAt_2: params.EndTime,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error getting productivity stats: %v", err))
		return
	}
	bestProductivityDay, err := apiCfg.DB.GetBestProductivityDay(r.Context(), database.GetBestProductivityDayParams{
		UserID:     user.ID,
		LoggedAt:   params.StartTime,
		LoggedAt_2: params.EndTime,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error getting best productivity day: %v", err))
		return
	}
	productivityDays, err := apiCfg.DB.GetProductivityDays(r.Context(), database.GetProductivityDaysParams{
		UserID:     user.ID,
		LoggedAt:   params.StartTime,
		LoggedAt_2: params.EndTime,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error getting productivity days: %v", err))
		return
	}
	databaseProductivityStats := DatabaseProductivityStats{
		ProductivityPoints:  productivityPoints,
		BestProductivityDay: bestProductivityDay,
		ProductivityDays:    productivityDays,
	}
	respondWithJson(w, 200, databaseProductivityStatsToProductivityStats(databaseProductivityStats))
}
