package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/jub0bs/cors"
	_ "github.com/lib/pq"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"os"
)

type apiConfig struct {
	DB *database.Queries
}

var apiconfig apiConfig
var db *sql.DB
var stopChan chan struct{} = make(chan struct{})
var api_key string = ""
var oauthConfig *oauth2.Config
var jwtSecret string
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

	oauthClientId := os.Getenv("OAUTH_GOOGLE_CLIENT_ID")
	if oauthClientId == "" {
		fmt.Println("OAUTH_GOOGLE_CLIENT_ID environment variable is not set")
		return
	}

	oauthClientSecret := os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET")
	if oauthClientSecret == "" {
		fmt.Println("OAUTH_GOOGLE_CLIENT_SECRET environment variable is not set")
		return
	}

	oauthConfig = &oauth2.Config{
		ClientID:     oauthClientId,
		ClientSecret: oauthClientSecret,
		RedirectURL:  "http://localhost:5173/google/auth/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	if os.Getenv("JWT_SECRET") != "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	} else {
		fmt.Println("JWT_SECRET environment variable is not set")
		return
	}

	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("Could not connect to database")
	}

	apiconfig = apiConfig{DB: database.New(db)}
	corsMw, err := cors.NewMiddleware(cors.Config{
		Origins:        []string{"http://localhost:5173", "http://localhost:5174"},
		Methods:        []string{"GET", "POST", "DELETE", "PUT"},
		RequestHeaders: []string{"Authorization"},
	})
	if err != nil {
		fmt.Println("Could not create cors middleware")
	}
	corsMw.SetDebug(true)
	router := http.NewServeMux()
	router.HandleFunc("GET /ws", handleConnections)
	router.HandleFunc("POST /register", apiconfig.CreateUser)
	router.HandleFunc("POST /login", apiconfig.LogInUser)
	router.HandleFunc("POST /forgot-password", apiconfig.ForgotPasswordHandler)
	router.HandleFunc("POST /reset-password", apiconfig.ResetPasswordHandler)
	router.HandleFunc("GET /user", apiconfig.middlewareAuth(apiconfig.GetUserByEmail))
	router.HandleFunc("POST /google/auth/callback", apiconfig.googleCallback)
	router.HandleFunc("GET /activities", apiconfig.middlewareAuth(apiconfig.GetActivites))
	router.HandleFunc("POST /activities", apiconfig.middlewareAuth(apiconfig.SetActivity))
	router.HandleFunc("DELETE /activities/{id}", apiconfig.DeleteActivity)
	router.HandleFunc("PUT /activities/{id}", apiconfig.EditActivity)
	router.HandleFunc("POST /activities/logs", apiconfig.middlewareAuth(apiconfig.SetActivityLog))
	router.HandleFunc("POST /activities/logs/specific", apiconfig.middlewareAuth(apiconfig.SetSpecificActivityLog))
	router.HandleFunc("POST /activities/logs/new", apiconfig.middlewareAuth(apiconfig.SetNewActivity))
	router.HandleFunc("GET /activities/logs/exist", apiconfig.middlewareAuth(apiconfig.CheckActivityLogExists))
	router.HandleFunc("GET /activities/daily/logs", apiconfig.middlewareAuth(apiconfig.GetDailyActivityLogs))
	router.HandleFunc("GET /dailystats", apiconfig.middlewareAuth(apiconfig.GetDailyStats))
	router.HandleFunc("POST /productivitystats", apiconfig.middlewareAuth(apiconfig.GetProductivityStats))
	router.HandleFunc("POST /productivitygoals", apiconfig.middlewareAuth(apiconfig.SetProductivityGoal))
	router.HandleFunc("POST /suggestFeature", apiconfig.middlewareAuth(apiconfig.createSuggestFeature))
	router.HandleFunc("GET /suggestFeature", apiconfig.GetSuggestFeature)
	router.HandleFunc("PUT /suggestFeature/upvote/{id}", apiconfig.SetSuggestFeatureUpVote)
	router.HandleFunc("PUT /suggestFeature/downvote/{id}", apiconfig.SetSuggestFeatureDownVote)
	router.HandleFunc("POST /teams", apiconfig.middlewareAuth(apiconfig.CreateTeam))
	router.HandleFunc("GET /teams", apiconfig.middlewareAuth(apiconfig.GetUserTeams))
	router.HandleFunc("GET /teams/{teamid}", apiconfig.GetTeamInfo)
	router.HandleFunc("GET /teams/{teamid}/activities", apiconfig.GetTeamActivities)
	router.HandleFunc("POST /teams/{teamid}/roles", apiconfig.SetTeamRole)
	router.HandleFunc("GET /teams/{teamid}/roles", apiconfig.GetTeamRoles)
	router.HandleFunc("POST /teams/{teamid}/activities", apiconfig.SetTeamActivity)
	router.HandleFunc("GET /teams/{teamid}/ownership", apiconfig.middlewareAuth(apiconfig.IsUserTeamOwner))
	router.HandleFunc("GET /teams/{teamid}/user/activities", apiconfig.middlewareAuth(apiconfig.GetUserTeamActivities))
	router.HandleFunc("GET /users", apiconfig.GetUsers)
	router.HandleFunc("POST /teams/{teamid}/invitation", apiconfig.CreateTeamInvitation)
	router.HandleFunc("GET /user/invitations", apiconfig.middlewareAuth(apiconfig.GetTeamInvitations))
	router.HandleFunc("GET /user/invitations/count", apiconfig.middlewareAuth(apiconfig.GetInvitationsCount))
	router.HandleFunc("UPDATE /user/invitations/seen", apiconfig.middlewareAuth(apiconfig.SetInvitationsAsSeen))
	router.HandleFunc("POST /user/invitations/accept", apiconfig.middlewareAuth(apiconfig.AcceptTeamInvite))
	router.HandleFunc("DELETE /user/invitations/{invitationid}", apiconfig.DeclineTeamInvite)
	router.HandleFunc("GET /teams/{teamid}/members", apiconfig.middlewareAuth(apiconfig.GetTeamMembers))
	router.HandleFunc("POST /teams/{teamid}/roles/{membership_id}", apiconfig.SetMemberRoles)
	router.HandleFunc("GET /teams/{teamid}/roles/{membership_id}", apiconfig.GetNotAssignedRoles)
	go handleMessages()
	handler := corsMw.Wrap(router)
	fmt.Println("Server running on port: " + port)
	http.ListenAndServe(":"+port, handler)
}
