package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // use pgx in database/sql mode

	"github.com/gorilla/mux"
)

// PostgreSQl configuration if not passed as env variables
const (
	host     = "notes-db"
	port     = 5433
	user     = "postgres"
	password = "postgres"
	dbname   = "notes"
)

var (
	err  error
	wait time.Duration
)

type App struct {
	Router   *mux.Router
	db       *sql.DB
	bindport string
	username string
}

func setupDatabase() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	log.Println("Connecting to PostgreSQL")
	log.Println(psqlInfo)
	os.Unsetenv("PGDATABASE")
	os.Unsetenv("PGHOST")
	os.Unsetenv("PGUSER")
	os.Unsetenv("PGPASSWORD")

	// Set env to db name notes
	os.Setenv("PGDATABASE", dbname)
	os.Setenv("PGHOST", host)
	os.Setenv("PGUSER", user)
	os.Setenv("PGPASSWORD", password)
	// Open a database connection using the pgx driver

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
