package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vnsonvo/chirpy-go/internal/database"
)

type apiConfig struct {
	fileServerHits int
	DB             *database.DB
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	apiConf := apiConfig{fileServerHits: 0,

		DB: db}
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

func apiRouter(apiConf *apiConfig) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", handlerReadiness)
	r.Get("/reset", apiConf.handlerReset)
	r.Post("/users", apiConf.handlerCreateUser)
	r.Route("/chirps", func(r chi.Router) {
		r.Get("/", apiConf.handlerChirpsRetrieve)
		r.Post("/", apiConf.handlerChirp)
		r.Route("/{chirpID}", func(r chi.Router) {
			r.Get("/", apiConf.handlerChirpsGet)
		})
	})

	return r
}
