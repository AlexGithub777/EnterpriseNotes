package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"time"
)

type UserShare struct {
	NoteID int `json:"note_id"`
	Username string `json: "username"`
	Privileges string `json: privileges`
}

type Note struct {
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
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int    `json:"role"`
}

type PageData struct {
    AllUsers []User 
}


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

func (a *App) importData() error {
	log.Printf("Creating tables...")

	// Create table as required, along with attribute constraints
	sql := `
	CREATE TABLE IF NOT EXISTS "notes" (
		id SERIAL PRIMARY KEY NOT NULL,
		title TEXT NOT NULL,
		noteType TEXT NOT NULL,
		description TEXT NOT NULL,
		noteCreated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		taskCompletionTime TEXT,
		taskCompletionDate TEXT,
		noteStatus TEXT,
		noteDelegation TEXT,
		owner TEXT,
		fts_text tsvector
		
	);

	
	CREATE TABLE IF NOT EXISTS "user_shares" (
		note_id INTEGER,
		username TEXT,
		privileges TEXT,
		PRIMARY KEY (username, note_id)
		
		
		
	);

	
	CREATE TABLE IF NOT EXISTS "users" (
		username TEXT UNIQUE PRIMARY KEY NOT NULL,
		password TEXT NOT NULL,
		role INTEGER NOT NULL
	);
	`

	_, err := a.db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Tables notes,user_shares and users created.")

	log.Printf("Inserting data...")

	// Prepare the notes insert query
	notesStmt, err := a.db.Prepare("INSERT INTO notes (title, noteType, description, owner) VALUES($1,$2,$3,$4)")
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the user_shares insert query
	userSharesStmt, err := a.db.Prepare("INSERT INTO user_shares (user_id, note_id) VALUES($1,$2)")
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the users insert query
	usersStmt, err := a.db.Prepare("INSERT INTO users (username, password, role) VALUES($1,$2,$3)")
	if err != nil {
		log.Fatal(err)
	}

	defer usersStmt.Close()

	// Open the CSV file for importing in PG database
	notesData, err := readData("data/notes.csv")
	if err != nil {
		log.Fatal(err)
	}

	var n Note
	// Prepare the SQL for multiple inserts
	for _, data := range notesData {
		n.Title = data[0]
		n.NoteType = data[1]
		n.Description = data[2]
		n.Owner = data[3]

		_, err := notesStmt.Exec(n.Title, n.NoteType, n.Description, n.Owner)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Open the CSV file for importing in PG database
	noteSharedData, err := readData("data/user_shares.csv") // Declare noteSharedData
	if err != nil {
		log.Fatal(err)
	}

	var us UserShare
	// Prepare the SQL for multiple inserts
	for _, data := range noteSharedData {
		us.NoteID, _ = strconv.Atoi(data[1])

		_, err := userSharesStmt.Exec(us.NoteID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Open the CSV file for importing in PG database
	userData, err := readData("data/users.csv")
	if err != nil {
		log.Fatal(err)
	}

	var user User
	// Prepare the SQL for multiple inserts
	for _, data := range userData {
	
		if err != nil {
			log.Fatal(err)
		}
		
		user.Username = data[0]
		user.Password = data[1]
		roleInt, err := strconv.Atoi(data[2]) // Convert data[2] to an integer
		if err != nil {
			log.Fatal(err)
		}
		user.Role = roleInt

		_, insertErr := usersStmt.Exec(user.Username, user.Password, user.Role)
		if insertErr != nil {
			log.Fatal(insertErr)
		}
	}


	// Create a temp file to notify data imported (can use the database directly, but this is an example)
	file, err := os.Create("./imported")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	return nil // Return nil to indicate success
}






	


