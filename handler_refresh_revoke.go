package main

import (
	"net/http"
	"strings"
)

func (apiConf *apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, " ")
	if len(splitToken) < 2 || splitToken[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "bad token")
		return
	}

	reqToken = splitToken[1]

	err := apiConf.DB.RevokeToken(reqToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session")
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
