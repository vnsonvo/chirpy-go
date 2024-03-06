package main

import (
	"fmt"
	"log"
	"net/http"

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

func apiRouter(apiConf *apiConfig) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", handlerReadiness)
	r.Get("/reset", apiConf.handlerReset)
	return r
}
