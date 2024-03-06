package main

import "net/http"

func (config *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	config.fileServerHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
