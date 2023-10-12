package main

import (
	"database/sql"
	"fmt"
	"log"
)

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