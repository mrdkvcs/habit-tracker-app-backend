package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	openai "github.com/sashabaranov/go-openai"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func extractFromInput(input string, api_key string) (string, int, error) {
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
		return "", 0, fmt.Errorf("error in creating chat completion: %v", err)
	}

	pattern := `Activity:\s*(?P<activity>[\w\s]+),\s*Duration:\s*(?P<duration>\d+)`

	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(response.Choices[0].Message.Content)

	if len(matches) < 2 {
		return "", 0, fmt.Errorf("invalid input")
	}

	activity := matches[1]
	durationStr := matches[2]

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid duration format")
	}

	return activity, duration, nil
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

type MatchedActivities struct {
	MatchedActivities []Activity  `json:"matched_activities"`
	Duration          int         `json:"duration"`
	Description       string      `json:"description"`
	Name              interface{} `json:"name"`
}

func (apiCfg *apiConfig) SetActivityLog(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		ActivityInput string `json:"activity_input"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in parsing json: %s", err))
		return
	}

	activity, duration, err := extractFromInput(params.ActivityInput, api_key)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in processing your input : %s", err))
		return
	}

	userActivities, err := apiCfg.DB.GetActivities(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting activities: %v", err))
		return
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
			matchedActivities = append(matchedActivities, databaseActivityToActivity(dbactivity))
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
		isGoalCompleted := totalPoints > goalPoints
		totalPoints = totalPoints + int32(roundedPoints)
		if totalPoints > goalPoints && !isGoalCompleted {
			stopChan <- struct{}{}
			err := apiCfg.DB.SetGoalCompleted(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error setting goal completed: %v", err))
				return
			}
		}
		if goalPoints > totalPoints && isGoalCompleted {
			err := apiCfg.DB.SetGoalUnCompleted(r.Context(), user.ID)
			if err != nil {
				respondWithError(w, 400, fmt.Sprintf("Error setting goal uncompleted: %v", err))
				return
			}
			startGoalTracker(user.ID, user.Email)
		}
		respondWithJson(w, 200, MatchedActivities{MatchedActivities: matchedActivities})
		return
	}
	if matchCounter > 1 {
		respondWithJson(w, 200, MatchedActivities{MatchedActivities: matchedActivities, Duration: duration, Description: params.ActivityInput})
		return
	}
	respondWithJson(w, 200, MatchedActivities{MatchedActivities: matchedActivities, Duration: duration, Description: params.ActivityInput, Name: activity})
}

func (apiCfg *apiConfig) SetSpecificActivityLog(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		ActivityID          string `json:"activity_id"`
		ActivityName        string `json:"activity_name"`
		ActivityPoints      int32  `json:"activity_points"`
		ActivityDuration    int32  `json:"activity_duration"`
		ActivityDescription string `json:"activity_description"`
	}
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
	isGoalCompleted := totalPoints > goalPoints
	totalPoints = totalPoints + int32(roundedPoints)
	if totalPoints > goalPoints && !isGoalCompleted {
		stopChan <- struct{}{}
		err := apiCfg.DB.SetGoalCompleted(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error setting goal completed: %v", err))
			return
		}
		return
	}
	if goalPoints > totalPoints && isGoalCompleted {
		err := apiCfg.DB.SetGoalUnCompleted(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error setting goal uncompleted: %v", err))
			return
		}
		startGoalTracker(user.ID, user.Email)
	}
}

func (apiCfg *apiConfig) GetDailyActivityLogs(w http.ResponseWriter, r *http.Request, user database.User) {
	dailyLogs, err := apiCfg.DB.GetDailyActivityLogs(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily activity logs: %v", err))
		return
	}
	respondWithJson(w, 200, databaseActivityLogsToActivityLogs(dailyLogs))
}

func (apiCfg *apiConfig) GetDailyPoints(w http.ResponseWriter, r *http.Request, user database.User) {
	points, err := apiCfg.DB.GetDailyPoints(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error getting daily activity points: %v", err))
		return
	}
	respondWithJson(w, 200, DatabaseDailyPointsToDailyPoints(points))
}
