package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	const filepathRoot = "."
	const port = "8080"
	
	apiConfig := apiConfig{fileserverHitCount: 0, filepathRoot: filepathRoot}
	fsHandler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(apiConfig.filepathRoot))))

	r := chi.NewRouter()
	
	// r.Mount("/app", getAppRouter(&apiConfig))
	r.Mount("/api", getApiRouter(&apiConfig))
	r.Mount("/admin", getAdminRouter(&apiConfig))
	r.Handle("/app", fsHandler)
	r.Handle("/app/*", fsHandler)
	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	log.Print("Serving files from " + filepathRoot + " on port " + port)
	log.Fatal(srv.ListenAndServe())
	

}