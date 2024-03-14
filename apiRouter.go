package main

import "github.com/go-chi/chi/v5"

func getApiRouter(cf *apiConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/healthz", handlerReadiness)
	r.Get("/metrics", cf.handlerGetHitCount)
	r.Get("/reset", cf.handlerResetHitCount)

	r.Get("/chirps/{id}", cf.handleGetChirp)
	r.Get("/chirps", cf.handleGetChirps)
	r.Post("/chirps", cf.handlePostChirp)
	r.Delete("/chirps/{id}", cf.handleDeleteChirp)
	// get chirps/id
	r.Get("/users/{id}", cf.handleGetUser)
	r.Get("/users", cf.handleGetUsers)
	r.Post("/users", cf.handlePostUsers)
	r.Put("/users", cf.handlePutUser)

	r.Post("/login", cf.handleLogin)
	r.Post("/refresh", cf.handleRefreshToken)
	r.Post("/revoke", cf.handleRevokeToken)

	r.Post("/polka/webhooks", cf.handlePolkaWebhook)
	return r
}