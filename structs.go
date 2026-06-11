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
	DB             *sql.DB
	DBQueries      *database.Queries
}

type apiChirpParams struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type apiUserRequest struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
