package main

import "github.com/go-chi/chi/v5"

func getApiRouter(cf *apiConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/healthz", handlerReadiness)
	r.Get("/metrics", cf.handlerGetHitCount)
	r.Get("/reset", cf.handlerResetHitCount)
	return r
}