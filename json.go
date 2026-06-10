package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type successResponse struct {
	Resp bool `json:"valid"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}

	type errorResponse struct {
		Err string `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{Err: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error Marshaling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
