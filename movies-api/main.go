package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"movies-api/db"
	"movies-api/server"
)

func main() {
	//initialize database
	database, err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	//seed with sample data
	err = db.SeedTables(database)
	if err != nil {
		panic(err)
	}
	// Setup signal context
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	//call for server
	srv := server.Server(database)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("server error:", err)
			stop()
		}
	}()
	// wait for the signal
	<-mainCtx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	//stop acceptinng new requests and waits for in-flight ones to finish
	errShutdown := srv.Shutdown(shutdownCtx)
	if errShutdown != nil {
		fmt.Println("shutdown error:", errShutdown)
	}
}
