package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type LoginUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func (apiConf *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could'nt decode parameter")
		return
	}

	user, err := apiConf.DB.LoginUser(params.Email, params.Password)
	if err != nil {
		respondWithJSON(w, http.StatusUnauthorized, "Couldn't authorize user")
		return
	}

	defaultExpiration := 60 * 60 * 24
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = defaultExpiration
	} else if params.ExpiresInSeconds > defaultExpiration {
		params.ExpiresInSeconds = defaultExpiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(params.ExpiresInSeconds) * time.Second)),
		Subject:   strconv.Itoa(user.ID),
	})

	signedToken, err := token.SignedString([]byte(apiConf.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't sign token")
	}

	respondWithJSON(w, http.StatusOK, LoginUser{
		ID:    user.ID,
		Email: user.Email,
		Token: signedToken,
	})
}
