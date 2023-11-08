// Package main contains the main entry point for the Go application
package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserShare represents a user's sharing permissions for a note.
type UserShare struct {
	NoteID int `json:"note_id"`
	Username sql.NullString `json:"Username"`
	Privileges sql.NullString `json:"Privileges"`
}

// Note represents a note in the application.
type Note struct {
	// (Field descriptions for the Note type)
	ID                 int    `json:"id"`
	Title              string `json:"title"`
	NoteType           string `json:"note_type"`
	Description        string `json:"description"`
	NoteCreated        time.Time `json:"note_created"`
	TaskCompletionTime sql.NullString `json:"task_completion_time"`
	TaskCompletionDate sql.NullString `json:"task_completion_date"`
	NoteStatus         sql.NullString `json:"note_status"`
	NoteDelegation     sql.NullString `json:"note_delegation"`
	Owner              string    `json:"owner"`
	FTSText            sql.NullString `json:"fts_text"`
	Privileges         string
	SharedUsers		   []UserShare
}

// User represents a user in the application.
type User struct {
	Id string
	Username string `json:"username"`
	Password string `json:"password"`
	
}

// SearchResult represents a search result in the application.
type SearchResult struct {
    Count       int
    Description string
}

// readData reads data from a CSV file and returns it as a 2D string array.
func readData(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	// Skip the first line as it is the CSV header
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

// importData initializes the database schema and imports data from CSV files.
func (a *App) importData() error {

	// Drop foreign key constraints if they exist
	dropFKConstraintsSQL := `
	ALTER TABLE IF EXISTS user_shares DROP CONSTRAINT IF EXISTS user_shares_note_id_fkey;
	ALTER TABLE IF EXISTS user_shares DROP CONSTRAINT IF EXISTS user_shares_username_fkey;
	ALTER TABLE IF EXISTS notes DROP CONSTRAINT IF EXISTS notes_owner_fkey;
	
	`

	_, err := a.db.Exec(dropFKConstraintsSQL)
	if err != nil {
		log.Println("Error dropping foreign key constraints:", err)
	} else {
		log.Printf("Foreign key constraints dropped.")
	}

	log.Printf("Dropping existing tables...")

	// Drop tables if they exist
	dropTablesSQL := `
	DROP TABLE IF EXISTS users;
	DROP TABLE IF EXISTS user_shares;
	DROP TABLE IF EXISTS notes;
	
	`

	_, err = a.db.Exec(dropTablesSQL)
	if err != nil {
		log.Println("Error dropping tables:", err)
	} else {
		log.Printf("Tables notes and user_shares dropped.")
	}

	
    log.Printf("Creating tables...")

    // Create table as required, along with attribute constraints
    createTablesSQL := `
    CREATE TABLE IF NOT EXISTS "users" (
        username VARCHAR(50) UNIQUE PRIMARY KEY NOT NULL,
        password VARCHAR(255) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS "notes" (
        id SERIAL PRIMARY KEY NOT NULL,
        title VARCHAR(255) NOT NULL,
        noteType VARCHAR(255) NOT NULL,
        description VARCHAR(255) NOT NULL,
        noteCreated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        taskCompletionTime VARCHAR(255),
        taskCompletionDate VARCHAR(255),
        noteStatus VARCHAR(20),
        noteDelegation VARCHAR(50),
        owner VARCHAR(50),
        fts_text tsvector,
        FOREIGN KEY (owner) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS "user_shares" (
        note_id INTEGER NOT NULL,
        username VARCHAR(50) NOT NULL,
        privileges VARCHAR(20) NOT NULL,
        PRIMARY KEY (username, note_id),
        FOREIGN KEY (note_id) REFERENCES notes (id) ON UPDATE CASCADE ON DELETE CASCADE,
        FOREIGN KEY (username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
    );
`

    _, err = a.db.Exec(createTablesSQL)
	if err != nil {
		log.Println("Error creating tables:", err)
	} else {
		log.Printf("Tables notes, user_shares, and users created.")
	}

    log.Printf("Inserting data...")

	// Insert two users with hashed passwords
    hashedPasswordMydog7, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    _, err = a.db.Exec("INSERT INTO users(username, password) VALUES($1, $2)", "mydog7", hashedPasswordMydog7)
    if err != nil {
        log.Fatal(err)
    }

    hashedPasswordBIGCAT, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    _, err = a.db.Exec("INSERT INTO users(username, password) VALUES($1, $2)", "BIGCAT", hashedPasswordBIGCAT)
    if err != nil {
        log.Fatal(err)
    }

    // Prepare the notes insert query
    notesStmt, err := a.db.Prepare("INSERT INTO notes (title, noteType, description, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner, fts_text) VALUES($1,$2,$3,$4,$5,$6,$7,$8, to_tsvector('english', $9))")
    if err != nil {
        log.Fatal(err)
    }

	
   
	/*// Prepare the user_shares insert query
	userSharesStmt, err := a.db.Prepare("INSERT INTO user_shares (note_id, username, privileges) VALUES($1, $2, $3)")
	if err != nil {
		log.Fatal(err)
	}*/


    importDataFromCSV(a, "data/notes.csv", notesStmt, importNotesData)
    /*importDataFromCSV(a, "data/user_shares.csv", userSharesStmt, importUserSharesData)*/

    // Create a temp file to notify data imported (can use the database directly, but this is an example)
    file, err := os.Create("./imported")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    return nil // Return nil to indicate success
}

// importDataFromCSV reads data from a CSV file and imports it into the database using the provided statement.
func importDataFromCSV(a *App, fileName string, stmt *sql.Stmt, dataImporter func(*App, []string) error) {
    data, err := readData(fileName)
    if err != nil {
        log.Fatal(err)
    }

    for _, row := range data {
        err := dataImporter(a, row)
        if err != nil {
            log.Fatal(err)
        }
    }
}

// importNotesData imports note data from a CSV row into the database.
func importNotesData(a *App, row []string) error {
    title := row[0]
    noteType := row[1]
    description := row[2]
    taskCompletionTime := row[3]
    taskCompletionDate := row[4]
    noteStatus := row[5]
    noteDelegation := row[6]
    owner := row[7]
    
    // Calculate fts_text using to_tsvector
    ftsText := fmt.Sprintf("%s %s %s %s %s %s %s %s", title, noteType, description, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner)

    _, err := a.db.Exec("INSERT INTO notes (title, noteType, description, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner, fts_text) VALUES($1,$2,$3,$4,$5,$6,$7,$8, to_tsvector('english', $9))", title, noteType, description, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner, ftsText)

    return err
}



/*
func importUserSharesData(a *App, row []string) error {
    // 1. Ensure Data Integrity
	// Verify that the data in the user_shares table aligns with the referenced tables (notes and users).
	
	// Drop foreign key constraints if they exist
	dropFKConstraintsSQL := `
	ALTER TABLE IF EXISTS user_shares DROP CONSTRAINT IF EXISTS user_shares_note_id_fkey;
	ALTER TABLE IF EXISTS user_shares DROP CONSTRAINT IF EXISTS user_shares_username_fkey;
	ALTER TABLE IF EXISTS notes DROP CONSTRAINT IF EXISTS notes_owner_fkey;
	
	`

	_, err := a.db.Exec(dropFKConstraintsSQL)
	if err != nil {
		log.Println("Error dropping foreign key constraints:", err)
	} else {
		log.Printf("Foreign key constraints dropped.")
	}
	

	
	

	// 3. Insert Data
	noteID, err := strconv.Atoi(row[0])
	if err != nil {
		log.Fatal(err)
	}

	_, err = a.db.Exec("INSERT INTO user_shares (note_id, username, privileges) VALUES($1, $2, $3)", noteID, row[1], row[2])
	if err != nil {
		log.Fatal(err)
	}

	

    return err
}
*/









	


