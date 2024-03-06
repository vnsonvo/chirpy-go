package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
}

func (config *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", config.fileServerHits)))
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

	apiConf := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiConf.middleWareMetricIncrement(
		http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot)))))

	mux.HandleFunc("/healthz", handlerReadiness)
	mux.HandleFunc("/metrics", apiConf.handlerMetrics)
	mux.HandleFunc("/reset", apiConf.handlerReset)

	corsMux := middlewareCors(mux)

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
