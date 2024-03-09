package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (apiConf *apiConfig) HandlerUpdateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, " ")
	if len(splitToken) < 2 || splitToken[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "bad token")
	}

	reqToken = splitToken[1]

	// parse token
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		reqToken,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(apiConf.jwtSecret), nil },
	)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token is expired")
		return
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	if issuer == string(RefreshTokenIssuer) {
		respondWithError(w, http.StatusUnauthorized, "Use refresh token to update user info")
		return
	}
	id, err := token.Claims.GetSubject()

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	userId, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't convert id")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could'nt decode parameter")
		return
	}

	user, err := apiConf.DB.UpdateUser(userId, params.Email, params.Password)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		Email: user.Email,
		ID:    user.ID,
	})
}
