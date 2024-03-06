package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileServerHits int
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

func (config *apiConfig) middleWareMetricIncrement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		config.fileServerHits++
		next.ServeHTTP(w, req)
	})
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	apiConf := apiConfig{fileServerHits: 0}
	r := chi.NewRouter()

	fsHandler := apiConf.middleWareMetricIncrement(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))

	r.Handle("/app", fsHandler)
	r.Handle("/app/*", fsHandler)
	r.Mount("/api", apiRouter(&apiConf))
	adminRoute := chi.NewRouter()
	adminRoute.Get("/metrics", apiConf.handlerMetrics)
	r.Mount("/admin", adminRoute)

	// mux := http.NewServeMux()
	// mux.Handle("/app/", apiConf.middleWareMetricIncrement(
	// 	http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot)))))
	// mux.HandleFunc("/healthz", handlerReadiness)
	// mux.HandleFunc("/metrics", apiConf.handlerMetrics)
	// mux.HandleFunc("/reset", apiConf.handlerReset)
	// corsMux := middlewareCors(mux)

	corsMux := middlewareCors(r)

	var server = &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerValidate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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

	val := sanitizeFunc(params.Body)

	respondWithJSON(w, http.StatusOK, returnVals{CleanedBody: val})
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

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	if statusCode > 499 {
		log.Printf("Responding with 5xx error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, statusCode, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}

func apiRouter(apiConf *apiConfig) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", handlerReadiness)
	r.Get("/reset", apiConf.handlerReset)
	r.Post("/validate_chirp", handlerValidate)
	return r
}
