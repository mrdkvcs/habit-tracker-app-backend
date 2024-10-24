package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"
)

var (
	tickers = make(map[uuid.UUID]*time.Ticker)
	mu      sync.Mutex
)

func (apiCfg *apiConfig) SetProductivityGoal(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		GoalDate   string `json:"goal_date"`
		GoalPoints int32  `json:"goal_points"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error decoding request: %v", err))
		return
	}
	layout := time.RFC3339
	parsedGoalDate, err := time.Parse(layout, params.GoalDate)
	err = apiCfg.DB.SetProductivityGoal(r.Context(), database.SetProductivityGoalParams{
		UserID:     user.ID,
		GoalDate:   parsedGoalDate,
		GoalPoints: params.GoalPoints,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error setting productivity goal: %v", err))
		return
	}
	goalPoints = params.GoalPoints
	respondWithJson(w, 200, "Productivity goal successfully set")
	startGoalTracker(user.ID, user.Email)
}

func startGoalTracker(userId uuid.UUID, userEmail string) {
	currentTime := time.Now()
	currentDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	if ticker, exists := tickers[userId]; exists {
		ticker.Stop()
	}
	ticker := time.NewTicker(3 * time.Hour)
	mu.Lock()
	tickers[userId] = ticker
	mu.Unlock()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
				if tDay.After(currentDay) {
					ticker.Stop()
					mu.Lock()
					delete(tickers, userId)
					mu.Unlock()
					return
				}
				sendUserReminder(userEmail)
			case <-stopChan:
				ticker.Stop()
				mu.Lock()
				delete(tickers, userId)
				mu.Unlock()
				return
			}
		}
	}()
}
func sendUserReminder(userEmail string) {
	sendEmail(userEmail, goalPoints, totalPoints)
}

func sendEmail(userEmail string, goalPoints, totalPoints int32) {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	from := os.Getenv("SMTP_USER")
	if from == "" {
		fmt.Println("Couldnt get username from .env file")
	}
	password := os.Getenv("SMTP_PASSWORD")
	if password == "" {
		fmt.Println("Couldnt get password from .env file")
	}
	body := ""
	to := userEmail
	subject := "Productivity Goal Reminder"
	body = fmt.Sprintf("Hello,\n\nYou are currently  below your productivity goal for the day.\nYour total daily points : %d\nYour goal points: %d\n\nKeep pushing to reach your target!\n\nBest regards,\nYour Productivity Tracker",
		totalPoints, goalPoints)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		fmt.Printf("error sending email: %v", err)
	} else {
		fmt.Println("Email sent succesfully")
	}
}
