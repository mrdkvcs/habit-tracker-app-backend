package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"log"
	"math"
	"net/http"
	"time"
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
	activity, err := apiCfg.DB.SetActivity(r.Context(), database.SetActivityParams{
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
	respondWithJson(w, 200, databaseActivityToActivity(SetActivityRowWrapper{activity}))
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
	isStreakRecord := false
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error decoding request body %s", err))
		return
	}
	if params.OneTime == "true" {
		pointsPerMinutes := float64(params.ActivityPoints) / float64(60)
		points := float64(params.ActivityDuration) * float64(pointsPerMinutes)
		roundedPoints := math.Round(points)
		err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
			ID:                  uuid.New(),
			UserID:              user.ID,
			ActivityID:          uuid.NullUUID{Valid: false},
			Duration:            params.ActivityDuration,
			Points:              int32(roundedPoints),
			LoggedAt:            time.Now(),
			ActivityDescription: params.ActivityDescription,
		})
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error setting activity log %s", err))
			return
		}

		dailyPoints, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error getting daily points: %v", err))
			return
		}

		if dailyPoints.GoalPoints > 0 {
			isGoalCompleted := dailyPoints.TotalPoints > dailyPoints.GoalPoints
			dailyPoints.TotalPoints = dailyPoints.TotalPoints + params.ActivityPoints
			if dailyPoints.TotalPoints > dailyPoints.GoalPoints && !isGoalCompleted {
				stopChan <- struct{}{}
				err := apiCfg.DB.SetGoalCompleted(r.Context(), user.ID)
				if err != nil {
					respondWithError(w, 400, fmt.Sprintf("Error setting goal completed: %v", err))
					return
				}
				return
			}
			if dailyPoints.GoalPoints > dailyPoints.TotalPoints && isGoalCompleted {
				err := apiCfg.DB.SetGoalUnCompleted(r.Context(), user.ID)
				if err != nil {
					respondWithError(w, 400, fmt.Sprintf("Error setting goal uncompleted: %v", err))
					return
				}
				startGoalTracker(user.ID, user.Email)
			}
		}
		dailyActivityLogCount, err := apiCfg.DB.GetDailyActivityLogsCount(r.Context(), user.ID)
		if err != nil {
			log.Printf("Error getting daily activity logs count: %s", err)
		}

		if dailyActivityLogCount == 1 {
			streakInfo, err := apiCfg.DB.GetStreakData(r.Context(), user.ID)
			if err != nil && err != sql.ErrNoRows {
				respondWithError(w, 400, fmt.Sprintf("Error getting streak info: %v", err))
				return
			}
			yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			if streakInfo.LastLoggedDate.Valid && streakInfo.LastLoggedDate.Time.Format("2006-01-02") == yesterday {
				streakInfo.CurrentStreak += 1
			} else {
				streakInfo.CurrentStreak = 1
			}
			if streakInfo.CurrentStreak > streakInfo.LongestStreak {
				isStreakRecord = true
				streakInfo.LongestStreak = streakInfo.CurrentStreak
			}
			err = apiCfg.DB.UpdateStreakData(r.Context(), database.UpdateStreakDataParams{
				UserID:         user.ID,
				CurrentStreak:  streakInfo.CurrentStreak,
				LongestStreak:  streakInfo.LongestStreak,
				LastLoggedDate: sql.NullTime{Valid: true, Time: time.Now()},
			})
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error updating streak info: %v", err))
				return
			}
			respondWithJson(w, 200, ActivityLogResponse{StreakCount: streakInfo.CurrentStreak, IsStreakRecord: isStreakRecord})
			return
		}
		respondWithJson(w, 200, ActivityLogResponse{})
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
	pointsPerMinutes := float64(params.ActivityPoints) / float64(60)
	points := float64(params.ActivityDuration) * float64(pointsPerMinutes)
	roundedPoints := math.Round(points)
	err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ActivityID:          uuid.NullUUID{UUID: activity.ID, Valid: true},
		Duration:            params.ActivityDuration,
		Points:              int32(roundedPoints),
		LoggedAt:            time.Now(),
		ActivityDescription: params.ActivityDescription,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error setting activity log %s", err))
		return
	}
	dailyPoints, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily points: %v", err))
		return
	}
	if dailyPoints.GoalPoints > 0 {
		isGoalCompleted := dailyPoints.TotalPoints > dailyPoints.GoalPoints
		dailyPoints.TotalPoints = dailyPoints.TotalPoints + params.ActivityPoints
		if dailyPoints.TotalPoints > dailyPoints.GoalPoints && !isGoalCompleted {
			stopChan <- struct{}{}
			err := apiCfg.DB.SetGoalCompleted(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error setting goal completed: %v", err))
				return
			}
			return
		}
		if dailyPoints.GoalPoints > dailyPoints.TotalPoints && isGoalCompleted {
			err := apiCfg.DB.SetGoalUnCompleted(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error setting goal uncompleted: %v", err))
				return
			}
			startGoalTracker(user.ID, user.Email)
		}
	}
	dailyActivityLogCount, err := apiCfg.DB.GetDailyActivityLogsCount(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error getting daily activity logs count: %s", err)
	}

	if dailyActivityLogCount == 1 {
		streakInfo, err := apiCfg.DB.GetStreakData(r.Context(), user.ID)
		if err != nil && err != sql.ErrNoRows {
			respondWithError(w, 400, fmt.Sprintf("Error getting streak info: %v", err))
			return
		}
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		if streakInfo.LastLoggedDate.Valid && streakInfo.LastLoggedDate.Time.Format("2006-01-02") == yesterday {
			streakInfo.CurrentStreak += 1
		} else {
			streakInfo.CurrentStreak = 1
		}
		if streakInfo.CurrentStreak > streakInfo.LongestStreak {
			isStreakRecord = true
			streakInfo.LongestStreak = streakInfo.CurrentStreak
		}
		err = apiCfg.DB.UpdateStreakData(r.Context(), database.UpdateStreakDataParams{
			UserID:         user.ID,
			CurrentStreak:  streakInfo.CurrentStreak,
			LongestStreak:  streakInfo.LongestStreak,
			LastLoggedDate: sql.NullTime{Valid: true, Time: time.Now()},
		})
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error updating streak info: %v", err))
			return
		}
		respondWithJson(w, 200, ActivityLogResponse{StreakCount: streakInfo.CurrentStreak, IsStreakRecord: isStreakRecord})
		return
	}
	respondWithJson(w, 200, ActivityLogResponse{})
	return
}
