package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const webhook_type string = "user.upgraded"

func (apiConf *apiConfig) handlerChirpWebhook(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		}
	}

	reqApiKey := req.Header.Get("Authorization")
	splitApiKey := strings.Split(reqApiKey, " ")
	if len(splitApiKey) < 2 || splitApiKey[0] != "ApiKey" {
		respondWithError(w, http.StatusUnauthorized, "Bad api key")
		return
	}

	reqApiKey = splitApiKey[1]

	if reqApiKey != apiConf.polkaAPIKey {
		respondWithError(w, http.StatusUnauthorized, "Bad api key")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	if params.Event != webhook_type {
		respondWithJSON(w, http.StatusOK, struct{}{})
		return
	}

	err = apiConf.DB.UpgradeUser(params.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user")
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
