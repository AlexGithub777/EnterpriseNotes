package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *App) initializeRoutes() {
	a.Router = mux.NewRouter()
	// setup static content route - strip ./assets/assets/[resource]
	// to keep /assets/[resource] as a route
	staticFileDirectory := http.Dir("./statics/")
	staticFileHandler := http.StripPrefix("/statics/", http.FileServer(staticFileDirectory))
	a.Router.PathPrefix("/statics/").Handler(staticFileHandler).Methods("GET")
	a.Router.HandleFunc("/", a.indexHandler).Methods("GET")
	a.Router.HandleFunc("/login", a.loginHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/user-logout", a.logoutHandler).Methods("GET")
	a.Router.HandleFunc("/register", a.registerHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/list", a.listHandler).Methods("GET")
	a.Router.HandleFunc("/create", a.createHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/update", a.updateHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/delete", a.deleteHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/share", a.shareHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/search", a.searchNotesHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/remove-shared-note", a.removeSharedNoteHandler).Methods("POST")
	a.Router.HandleFunc("/getSharedUsersForNote/{noteID:[0-9]+}", a.getSharedUsersForNoteHandler).Methods("GET")
	a.Router.HandleFunc("/getUnsharedUsersForNote/{noteID:[0-9]+}", a.getUnsharedUsersForNoteHandler).Methods("GET")
	a.Router.HandleFunc("/find/{noteID:[0-9]+}", a.findInNoteHandler).Methods("GET")
	a.Router.HandleFunc("/update-privileges", a.updatePrivilegesHandler).Methods("POST")
	a.Router.HandleFunc("/remove-delegation", a.removeDelegationHandler).Methods("POST")
	



	log.Println("Routes established")
}