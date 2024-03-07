package main

import (
	"github.com/go-chi/chi/v5"
)

func getAdminRouter(cf *apiConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/metrics", cf.handlerViewHitCount)
	return r
}