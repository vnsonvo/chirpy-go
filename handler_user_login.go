package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type LoginUser struct {
	ID            int    `json:"id"`
	Email         string `json:"email"`
	Token         string `json:"token"`
	Refresh_Token string `json:"refresh_token"`
	IsChirpyRed   bool   `json:"is_chirpy_red"`
}

type TokenType string

const (
	RefreshTokenIssuer TokenType = "chirpy-refresh"
	AccessTokenIssuer  TokenType = "chirpy-acces"
)

func (apiConf *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(AccessTokenIssuer),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
		Subject:   strconv.Itoa(user.ID),
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(RefreshTokenIssuer),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * 24 * 60)),
		Subject:   strconv.Itoa(user.ID),
	})

	signedToken, err := token.SignedString([]byte(apiConf.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't sign access token")
	}

	signedRefreshToken, err := refreshToken.SignedString([]byte(apiConf.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldnt sign refresh token")
	}

	respondWithJSON(w, http.StatusOK, LoginUser{
		ID:            user.ID,
		Email:         user.Email,
		Token:         signedToken,
		Refresh_Token: signedRefreshToken,
		IsChirpyRed:   user.IsChirpyRed,
	})
}
