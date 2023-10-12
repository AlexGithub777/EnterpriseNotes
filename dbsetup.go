package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/mux"
)

// PostgreSQl configuration if not passed as env variables
const (
	host     = "localhost" //127.0.0.1
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)

var (
	err  error
	wait time.Duration
)

type App struct {
	Router        *mux.Router
	db            *sql.DB
	bindport      string
	username      string
}

func setupDatabase() (*sql.DB, error) {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
    log.Println("Connecting to PostgreSQL")
    log.Println(psqlInfo)
    db, err := sql.Open("pgx", psqlInfo)
    if err != nil {
        log.Println("Invalid DB arguments, or github.com/lib/pq not installed")
        return nil, err
    }

    // test connection
    err = db.Ping()
    if err != nil {
        return nil, err
    }

    log.Println("Database connected successfully")
    return db, nil
}