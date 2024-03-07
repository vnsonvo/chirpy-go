package main

import (
	"fmt"
	"net/http"
)

func (config *apiConfig) middleWareMetricIncrement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		config.fileServerHits++
		next.ServeHTTP(w, req)
	})
}

func (config *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>	
	`, config.fileServerHits)))
}
