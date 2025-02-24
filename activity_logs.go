package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	openai "github.com/sashabaranov/go-openai"
)

type DBDailyStats struct {
	TotalPoints            interface{}
	GoalPoints             interface{}
	DailyProductiveTime    database.GetDailyProductiveTimeRow
	RecentActivities       []database.GetRecentActivitiesRow
	DailyActivityLogsCount int64
	CurrentStreak          int32
	LongestStreak          int32
	StreakMessage          string
}

func extractFromInput(input string, api_key string) (map[string]int, error) {
	client := openai.NewClient(api_key)
	response, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: extractSystemInstruction,
			},
			{
				Role:    "user",
				Content: input,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error in creating chat completion: %v", err)
	}

	pattern := `Activity\d*:\s*(?P<activity>[\w\s]+),\s*Duration\d*:\s*(?P<duration>\d+)`

	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(response.Choices[0].Message.Content, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("Invalid input")
	}

	activitiesMap := make(map[string]int)

	for _, match := range matches {
		if len(match) >= 3 {
			activity := match[1]
			durationStr := match[2]
			duration, err := strconv.Atoi(durationStr)
			if err != nil {
				return nil, fmt.Errorf("Invalid duration format")
			}
			activitiesMap[activity] = duration
		}
	}

	return activitiesMap, nil
}

func compareActivities(extractedActivity string, databaseActivity string) (bool, error) {
	input := fmt.Sprintf("%s , %s", extractedActivity, databaseActivity)
	client := openai.NewClient(api_key)
	response, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: compareSystemInstruction,
			},
			{
				Role:    "user",
				Content: input,
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("error in creating chat completion: %v", err)
	}
	if response.Choices[0].Message.Content == "false" {
		return false, nil
	}
	return true, nil
}

type ActivityLogResponse struct {
	MatchedActivities []Activity `json:"matched_activities"`
	Duration          int        `json:"duration,omitzero"`
	Description       string     `json:"description,omitzero"`
	Name              string     `json:"name,omitzero"`
	StreakCount       int32      `json:"streak_count,omitzero"`
	IsStreakRecord    bool       `json:"is_streak_record,omitzero"`
}

type MultipleActivityLogResponse struct {
	ActivityLogs   []ActivityLogResponse `json:"activity_logs"`
	StreakCount    int32                 `json:"streak_count,omitzero"`
	IsStreakRecord bool                  `json:"is_streak_record,omitzero"`
}

