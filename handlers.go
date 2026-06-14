package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sanyo2/chirpy/internal/auth"
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

func (cfg *apiConfig) apiAddChirp(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.SECRETKEY)
	if userID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := apiChirpParams{}
	err = decoder.Decode(&params)

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
		UserID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp(entry))
}

func (cfg *apiConfig) apiGetChirps(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	log.Println(id)
	var values []Chirp

	if id == "" {
		entries, err := cfg.DBQueries.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Cannot get chirps", err)
			return
		}

		for _, entry := range entries {
			values = append(values, Chirp(entry))
		}

		respondWithJSON(w, http.StatusOK, values)

	} else {
		uuidVal, err := uuid.Parse(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Cannot parse UUID", err)
			return
		}
		entries, err := cfg.DBQueries.GetChirpByID(r.Context(), uuidVal)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Cannot get chirps", err)
			return
		}

		respondWithJSON(w, http.StatusOK, Chirp(entries))
	}
}

func (cfg *apiConfig) apiUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := apiUserRequest{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot parse JSON", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot hash password", err)
		return
	}

	entry, err := cfg.DBQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        entry.ID,
		CreatedAt: entry.CreatedAt,
		UpdatedAt: entry.UpdatedAt,
		Email:     entry.Email,
	})
}

func (cfg *apiConfig) apiUsersLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := apiUserRequest{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot parse JSON", err)
		return
	}

	entry, err := cfg.DBQueries.GetUserPasswordByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot get user", err)
		return
	}

	passwordOk, err := auth.CheckPasswordHash(params.Password, entry.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot hash password", err)
		return
	}

	if !passwordOk {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	expirationTime := time.Duration(60 * 60 * time.Second)
	if params.ExpiresInSeconds < expirationTime && params.ExpiresInSeconds != 0 {
		expirationTime = params.ExpiresInSeconds
	}
	userToken, err := auth.MakeJWT(entry.ID, cfg.SECRETKEY, expirationTime)
	fmt.Println("Token: " + userToken)
	fmt.Printf("\nExp time: %v\n", expirationTime)

	respondWithJSON(w, http.StatusOK, User{
		ID:        entry.ID,
		CreatedAt: entry.CreatedAt,
		UpdatedAt: entry.UpdatedAt,
		Email:     entry.Email,
		Token:     userToken,
	})
}
