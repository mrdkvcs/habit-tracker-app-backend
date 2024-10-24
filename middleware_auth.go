package main

import (
	"github.com/mrdkvcs/go-base-backend/auth"
	"github.com/mrdkvcs/go-base-backend/internal/database"
	"net/http"
)

type authHandler func(http.ResponseWriter, *http.Request, database.User)

func (apiCfg *apiConfig) middlewareAuth(handler authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetApiKey(r.Header)
		if err != nil {
			respondWithError(w, 401, "Auth error")
			return
		}
		user, err := apiCfg.DB.GetUserByApikey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, 400, "Couldnt get user")
			return
		}
		handler(w, r, user)
	}
}
