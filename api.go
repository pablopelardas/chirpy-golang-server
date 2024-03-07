package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHitCount int
	filepathRoot       string
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHitCount++
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) handlerGetHitCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", c.fileserverHitCount)))
}

func (c *apiConfig) handlerResetHitCount(w http.ResponseWriter, r *http.Request) {
	c.fileserverHitCount = 0
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Hit count reset"))
}

func (c *apiConfig) handlerViewHitCount(w http.ResponseWriter, r *http.Request) {
	template := fmt.Sprintf(`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	
	</html>`, c.fileserverHitCount)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(template))
}