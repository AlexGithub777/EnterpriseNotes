package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/icza/session"
	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
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
	role          string
}

func (a *App) Initialize() {
	a.bindport = "80"

	//check if a different bind port was passed from the CLI
	//os.Setenv("PORT", "8080")
	tempport := os.Getenv("PORT")
	if tempport != "" {
		a.bindport = tempport
	}

	if len(os.Args) > 1 {
		s := os.Args[1]

		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			log.Printf("Using port %s", s)
			a.bindport = s
		}
	}

	// Create a string that will be used to make a connection later
	// Note Password has been left out, which is best to avoid issues when using null password
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	log.Println("Connecting to PostgreSQL")
	log.Println(psqlInfo)
	db, err := sql.Open("pgx", psqlInfo)
	a.db = db
	//db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Println("Invalid DB arguments, or github.com/lib/pq not installed")
		log.Fatal(err)
	}

	// test connection
	err = a.db.Ping()
	if err != nil {
		log.Fatal("Connection to specified database failed: ", err)
	}

	log.Println("Database connected successfully")

	// create users table if not exists
    createUsersTable := `CREATE TABLE IF NOT EXISTS users(
        username TEXT UNIQUE PRIMARY KEY NOT NULL,
        password TEXT NOT NULL,
        role TEXT NOT NULL
    );`

    _, err = a.db.Exec(createUsersTable)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Users table checked/created successfully")

	// creating notes table
    createNotesTable := `CREATE TABLE IF NOT EXISTS notes ( 
    	id SERIAL PRIMARY KEY, 
    	title TEXT NOT NULL,
    	noteType TEXT NOT NULL, 
    	description TEXT NOT NULL, 
    	noteCreated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    	taskCompletionTime TEXT, 
    	taskCompletionDate TEXT, 
    	noteStatus TEXT, 
    	noteDelegation TEXT, 
    	owner TEXT NOT NULL,
    	fts_text tsvector, 
        FOREIGN KEY (owner) REFERENCES users (username)
    )` 
    
    _, err = a.db.Exec(createNotesTable)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Notes table checked/created successfully")

    // creating user_shares table
    createUserSharesTable := `CREATE TABLE IF NOT EXISTS user_shares (
        note_id INTEGER,
		username TEXT,
		privileges TEXT,
        PRIMARY KEY (username, note_id),
        FOREIGN KEY (note_id) REFERENCES notes (id) ON DELETE CASCADE,
		FOREIGN KEY (username) REFERENCES users (username) ON DELETE CASCADE
    )`
    
    _, err = a.db.Exec(createUserSharesTable)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("User Shares table checked/created successfully")

	//check data import status
	//_, err = os.Stat("./imported")
	//if os.IsNotExist(err) {
	//	log.Println("--- Importing demo data")
	//	a.importData()
	

	

	// Initialize the session manager - this is a global
	// For testing purposes, we want cookies to be sent over HTTP too (not just HTTPS)
	// refer to the auth.go for the authentication handlers using the sessions
	session.Global.Close()
	session.Global = session.NewCookieManagerOptions(session.NewInMemStore(), &session.CookieMngrOptions{AllowHTTP: true})

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// setup static content route - strip ./assets/assets/[resource]
	// to keep /assets/[resource] as a route
	staticFileDirectory := http.Dir("./statics/")
	staticFileHandler := http.StripPrefix("/statics/", http.FileServer(staticFileDirectory))
	a.Router.PathPrefix("/statics/").Handler(staticFileHandler).Methods("GET")
	a.Router.HandleFunc("/", a.indexHandler).Methods("GET")
	a.Router.HandleFunc("/login", a.loginHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/logout", a.logoutHandler).Methods("GET")
	a.Router.HandleFunc("/register", a.registerHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/list", a.listHandler).Methods("GET")
	a.Router.HandleFunc("/create", a.createHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/update", a.updateHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/delete", a.deleteHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/share", a.shareHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/search", a.searchNotesHandler).Methods("POST", "GET")
	/*a.Router.HandleFunc("/share-note", a.shareNoteHandler).Methods("POST")*/
	a.Router.HandleFunc("/remove-shared-note", a.removeSharedNoteHandler).Methods("POST")

	log.Println("Routes established")
}

func (a *App) Run(addr string) {
	if addr != "" {
		a.bindport = addr
	}

	log.Printf("Starting HTTP service on %s", a.bindport)
	// setup HTTP on gorilla mux for a gracefull shutdown
	srv := &http.Server{
		Addr: "0.0.0.0:" + a.bindport,

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
