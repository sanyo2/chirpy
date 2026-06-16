package main

import (
	"database/sql"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sanyo2/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	DBURL          string
	PLATFORM       string
	SECRETKEY      string
	POLKAKEY       string
	DB             *sql.DB
	DBQueries      *database.Queries
}

type apiChirpParams struct {
	Body             string    `json:"body"`
	UserID           uuid.UUID `json:"user_id"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

type apiUserRequest struct {
	Email            string        `json:"email"`
	Password         string        `json:"password"`
	ExpiresInSeconds time.Duration `json:"expires_in_seconds"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type Token struct {
	Token string `json:"token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
