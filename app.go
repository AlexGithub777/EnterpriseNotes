package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/icza/session"
	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
)



func (a *App) Initialize() {
    // Create the database connection
    db, err := setupDatabase()
    if err != nil {
        log.Fatal(err)
    }
    a.db = db

    a.importData()

    // Initialize the session manager
    session.Global.Close()
    session.Global = session.NewCookieManagerOptions(session.NewInMemStore(), &session.CookieMngrOptions{AllowHTTP: true})

    a.Router = mux.NewRouter()
    a.initializeRoutes()

	// Set the bindport directly here
    a.bindport = "8080"  // Default value

    // Check if a different bind port was passed from the CLI
    tempport := os.Getenv("PORT")
    if tempport != "" {
        a.bindport = tempport
    }
}


func (a *App) Run(addr string) {
	if addr != "" {
		a.bindport = addr
	}

	// get the local IP that has Internet connectivity
	ip := GetOutboundIP()

	log.Printf("Starting HTTP service on http://%s:%s", ip, a.bindport)
	// setup HTTP on gorilla mux for a gracefull shutdown
	srv := &http.Server{
		//Addr: "0.0.0.0:" + a.bindport,
		Addr: ip + ":" + a.bindport,

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router,
	}

	// HTTP listener is in a goroutine as its blocking
	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// setup a ctrl-c trap to ensure a graceful shutdown
	// this would also allow shutting down other pipes/connections. eg DB
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