func (apiCfg *apiConfig) SetActivityLog(w http.ResponseWriter, r *http.Request, user database.User) {

	type parameters struct {
		ActivityInput string `json:"activity_input"`
	}

	isStreakRecord := false
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}

	activitiesMap, err := extractFromInput(params.ActivityInput, api_key)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in processing your input : %s", err))
		return
	}

	userActivities, err := apiCfg.DB.GetActivities(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting activities: %v", err))
		return
	}

	if len(activitiesMap) == 1 {
		var activity string
		var duration int
		for k, v := range activitiesMap {
			activity = k
			duration = v
		}
		matchCounter := 0
		matchedActivities := []Activity{}
		for _, dbactivity := range userActivities {
			match, err := compareActivities(activity, dbactivity.Name)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error comparing activities: %v", err))
				return
			}
			if match {
				matchCounter += 1
				matchedActivities = append(matchedActivities, databaseActivityToActivity(GetActivitiesRowWrapper{dbactivity}))
			}
		}
		if matchCounter == 1 {
			pointsPerMinutes := float64(matchedActivities[0].Points) / float64(60)
			points := float64(duration) * float64(pointsPerMinutes)
			roundedPoints := math.Round(points)

			dailyMinutes, err := apiCfg.DB.GetDailyMinutes(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error getting daily minutes: %v", err))
				return
			}

			if (dailyMinutes + int64(duration)) > 1440 {
				respondWithError(w, 400, "You have reached your daily limit , you would have more than 24 hours of activity in a day")
				return
			}

			err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
				ID:                  uuid.New(),
				UserID:              user.ID,
				ActivityID:          uuid.NullUUID{UUID: matchedActivities[0].ActivityID, Valid: true},
				Duration:            int32(duration),
				Points:              int32(roundedPoints),
				ActivityDescription: params.ActivityInput,
				LoggedAt:            time.Now(),
			})
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error setting activity log: %v", err))
				return
			}

			dailyPoints, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error getting daily points: %v", err))
				return
			}

			if dailyPoints.GoalPoints > 0 {
				isGoalCompleted := dailyPoints.TotalPoints > dailyPoints.GoalPoints
				dailyPoints.TotalPoints = dailyPoints.TotalPoints + int32(roundedPoints)
				if dailyPoints.TotalPoints > dailyPoints.GoalPoints && !isGoalCompleted {
					stopChan <- struct{}{}
					err := apiCfg.DB.SetGoalCompleted(r.Context(), user.ID)
					if err != nil {
						respondWithError(w, 400, fmt.Sprintf("Error setting goal completed: %v", err))
						return
					}
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
				respondWithJson(w, 200, ActivityLogResponse{MatchedActivities: matchedActivities, StreakCount: streakInfo.CurrentStreak, IsStreakRecord: isStreakRecord})
				return
			}
			respondWithJson(w, 200, ActivityLogResponse{MatchedActivities: matchedActivities})
			return
		}
		if matchCounter > 1 {
			respondWithJson(w, 200, ActivityLogResponse{MatchedActivities: matchedActivities, Duration: duration, Description: params.ActivityInput})
			return
		}
		respondWithJson(w, 200, ActivityLogResponse{MatchedActivities: matchedActivities, Duration: duration, Description: params.ActivityInput, Name: activity})
		return
	}

	multipleMatchedActivities := MultipleActivityLogResponse{}
	for k, v := range activitiesMap {
		matchedActivities := []Activity{}
		matchCounter := 0
		for _, dbActivity := range userActivities {
			match, err := compareActivities(k, dbActivity.Name)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error comparing activities : %s", err))
			}
			if match {
				matchedActivities = append(matchedActivities, databaseActivityToActivity(GetActivitiesRowWrapper{dbActivity}))
				matchCounter += 1
			}
		}
		if matchCounter == 1 {
			pointsPerMinutes := float64(matchedActivities[0].Points) / float64(60)
			points := float64(v) * float64(pointsPerMinutes)
			roundedPoints := math.Round(points)

			dailyMinutes, err := apiCfg.DB.GetDailyMinutes(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error getting daily minutes: %v", err))
				return
			}

			if (dailyMinutes + int64(v)) > 1440 {
				respondWithError(w, 400, "You have reached your daily limit , you would have more than 24 hours of activity in a day")
				return
			}

			err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
				ID:                  uuid.New(),
				UserID:              user.ID,
				ActivityID:          uuid.NullUUID{UUID: matchedActivities[0].ActivityID, Valid: true},
				Duration:            int32(v),
				Points:              int32(roundedPoints),
				ActivityDescription: params.ActivityInput,
				LoggedAt:            time.Now(),
			})
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error setting activity log: %v", err))
				return
			}

			dailyPoints, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error getting daily points: %v", err))
				return
			}

			if dailyPoints.GoalPoints > 0 {
				isGoalCompleted := dailyPoints.TotalPoints > dailyPoints.GoalPoints
				dailyPoints.TotalPoints = dailyPoints.TotalPoints + int32(roundedPoints)
				if dailyPoints.TotalPoints > dailyPoints.GoalPoints && !isGoalCompleted {
					stopChan <- struct{}{}
					err := apiCfg.DB.SetGoalCompleted(r.Context(), user.ID)
					if err != nil {
						respondWithError(w, 400, fmt.Sprintf("Error setting goal completed: %v", err))
						return
					}
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
				multipleMatchedActivities.IsStreakRecord = isStreakRecord
				multipleMatchedActivities.StreakCount = streakInfo.CurrentStreak
			}
			activityLogResponse := ActivityLogResponse{MatchedActivities: matchedActivities}
			multipleMatchedActivities.ActivityLogs = append(multipleMatchedActivities.ActivityLogs, activityLogResponse)
		} else if matchCounter > 1 {
			activityLogResponse := ActivityLogResponse{MatchedActivities: matchedActivities, Duration: v, Description: params.ActivityInput, Name: k}
			multipleMatchedActivities.ActivityLogs = append(multipleMatchedActivities.ActivityLogs, activityLogResponse)
		} else {
			activityLogResponse := ActivityLogResponse{MatchedActivities: matchedActivities, Duration: v, Description: params.ActivityInput, Name: k}
			multipleMatchedActivities.ActivityLogs = append(multipleMatchedActivities.ActivityLogs, activityLogResponse)
		}
	}
	respondWithJson(w, 200, multipleMatchedActivities)
}

