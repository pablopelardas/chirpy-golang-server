package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (c *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request){

	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	strApiKey := apiKey[7:]
	log.Printf("API Key: %s", strApiKey)
	if strApiKey != c.polkaApiKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	type requestBody struct{
		Event string `json:"event"`
		Data struct {
			UserId int `json:"user_id"`
		} `json:"data"`
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

	if rBody.Event != "user.upgraded"{
		w.WriteHeader(http.StatusOK)
		return
	}

	err = c.DB.UpgradeUserToChirpyRed(rBody.Data.UserId)

	if err != nil {
		if err.Error() == "user not found" {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error upgrading user")
		return
	}

	// return success
	w.WriteHeader(http.StatusOK)

}