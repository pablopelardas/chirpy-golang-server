package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
    response, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.WriteHeader(code)
    w.Write(response)
    return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
    return respondWithJSON(w, code, map[string]string{"error": msg})
}

func getAccessTokenData(r *http.Request, jwtSecret string) (MyCustomClaims, error) {
    token := r.Header.Get("Authorization")
	if token == "" {
		log.Printf("No token provided")
		return MyCustomClaims{}, errors.New("No token provided")
	}
	// strip bearer
    token = token[7:]
    
    claims := &MyCustomClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token %s", err)
		return MyCustomClaims{}, err
	}
	if !tkn.Valid {
		log.Printf("Token is not valid")
        return MyCustomClaims{}, errors.New("Token is not valid")
	}
	if claims.Issuer != "chirpy-access"{
        log.Printf("Invalid issuer")
        return MyCustomClaims{}, errors.New("Invalid issuer")
	}
    return *claims, nil
}