func (apiCfg *apiConfig) SetSpecificActivityLog(w http.ResponseWriter, r *http.Request, user database.User) {

	type parameters struct {
		ActivityID          string `json:"activity_id"`
		ActivityName        string `json:"activity_name"`
		ActivityPoints      int32  `json:"activity_points"`
		ActivityDuration    int32  `json:"activity_duration"`
		ActivityDescription string `json:"activity_description"`
	}

	isStreakRecord := false

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}

	activityUUID, err := uuid.Parse(params.ActivityID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing activity uuid: %s", err))
		return
	}

	pointsInMinutes := float64(params.ActivityPoints) / float64(60)
	points := float64(params.ActivityDuration) * float64(pointsInMinutes)
	roundedPoints := math.Round(points)

	err = apiCfg.DB.SetActivityLog(r.Context(), database.SetActivityLogParams{
		ID:                  uuid.New(),
		UserID:              user.ID,
		ActivityID:          uuid.NullUUID{UUID: activityUUID, Valid: true},
		Duration:            params.ActivityDuration,
		Points:              int32(roundedPoints),
		ActivityDescription: params.ActivityDescription,
		LoggedAt:            time.Now(),
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error setting activity log: %v", err))
		return
	}

	dailyPoints, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily points: %v", err))
		return
	}

	if dailyPoints.GoalPoints > 0 {
		isGoalCompleted := dailyPoints.TotalPoints > dailyPoints.GoalPoints
		dailyPoints.TotalPoints = dailyPoints.TotalPoints + int32(roundedPoints)
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

func (apiCfg *apiConfig) GetDailyActivityLogs(w http.ResponseWriter, r *http.Request, user database.User) {
	dailyLogs, err := apiCfg.DB.GetDailyActivityLogs(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily activity logs: %v", err))
		return
	}
	respondWithJson(w, 200, databaseActivityLogsToActivityLogs(dailyLogs))
}

func (apiCfg *apiConfig) GetDailyStats(w http.ResponseWriter, r *http.Request, user database.User) {
	points, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
	var message string
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily points: %v", err))
		return
	}
	dailyTime, err := apiCfg.DB.GetDailyProductiveTime(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily productivity time: %v", err))
		return
	}
	recentActivities, err := apiCfg.DB.GetRecentActivities(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting recent activities: %v", err))
		return
	}
	dailyActivityLogsCount, err := apiCfg.DB.GetDailyActivityLogsCount(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily activity logs count: %v", err))
		return
	}
	yesterday := time.Now().AddDate(0, 0, -1)
	streakInfo, err := apiCfg.DB.GetStreakData(r.Context(), user.ID)
	if err == sql.ErrNoRows {
		dbDailyStats := DBDailyStats{
			TotalPoints:            points.TotalPoints,
			GoalPoints:             points.GoalPoints,
			DailyProductiveTime:    dailyTime,
			RecentActivities:       recentActivities,
			DailyActivityLogsCount: dailyActivityLogsCount,
			CurrentStreak:          0,
			LongestStreak:          0,
			StreakMessage:          "Welcome here ! You can start a productivity streak by setting activities daily!🚀",
		}
		respondWithJson(w, 200, DatabaseDailyStatsToDailyStats(dbDailyStats))
		return
	} else if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting streak info: %v", err))
		return
	}
	if streakInfo.LastLoggedDate.Time.Format("2006-01-02") < yesterday.Format("2006-01-02") && streakInfo.CurrentStreak != 0 {
		streakInfo.CurrentStreak = 0
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
		message = "😔 Oops, your streak ended yesterday. But every day is a new opportunity! Start fresh today and build it back up! 🌟💪"
	}

	if streakInfo.LastLoggedDate.Time.Format("2006-01-02") < yesterday.Format("2006-01-02") {
		message = "😢 Oh no, your streak is at 0! But guess what? Today is a new chance to start strong! 🌟 Dive back in and build it up! 💪"
	}

	if streakInfo.LastLoggedDate.Time.Format("2006-01-02") == time.Now().Format("2006-01-02") {
		message = "🎉 You're on a roll! You've completed your daily streak today! Keep up the amazing work! 🔥💪"
	}

	if streakInfo.LastLoggedDate.Time.Format("2006-01-02") == yesterday.Format("2006-01-02") {
		message = "🔥 You're on a streak! Don't forget to log an activity today to keep it going! You’ve got this! 💪✨"
	}

	dbDailyStats := DBDailyStats{
		TotalPoints:            points.TotalPoints,
		GoalPoints:             points.GoalPoints,
		DailyProductiveTime:    dailyTime,
		RecentActivities:       recentActivities,
		DailyActivityLogsCount: dailyActivityLogsCount,
		CurrentStreak:          streakInfo.CurrentStreak,
		LongestStreak:          streakInfo.LongestStreak,
		StreakMessage:          message,
	}
	respondWithJson(w, 200, DatabaseDailyStatsToDailyStats(dbDailyStats))
}
