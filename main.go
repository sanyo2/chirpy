package main

import (
	"fmt"
	"net/http"
)

type apiHandler struct{}

func main() {
	mux := http.NewServeMux()
	server := http.Server{}
	server.Handler = mux
	server.Addr = ":8080"
	mux.Handle("/", http.FileServer(http.Dir(".")))

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
