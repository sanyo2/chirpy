package main

import (
	"encoding/json"
	"fmt"
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
	fmt.Println("apiAddChirp")
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
	fmt.Println("apiGetChirps")
	id := r.PathValue("chirpID")

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

func (cfg *apiConfig) apiDeleteChirps(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	userID, err := auth.ValidateJWT(authToken, cfg.SECRETKEY)
	if userID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if id == uuid.Nil || err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp ID is empty", err)
		return
	}

	entries, err := cfg.DBQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Cannot get chirps", err)
		return
	}

	if entries.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Chirp update forbidden for this user", err)
		return
	}

	err = cfg.DBQueries.DeleteChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot delete chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) apiUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("apiUsers")
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

func (cfg *apiConfig) apiChangeUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	userID, err := auth.ValidateJWT(authToken, cfg.SECRETKEY)
	if userID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := apiUserRequest{}
	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot parse JSON", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot hash password", err)
		return
	}

	entry, err := cfg.DBQueries.UpdateEmailPasswordByID(r.Context(), database.UpdateEmailPasswordByIDParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot update user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        entry.ID,
		CreatedAt: entry.CreatedAt,
		UpdatedAt: entry.UpdatedAt,
		Email:     entry.Email,
	})

}

func (cfg *apiConfig) apiUsersLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("apiUsersLogin")
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.DBQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.SECRETKEY,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	_, err = cfg.DBQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	fmt.Printf("apiUsersLogin:\nToken:%v\nRefreshToken:%v\n", accessToken, refreshToken)
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) apiRefresh(w http.ResponseWriter, r *http.Request) {
	fmt.Println("apiRefresh")
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	user, err := cfg.DBQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.SECRETKEY,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) apiRevoke(w http.ResponseWriter, r *http.Request) {
	fmt.Println("apiRevoke")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", err)
		return
	}

	err = cfg.DBQueries.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot revoke token", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
