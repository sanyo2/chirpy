package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sanyo2/chirpy/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	const filepathRoot = "."
	const port = "8080"
	godotenv.Load()

	apiCFG := apiConfig{
		DBURL:    os.Getenv("DB_URL"),
		PLATFORM: os.Getenv("PLATFORM"),
	}

	db, err := sql.Open("postgres", apiCFG.DBURL)
	if err != nil {
		log.Fatal("DB Connection error")
	}
	apiCFG.DB = db
	apiCFG.DBQueries = database.New(apiCFG.DB)

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app",
		apiCFG.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handleReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCFG.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCFG.apiReset)
	mux.HandleFunc("POST /api/chirps", apiCFG.apiAddChirp)
	mux.HandleFunc("GET /api/chirps", apiCFG.apiGetChirps)
	mux.HandleFunc("POST /api/users", apiCFG.apiUsers)
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
