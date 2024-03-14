package main

import (
	"flag"
	"internal/database"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaApiKey := os.Getenv("POLKA_KEY")
	const filepathRoot = "."
	const port = "8080"
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		log.Print("Debug mode enabled")
		// delete database.json
		_,err := database.DeleteDB("database.json")
		if err != nil {
			log.Fatal(err)
		}
	}

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}
	apiConfig := apiConfig{fileserverHitCount: 0, filepathRoot: filepathRoot, DB: db, jwtSecret: jwtSecret, polkaApiKey:polkaApiKey}
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