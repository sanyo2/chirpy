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
		DBURL:     os.Getenv("DB_URL"),
		PLATFORM:  os.Getenv("PLATFORM"),
		SECRETKEY: os.Getenv("SECRET_KEY"),
		POLKAKEY:  os.Getenv("POLKA_KEY"),
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
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCFG.apiGetChirps)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCFG.apiDeleteChirps)
	mux.HandleFunc("POST /api/users", apiCFG.apiUsers)
	mux.HandleFunc("PUT /api/users", apiCFG.apiChangeUsers)
	mux.HandleFunc("POST /api/login", apiCFG.apiUsersLogin)
	mux.HandleFunc("POST /api/refresh", apiCFG.apiRefresh)
	mux.HandleFunc("POST /api/revoke", apiCFG.apiRevoke)
	mux.HandleFunc("POST /api/polka/webhooks", apiCFG.apiPolkaWebhook)
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port %s", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
