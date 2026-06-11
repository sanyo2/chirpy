package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sanyo2/chirpy/internal/database"
)

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
	<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
	</body>
	</html>`, cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) apiReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits.Swap(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if cfg.PLATFORM != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.DBQueries.DeleteUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting users"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileServerHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(res, req)
	})
}

func (cfg *apiConfig) apiChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := apiChirpParams{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot parse JSON", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	params.Body = replaceBadWords(params.Body)

	entry, err := cfg.DBQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp(entry))
}

func (cfg *apiConfig) apiUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := apiUserRequest{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot parse JSON: %v", err)
		return
	}

	entry, err := cfg.DBQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create user: %v", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User(entry))
}
