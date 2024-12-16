package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
	"strings"
)

type authHandler func(http.ResponseWriter, *http.Request, database.User)

func (apiCfg *apiConfig) middlewareAuth(handler authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			cookie, err := r.Cookie("token")
			if err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			respondWithJson(w, 401, "Unauthorized user , missing token")
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			respondWithJson(w, 401, "Unauthorized user , invalid token")
			return
		}

		if !token.Valid {
			respondWithJson(w, 401, "Unauthorized user , token expired")
			return
		}
		user, err := apiCfg.DB.GetUserByEmail(r.Context(), claims.Email)
		if err != nil {
			respondWithError(w, 500, fmt.Sprintf("Could not find user with email %s", claims.Email))
			return
		}
		handler(w, r, user)
	}
}
