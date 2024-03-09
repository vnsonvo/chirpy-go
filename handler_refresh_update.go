package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func refreshToken(tokenSecret, userIDString string) (string, error) {
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(AccessTokenIssuer),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
		Subject:   userIDString,
	})
	signedToken, err := newToken.SignedString([]byte(tokenSecret))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (apiConf *apiConfig) handlerRefreshToken(w http.ResponseWriter, req *http.Request) {
	type response struct {
		Token string `json:"token"`
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
	if issuer != string(RefreshTokenIssuer) {
		respondWithError(w, http.StatusUnauthorized, "Use access token to request refresh token")
		return
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	isRevoked, err := apiConf.DB.IsTokenRevoked(reqToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't check session")
		return
	}
	if isRevoked {
		respondWithError(w, http.StatusUnauthorized, "Refresh token is revoked")
		return
	}

	accessToken, err := refreshToken(apiConf.jwtSecret, userIDString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}
