package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (apiConf *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, " ")
	if len(splitToken) < 2 || splitToken[0] != "Bearer" {
		respondWithError(w, http.StatusForbidden, "bad token")
		return
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
		respondWithError(w, http.StatusForbidden, "Token is expired")
		return
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		respondWithError(w, http.StatusForbidden, "invalid token")
		return
	}
	if issuer == string(RefreshTokenIssuer) {
		respondWithError(w, http.StatusForbidden, "Use refresh token to delete chirp")
		return
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid token")
		return
	}

	authorId, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Couldn't delete chirp")
		return
	}

	chirpIDString := req.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Invalid chirp ID")
		return
	}

	err = apiConf.DB.DeleteChirp(authorId, chirpID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Couldn't delete chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})

}
