package main

import (
	"encoding/json"
	"net/http"
)

const webhook_type string = "user.upgraded"

func (apiConf *apiConfig) handlerChirpWebhook(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		}
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
