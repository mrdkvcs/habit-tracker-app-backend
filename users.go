package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	IsGoogle bool      `json:"is_google"`
	jwt.RegisteredClaims
}

type jwtTokenResponse struct {
	Token string `json:"token"`
}

func (apiCfg *apiConfig) googleCallback(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Code string `json:"code"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt decode request body: %s", err))
		return
	}

	token, err := oauthConfig.Exchange(r.Context(), params.Code)
	if err != nil {
		respondWithError(w, 400, "Couldnt exchange code for token")
		return
	}

	client := oauthConfig.Client(r.Context(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	var userInfo map[string]interface{}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&userInfo)
	if err != nil {
		respondWithError(w, 400, "Couldnt parse user info")
		return
	}

	googleId := userInfo["sub"].(string)
	userEmail := userInfo["email"].(string)
	userName := userInfo["name"].(string)

	existingUser, err := apiCfg.DB.GetUserByGoogleId(r.Context(), sql.NullString{String: googleId, Valid: true})
	if err == sql.ErrNoRows {
		user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
			ID:           uuid.New(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			Username:     userName,
			Email:        userEmail,
			PasswordHash: sql.NullString{String: "", Valid: false},
			GoogleID:     sql.NullString{String: googleId, Valid: true},
		})
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Couldnt create google user: %s", err))
			return
		}

		jwtToken, err := generateJWTForGoogleUser(user.ID, user.Email, user.Username)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Error in generating JWT token for google user: %s", err))
			return
		}
		respondWithJson(w, 200, jwtTokenResponse{Token: jwtToken})
		return
	} else if err != nil {
		respondWithError(w, 500, fmt.Sprintf("DB Error: Couldnt get user by google id: %s", err))
		return
	}
	jwtToken, err := generateJWTForGoogleUser(existingUser.ID, existingUser.Email, existingUser.Username)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error in generating JWT token for google user: %s", err))
		return
	}
	respondWithJson(w, 200, jwtTokenResponse{Token: jwtToken})
}

func generateJWTForGoogleUser(userId uuid.UUID, userEmail string, userName string) (string, error) {
	claims := Claims{
		UserID:   userId,
		Email:    userEmail,
		Username: userName,
		IsGoogle: true,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
func generateJWTForRegularUser(userId uuid.UUID, userEmail string, userName string) (string, error) {
	claims := Claims{
		UserID:   userId,
		Email:    userEmail,
		Username: userName,
		IsGoogle: false,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func (apiCfg *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Email      string `json:"email"`
		SetDefault string `json:"set_default"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt parse json: %s", err))
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, 400, "Couldnt hash password")
		return
	}
	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Username:     params.Username,
		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		Email:        params.Email,
		GoogleID:     sql.NullString{String: "", Valid: false},
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt create user: %s", err))
		return
	}
	if params.SetDefault == "true" {
		err := apiCfg.DB.SetDefaultActivities(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Couldnt set default activities: %s", err))
			return
		}
	}
	jwtToken, err := generateJWTForRegularUser(user.ID, user.Email, user.Username)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in generating JWT token for user: %s", err))
		return
	}
	respondWithJson(w, 200, jwtTokenResponse{Token: jwtToken})
}

func (apiCfg *apiConfig) LogInUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt parse json: %s", err))
		return
	}
	user, err := apiCfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err == sql.ErrNoRows {
		respondWithError(w, 401, "Error when logging in : Invalid email or password")
		return
	} else if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error when logging in : %s", err))
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(params.Password))
	if err != nil {
		respondWithError(w, 401, "Error when logging in : Invalid email or password")
		return
	}
	jwtToken, err := generateJWTForRegularUser(user.ID, user.Email, user.Username)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error in generating JWT token for user: %s", err))
		return
	}
	respondWithJson(w, 200, jwtTokenResponse{Token: jwtToken})
}

func (apiCfg *apiConfig) GetUserByEmail(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJson(w, 200, databaseUserToUser(user))
}

func (apiCfg *apiConfig) GetUsers(w http.ResponseWriter, r *http.Request) {
	teamId := r.URL.Query().Get("teamid")
	userId := r.URL.Query().Get("userid")
	parsedTeamID, err := uuid.Parse(teamId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt parse team id: %s", err))
		return
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt parse user id: %s", err))
		return
	}
	queryParam := r.URL.Query().Get("q")
	users, err := apiCfg.DB.GetUsers(r.Context(), database.GetUsersParams{
		TeamID:   parsedTeamID,
		ID:       parsedUserID,
		Username: fmt.Sprintf("%%%s%%", queryParam),
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt get users: %s", err))
		return
	}
	respondWithJson(w, 200, databaseUsersToUsers(users))
}
