package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)
func (c *apiConfig) handlePostUsers(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	type requestBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type returnBody struct {
		Id int `json:"id"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}
	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error reading body")
		return
	}
	rBody := requestBody{}
	err = json.Unmarshal(dat, &rBody)
	if err != nil {
		log.Printf("Error unmarshalling JSON %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error unmarshalling JSON")
		return	
	}
	hasPassword := len(rBody.Password) > 0
	hasEmail := len(rBody.Email) > 0

	if !hasPassword || !hasEmail {
		log.Printf("Email and password are required")
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// save to file database.json
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rBody.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user, err := c.DB.CreateUser(rBody.Email, string(hashedPassword))
	if err != nil {
		log.Printf("Error creating chirp %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}


	// respond with id and cleaned body
	respondWithJSON(w, http.StatusCreated, returnBody{
		Id: user.ID,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (c *apiConfig) handleGetUsers(w http.ResponseWriter, r *http.Request){
	users, err := c.DB.GetUsers()
	if err != nil {
		log.Printf("Error getting users %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting users")
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (c *apiConfig) handleGetUser(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")

	user, err := c.DB.GetUser(id)
	if err != nil {
		log.Printf("Error getting user %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting user")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}


type MyCustomClaims struct {
    Email string `json:"email"`
    Id int `json:"id"`
    jwt.StandardClaims
}

func (c *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	type requestBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type returnBody struct {
		Id int `json:"id"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error reading body")
		return
	}
	rBody := requestBody{}
	err = json.Unmarshal(dat, &rBody)
	if err != nil {
		log.Printf("Error unmarshalling JSON %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error unmarshalling JSON")
		return	
	}
	hasPassword := len(rBody.Password) > 0
	hasEmail := len(rBody.Email) > 0

	if !hasPassword || !hasEmail {
		log.Printf("Email and password are required")
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := c.DB.GetUserByEmail(rBody.Email)
	if err != nil {
		log.Printf("Error getting user %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting user")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rBody.Password))
	if err != nil {
		log.Printf("Error comparing password %s", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	// create token


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyCustomClaims{
		user.Email,
		user.ID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
			Issuer: "chirpy-access",
			// time now utc
			IssuedAt: time.Now().Unix(),
			Subject: fmt.Sprint(user.ID),
		},
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60).Unix(),
		Issuer: "chirpy-refresh",
		IssuedAt: time.Now().Unix(),
		Subject: fmt.Sprint(user.ID),
	})

	tokenString, err := token.SignedString([]byte(c.jwtSecret))
	refreshTokenString, err := refreshToken.SignedString([]byte(c.jwtSecret))
	if err != nil {
		log.Printf("Error signing token %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error signing token")
		return
	}

	// respond with id and cleaned body
	respondWithJSON(w, http.StatusOK, returnBody{
		Id: user.ID,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
		Token: tokenString,
		RefreshToken: refreshTokenString,
	})
}

func (c *apiConfig) handlePutUser(w http.ResponseWriter, r *http.Request){
	// get ID from jwt header Authorization

	token := r.Header.Get("Authorization")
	if token == "" {
		log.Printf("No token provided")
		respondWithError(w, http.StatusUnauthorized, "No token provided")
		return
	}
	// strip bearer
	token = token[7:]
	claims := &MyCustomClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.jwtSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error parsing token")
		return
	}
	if !tkn.Valid {
		log.Printf("Token is not valid")
		respondWithError(w, http.StatusUnauthorized, "Token is not valid")
		return
	}
	if claims.Issuer == "chirpy-refresh"{
		log.Printf("Trying to access with refreshToken")
		respondWithError(w, http.StatusUnauthorized, "Token is not valid")
	}
	id := fmt.Sprint(claims.Id)
	log.Printf("ID PUT: %s", id)
	
	defer r.Body.Close()
	type requestBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type returnBody struct {
		Id int `json:"id"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}
	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error reading body")
		return
	}
	rBody := requestBody{}
	err = json.Unmarshal(dat, &rBody)
	if err != nil {
		log.Printf("Error unmarshalling JSON %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error unmarshalling JSON")
		return	
	}
	hasPassword := len(rBody.Password) > 0
	hasEmail := len(rBody.Email) > 0

	if !hasPassword || !hasEmail {
		log.Printf("Email and password are required")
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// save to file database.json
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rBody.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user, err := c.DB.UpdateUser(id, rBody.Email, string(hashedPassword))
	if err != nil {
		log.Printf("Error creating user %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating user")
		return
	}

	respondWithJSON(w, http.StatusOK, returnBody{
		Id: user.ID,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (c *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request){
	// get ID from jwt header Authorization

	type returnBody struct{
		Token string `json:"token"`
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		log.Printf("No token provided")
		respondWithError(w, http.StatusUnauthorized, "No token provided")
		return
	}
	// strip bearer
	token = token[7:]
	claims := &jwt.StandardClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.jwtSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error parsing token")
		return
	}
	if !tkn.Valid {
		log.Printf("Token is not valid")
		respondWithError(w, http.StatusUnauthorized, "Token is not valid")
		return
	}
	if claims.Issuer != "chirpy-refresh"{
		log.Printf("Trying to refresh token with accessToken")
		respondWithError(w, http.StatusUnauthorized, "Token is not valid")
	}

	// check 
	revoked, err := c.DB.CheckIfTokenRevoked(tkn.Raw)
	if err != nil {
		log.Printf("Error checking if token is revoked %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error checking if token is revoked")
		return
	}
	if revoked{
		log.Printf("Token is revoked")
		respondWithError(w, http.StatusUnauthorized, "Token is revoked")
		return
	}
	

	user, err := c.DB.GetUser(claims.Subject)
	if err != nil {
		log.Printf("Error getting user %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting user")
		return
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, MyCustomClaims{
		user.Email,
		user.ID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
			Issuer: "chirpy-access",
			// time now utc
			IssuedAt: time.Now().Unix(),
			Subject: fmt.Sprint(user.ID),
		},
	})
	tokenString, err := newToken.SignedString([]byte(c.jwtSecret))
	if err != nil {
		log.Printf("Error signing token %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error signing token")
		return
	}

	// respond with id and cleaned body
	respondWithJSON(w, http.StatusOK, returnBody{
		Token: tokenString,
	})
}

func (c *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request){
	// get ID from jwt header Authorization

	token := r.Header.Get("Authorization")
	if token == "" {
		log.Printf("No token provided")
		respondWithError(w, http.StatusUnauthorized, "No token provided")
		return
	}
	// strip bearer
	token = token[7:]
	claims := &jwt.StandardClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.jwtSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error parsing token")
		return
	}
	if !tkn.Valid {
		log.Printf("Token is not valid")
		respondWithError(w, http.StatusUnauthorized, "Token is not valid")
		return
	}
	if claims.Issuer != "chirpy-refresh"{
		log.Printf("Trying to refresh token with accessToken")
		respondWithError(w, http.StatusUnauthorized, "Token is not valid")
	}

	// check 
	revoked, err := c.DB.RevokeToken(tkn.Raw)
	if err != nil {
		log.Printf("Error revoking token %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error revoking token")
		return
	}
	if !revoked{
		log.Printf("Token is not revoked")
		respondWithError(w, http.StatusUnauthorized, "Token is not revoked")
		return
	}
	respondWithJSON(w, http.StatusOK, "Token revoked")
}