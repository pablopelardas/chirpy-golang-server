package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir(filepathRoot))

	mux.Handle("/app/", http.StripPrefix("/app/", fs))
	mux.HandleFunc("/healthz", handlerReadiness)

	corsMux := middlewareCors(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}
	log.Print("Serving files from " + filepathRoot + " on port " + port)
	log.Fatal(srv.ListenAndServe())
	

}