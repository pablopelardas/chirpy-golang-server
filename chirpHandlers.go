package main

import (
	"encoding/json"
	"internal/database"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)
func (c *apiConfig) handlePostChirp(w http.ResponseWriter, r *http.Request){
	// get token

	tokenClaims, err := getAccessTokenData(r, c.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	type requestBody struct {
		Body string `json:"body"`
	}
	type returnBody struct {
		Id int `json:"id"`
		Cleaned_body string `json:"body"`
		Author int `json:"author_id"`
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
	if len(rBody.Body) > 140 {
		log.Printf("Chirp too long")
		respondWithError(w, http.StatusBadRequest, "Chirp too long")
		return
	}
	badwords := []string{"kerfuffle", "sharbert", "fornax"}
	// clean the input of bad words case insensitive
	cleaned := rBody.Body
	for _, word := range strings.Split(cleaned, " ") {
		for _, badword := range badwords {
			if strings.ToLower(word) == badword {
				cleaned = strings.ReplaceAll(cleaned, word, "****")
			}
		}
	}

	// save to file database.json
	chirp, err := c.DB.CreateChirp(cleaned, tokenClaims.Id)
	if err != nil {
		log.Printf("Error creating chirp %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp")
		return
	}


	// respond with id and cleaned body
	respondWithJSON(w, http.StatusCreated, returnBody{
		Id: chirp.ID,
		Cleaned_body: chirp.Body,
		Author: chirp.Author,
	})
}

func (c *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request){
	// get from database
	dbChirps, err := c.DB.GetChirps()
	
	authorId := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")

	if err != nil {
		log.Printf("Error getting chirps %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps")
		return
	}
	chirps := []database.Chirp{}
	for _, chirp := range dbChirps {
		if authorId != "" {
			if strconv.Itoa(chirp.Author) != authorId {
				continue
			}
		}
		chirps = append(chirps, database.Chirp{
			ID: chirp.ID,
			Body: chirp.Body,
			Author: chirp.Author,
		})
	}
	if sortQuery == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	} else if sortQuery == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (c *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request){
	// get from database
	id := r.PathValue("id")
	dbChirp, err := c.DB.GetChirp(id)
	if err != nil {
		if err.Error() == "chirp not found" {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		log.Printf("Error getting chirp %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting chirp")
		return
	}
	respondWithJSON(w, http.StatusOK, dbChirp)
}

func (c *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request){
	// get from database
	tokenClaims, err := getAccessTokenData(r, c.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	id,err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid id")
		return
	}
	err = c.DB.DeleteChirp(id, tokenClaims.Id)
	if err != nil {
		if err.Error() == "chirp not found" {
			log.Printf("Chirp not found")
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		if err.Error() == "unauthorized" {
			log.Printf("Unauthorized")
			respondWithError(w, http.StatusForbidden, "Unauthorized")
			return
		}
		log.Printf("Error deleting chirp %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error deleting chirp")
		return
	}
	respondWithJSON(w, http.StatusOK, "Chirp deleted")

}