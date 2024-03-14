module github.com/pablopelardas/chirpy-golang-server

go 1.22.0

require (
	github.com/go-chi/chi/v5 v5.0.12 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/crypto v0.21.0 // indirect
)

require internal/database v1.0.0

replace internal/database => ./internal/database
