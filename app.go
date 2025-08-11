// Package main contains the main entry point for the Go application
package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
)

func (a *App) Initialize() {
	// Initialize sets up the application by:
	// - Creating the database connection
	// - Checking data import status and importing demo data if needed
	// - Setting up authentication
	// - Initializing the application's routes
	db, err := setupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	a.db = db

	// Check data import status
	_, err = os.Stat("./imported")
	if os.IsNotExist(err) {
		log.Println("--- Importing demo data")
		a.importData()
	}

	// Setup authentication (if applicable)
	a.setupAuth()

	// Initialize the application's routes
	a.initializeRoutes()

	// Set the default bind port
	a.bindport = "60"

	// Check if a different bind port was passed from the CLI via the PORT environment variable
	tempport := os.Getenv("PORT")
	if tempport != "" {
		a.bindport = tempport
	}
}

// Run starts the HTTP server on the specified address or the default if not provided.
// It handles graceful shutdown on interrupt signal.
func (a *App) Run(addr string) {
	if addr != "" {
		a.bindport = addr
	}

	// In Docker, bind to 0.0.0.0
	srv := &http.Server{
		Addr:         "0.0.0.0:" + a.bindport,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router,
	}

	log.Printf("Starting HTTP service on port %s", a.bindport)

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	log.Println("shutting HTTP service down")
	srv.Shutdown(ctx)
	log.Println("closing database connections")
	a.db.Close()
	log.Println("shutting down")
	os.Exit(0)
}
