package server

import (
	"database/sql"
	"log"
	"net/http"
)

func Server(database *sql.DB) *http.Server {
	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	srv := &http.Server{
		Addr:    "localhost" + ":" + "8081",
		Handler: mux,
	}

	log.Println("Running server at", srv.Addr)
	return srv
}
