package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jub0bs/cors"
	_ "github.com/lib/pq"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
	"os"
)

type apiConfig struct {
	DB *database.Queries
}

var db *sql.DB
var totalPoints int32 = 0
var goalPoints int32 = 0
var stopChan chan struct{} = make(chan struct{})
var api_key string = ""
var extractSystemInstruction = `
You are an assistant that extracts the core activity and duration from user input. Your goal is to interpret the user's activity and extract the duration in minutes. Before processing, translate the input into English to ensure accurate extraction. Focus on extracting the activity itself, and convert any time mentioned to the equivalent number of minutes. The output should be in the following structured format: Activity: <activity>, Duration: <duration>. The duration should be expressed as a number of minutes without any units or extra text.

Examples:

"I went running for 2 hours" → Activity: Running, Duration: 120
"Studied for 45 minutes" → Activity: Studying, Duration: 45
"Watched Netflix for 3 hours" → Activity: Watching Netflix, Duration: 180
"Cleaned the house for half an hour" → Activity: Cleaning, Duration: 30
"Spent 15 minutes scrolling on Instagram" → Activity: Scrolling on Instagram, Duration: 15
"Played chess for 1 hour" → Activity: Chess, Duration: 60 (This is a custom activity)

Ensure the input is translated into English first before extracting the activity and duration.
`

var compareSystemInstruction = `
You are an assistant that compares an extracted activity with activities stored in a database. Your task is to compare two activities provided in the format {extracted_activity}, {database_activity}, where:

The first activity is the one extracted from the user input.
The second activity is the one retrieved from the database.
Ensure that you translate both the extracted activity and the database activity to English and compare them that way. Your goal is to determine whether these two activities are associated or equivalent by taking into account possible synonyms or closely related terms.

Examples of associations between activities:

Running, Exercise
Studying, Learning
Watching Netflix, Watching Series
Cleaning the house, Household Chores
Scrolling on Instagram, Social Media Scrolling
The activity comparison should be flexible enough to recognize synonymous terms or variations in phrasing.

Only output "true" or "false". Do not explain or add any additional text—just "true" or "false."
`

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Couldnt get port number from .env file")
	}
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		fmt.Println("Could not find database url in .env file")
	}
	api_key = os.Getenv("OPENAI_API_KEY")
	if api_key == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		return
	}
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("Could not connect to database")
	}
	apiConfig := apiConfig{DB: database.New(db)}
	corsMw, err := cors.NewMiddleware(cors.Config{
		Origins:        []string{"http://localhost:5173", "http://localhost:5174"},
		Methods:        []string{"GET", "POST", "DELETE"},
		RequestHeaders: []string{"Authorization"},
	})
	if err != nil {
		fmt.Println("Could not create cors middleware")
	}
	corsMw.SetDebug(true)
	router := http.NewServeMux()
	router.HandleFunc("GET /ws", handleConnections)
	router.HandleFunc("POST /register", apiConfig.CreateUser)
	router.HandleFunc("POST /login", apiConfig.GetUserByEmail)
	router.HandleFunc("GET /user/{id}", apiConfig.GetUserApiKey)
	router.HandleFunc("GET /activities", apiConfig.middlewareAuth(apiConfig.GetActivites))
	router.HandleFunc("POST /activities", apiConfig.middlewareAuth(apiConfig.SetActivity))
	router.HandleFunc("DELETE /activities/{id}", apiConfig.DeleteActivity)
	router.HandleFunc("PUT /activities/{id}", apiConfig.EditActivity)
	router.HandleFunc("POST /activities/logs", apiConfig.middlewareAuth(apiConfig.SetActivityLog))
	router.HandleFunc("POST /activities/logs/specific", apiConfig.middlewareAuth(apiConfig.SetSpecificActivityLog))
	router.HandleFunc("GET /activities/logs/exist", apiConfig.middlewareAuth(apiConfig.CheckActivityLogExists))
	router.HandleFunc("GET /activities/daily/logs", apiConfig.middlewareAuth(apiConfig.GetDailyActivityLogs))
	router.HandleFunc("GET /dailypoints", apiConfig.middlewareAuth(apiConfig.GetDailyPoints))
	router.HandleFunc("POST /productivitystats", apiConfig.middlewareAuth(apiConfig.GetProductivityStats))
	router.HandleFunc("POST /productivitygoals", apiConfig.middlewareAuth(apiConfig.SetProductivityGoal))
	router.HandleFunc("POST /teams", apiConfig.middlewareAuth(apiConfig.CreateTeam))
	router.HandleFunc("GET /teams", apiConfig.middlewareAuth(apiConfig.GetUserTeams))
	router.HandleFunc("GET /teams/{teamid}", apiConfig.GetTeamInfo)
	router.HandleFunc("GET /teams/{teamid}/activities", apiConfig.GetTeamActivities)
	router.HandleFunc("POST /teams/{teamid}/roles", apiConfig.SetTeamRole)
	router.HandleFunc("GET /teams/{teamid}/roles", apiConfig.GetTeamRoles)
	router.HandleFunc("POST /teams/{teamid}/activities", apiConfig.SetTeamActivity)
	router.HandleFunc("GET /teams/{teamid}/ownership", apiConfig.middlewareAuth(apiConfig.IsUserTeamOwner))
	router.HandleFunc("GET /teams/{teamid}/user/activities", apiConfig.middlewareAuth(apiConfig.GetUserTeamActivities))
	router.HandleFunc("GET /users", apiConfig.GetUsers)
	router.HandleFunc("POST /teams/{teamid}/invitation", apiConfig.CreateTeamInvitation)
	router.HandleFunc("GET /user/invitations", apiConfig.middlewareAuth(apiConfig.GetTeamInvitations))
	router.HandleFunc("GET /user/invitations/count", apiConfig.middlewareAuth(apiConfig.GetInvitationsCount))
	router.HandleFunc("UPDATE /user/invitations/seen", apiConfig.middlewareAuth(apiConfig.SetInvitationsAsSeen))
	router.HandleFunc("POST /user/invitations/accept", apiConfig.middlewareAuth(apiConfig.AcceptTeamInvite))
	router.HandleFunc("DELETE /user/invitations/{invitationid}", apiConfig.DeclineTeamInvite)
	router.HandleFunc("GET /teams/{teamid}/members", apiConfig.middlewareAuth(apiConfig.GetTeamMembers))
	router.HandleFunc("POST /teams/{teamid}/roles/{membership_id}", apiConfig.SetMemberRoles)
	router.HandleFunc("GET /teams/{teamid}/roles/{membership_id}", apiConfig.GetNotAssignedRoles)
	go handleMessages()
	handler := corsMw.Wrap(router)
	fmt.Println("Server running on port: " + port)
	http.ListenAndServe(":"+port, handler)
}
