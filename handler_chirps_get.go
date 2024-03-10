package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, req *http.Request) {

	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	id := -1
	author_id := req.URL.Query().Get("author_id")
	if author_id != "" {
		id, err = strconv.Atoi(author_id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't convert author id")
			return
		}
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		if id != -1 && dbChirp.AuthorID != id {
			continue
		}
		chirps = append(chirps, Chirp{
			ID:       dbChirp.ID,
			Body:     dbChirp.Body,
			AuthorID: dbChirp.AuthorID,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, req *http.Request) {
	chirpIDString := req.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:   dbChirp.ID,
		Body: dbChirp.Body,
	})
}
