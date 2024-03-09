package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (apiConf *apiConfig) handlerChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxchirpLength = 140
	if len(params.Body) > maxchirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, " ")
	if len(splitToken) < 2 || splitToken[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "bad token")
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
		respondWithError(w, http.StatusUnauthorized, "Token is expired")
		return
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	if issuer == string(RefreshTokenIssuer) {
		respondWithError(w, http.StatusUnauthorized, "Use refresh token to create chirp")
		return
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	val := sanitizeFunc(params.Body)
	idVal, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	chirp, err := apiConf.DB.CreateChirp(val, idVal)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:       chirp.ID,
		Body:     chirp.Body,
		AuthorID: chirp.AuthorID,
	})

}

func sanitizeFunc(val string) string {
	var profane = map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	str := strings.Split(val, " ")
	for i, v := range str {
		key := strings.ToLower(v)
		if _, ok := profane[key]; ok {
			str[i] = "****"
		}
	}
	return strings.Join(str, " ")
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
