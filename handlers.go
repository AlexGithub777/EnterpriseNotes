package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/icza/session"
)



func (a *App) listHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}

    sess := session.Get(r)
    username := "[guest]"

    if sess != nil {
        username = sess.CAttr("username").(string)
    }

    // Check for a message cookie
    cookie, err := r.Cookie("errorMessage")
    var message string
    if err == nil {
        message = cookie.Value

        // Delete the cookie
        deleteCookie := http.Cookie{Name: "errorMessage", MaxAge: -1, Path: "/list"}
        http.SetCookie(w, &deleteCookie)
    }


    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Retrieve all notes
    notes, err := a.retrieveNotes(username)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    // Retrieve all shared notes with privileges
    sharedNotes, err := a.retrieveSharedNotesWithPrivileges(username)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    // Get the list of all users
    allUsers, err := a.getAllUsers(username)
    if err != nil {
        // Handle the error appropriately (e.g., log it or show an error page)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Fetch the list of shared users for each note
    for i := range notes {
        sharedUsers, err := a.getSharedUsersForNote(notes[i].ID)
        if err != nil {
            checkInternalServerError(err, w)
            return
        }
        notes[i].SharedUsers = sharedUsers
    }

    // Pass the shared notes with privileges to the template
    data := struct {
        Username      string
        Notes         []Note
        AllUsers      []User
        SharedNotes   []Note
        Message string
        
    }{
        Username:      username,
        Notes:         notes,
        AllUsers:      allUsers,
        SharedNotes:   sharedNotes,
        Message: message,
        
    }

    t, err := template.New("list.html").Funcs(template.FuncMap{
		"formatTime": func(layout, value string) (string, error) {
			t, err := time.Parse(layout, value)
			if err != nil {
				return "", err
			}
			return t.Format("3:04 PM"), nil
		},
		"formatDate": func(layout, value string) (string, error) {
			t, err := time.Parse(layout, value)
			if err != nil {
				return "", err
			}
			return t.Format("02/01/2006"), nil
		},
	}).ParseFiles("tmpl/list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
	

    var buf bytes.Buffer
    err = t.Execute(&buf, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=UTF-8")
    buf.WriteTo(w)
}

func (a *App) getSharedUsersForNoteHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    noteIDStr, ok := vars["noteID"]
    if !ok {
        http.Error(w, "Missing noteID in URL", http.StatusBadRequest)
        return
    }

    noteID, err := strconv.Atoi(noteIDStr)
    if err != nil {
        http.Error(w, "Invalid noteID", http.StatusBadRequest)
        return
    }


    // Fetch the shared users for the given noteID
    sharedUsers, err := a.getSharedUsersForNote(noteID)
    if err != nil {
        http.Error(w, "Failed to fetching shared users: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Marshal the sharedUsers slice into JSON
    responseJSON, err := json.Marshal(sharedUsers)
    if err != nil {
        http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
        return
    }

    // Set the Content-Type header to indicate JSON response
    w.Header().Set("Content-Type", "application/json")

    // Write the JSON response to the HTTP response writer
    _, err = w.Write(responseJSON)
    if err != nil {
        http.Error(w, "Failed to write response", http.StatusInternalServerError)
        return
    }
}

func (a *App) getUnsharedUsersForNoteHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    noteIDStr, ok := vars["noteID"]
    if !ok {
        http.Error(w, "Missing noteID in URL", http.StatusBadRequest)
        return
    }

    noteID, err := strconv.Atoi(noteIDStr)
    if err != nil {
        http.Error(w, "Invalid noteID", http.StatusBadRequest)
        return
    }

	// Get the current username from the session
    sess := session.Get(r)
    username := "[guest]"

    if sess != nil {
        username = sess.CAttr("username").(string)
    }

    // Fetch the unshared users for the given noteID
    unsharedUsers, err := a.getUnsharedUsersForNote(noteID, username)
    if err != nil {
        http.Error(w, "Failed to fetch unshared users: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Marshal the unsharedUsers slice into JSON
    responseJSON, err := json.Marshal(unsharedUsers)
    if err != nil {
        http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
        return
    }

    // Set the Content-Type header to indicate JSON response
    w.Header().Set("Content-Type", "application/json")

    // Write the JSON response to the HTTP response writer
    _, err = w.Write(responseJSON)
    if err != nil {
        http.Error(w, "Failed to write response", http.StatusInternalServerError)
        return
    }
}

func (a *App) searchNotesHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}

	sess := session.Get(r)
    username := "[guest]"

    if sess != nil {
        username = sess.CAttr("username").(string)
    }

	// Get the list of all users
    allUsers, err := a.getAllUsers(username)
    if err != nil {
        // Handle the error appropriately (e.g., log it or show an error page)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

	// Retrieve all shared notes with privileges
    sharedNotes, err := a.retrieveSharedNotesWithPrivileges(username)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    searchQuery := r.FormValue("searchQuery")

	const MaxSearchLength = 50

	// Validate the length of title and description
    if len(searchQuery) > MaxSearchLength  {
        
        
        http.SetCookie(w, &http.Cookie{
            Name:  "errorMessage",
            Value: "Search Error: Search query exceeds 50 characters.", // Set your error message
            Path:  "/list", // Set the path as needed
        })
        http.Redirect(w, r, "/list", http.StatusSeeOther)
        return
    }

    // Query your database using FTS to search for notes based on searchQuery
    results, err := a.searchNotesInDatabase(searchQuery, username)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Retrieve shared users for each note in the search results
    for i, note := range results {
        sharedUsers, err := a.getSharedUsersForNote(note.ID)
        if err != nil {
            http.Error(w, "Failed to fetch shared users: "+err.Error(), http.StatusInternalServerError)
            return
        }
        results[i].SharedUsers = sharedUsers
    }

    // Pass the search results with shared users to the template
    data := struct {
		Username string
        SearchResults []Note
		SearchQuery string
		AllUsers      []User
		SharedNotes   []Note
        
		
    }{
		Username: username,
        SearchResults: results,
		SearchQuery: searchQuery,
		AllUsers:      allUsers,
		SharedNotes:   sharedNotes,
        
    }

	var funcMap = template.FuncMap{
		"formatTime": func(layout, value string) (string, error) {
			t, err := time.Parse(layout, value)
			if err != nil {
				return "", err
			}
			return t.Format("3:04 PM"), nil
		},
		"formatDate": func(layout, value string) (string, error) {
			t, err := time.Parse(layout, value)
			if err != nil {
				return "", err
			}
			return t.Format("02/01/2006"), nil
		},
	}

    t, err := template.New("search_results.html").Funcs(funcMap).ParseFiles("tmpl/search_results.html")

	var buf bytes.Buffer
    err = t.Execute(&buf, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=UTF-8")
    buf.WriteTo(w)

}

func (a *App) createHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
    }

    sess := session.Get(r)
    if sess == nil {
        // Handle the case when session is not found or authenticated
        http.Error(w, "Unauthorized", http.StatusSeeOther)
        return
    }

    username := sess.CAttr("username").(string)

    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    const MaxNoteLength = 256


    var note Note
    note.Title = r.FormValue("Title")
    note.NoteType = r.FormValue("NoteType")
    note.Description = r.FormValue("Description")
    note.Owner = username
    note.TaskCompletionDate.String = r.FormValue("TaskCompletionDate")
    note.TaskCompletionTime.String = r.FormValue("TaskCompletionTime")
    note.NoteStatus.String = r.FormValue("NoteStatus")
    note.NoteDelegation.String = r.FormValue("NoteDelegation")

	

    // Validate the length of title and description
    if len(note.Title) > MaxNoteLength || len(note.Description) > MaxNoteLength {
        
        
        http.SetCookie(w, &http.Cookie{
            Name:  "errorMessage",
            Value: "Create Error: Note title or description exceeds 256 characters.", // Set your error message
            Path:  "/list", // Set the path as needed
        })
        http.Redirect(w, r, "/list", http.StatusSeeOther)
        return
    }

    // Insert the new note into the database
    err := a.insertNoteIntoDatabase(note)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) updateHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}

    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

	const MaxNoteLength = 256

    var note Note
    note.ID, _ = strconv.Atoi(r.FormValue("Id")) // Given ID
    note.Title = r.FormValue("Title")
    note.NoteType = r.FormValue("NoteType")
    note.Description = r.FormValue("Description")
    note.TaskCompletionTime.String = r.FormValue("TaskCompletionTime")
    note.TaskCompletionDate.String = r.FormValue("TaskCompletionDate")
    note.NoteStatus.String = r.FormValue("NoteStatus")
    note.NoteDelegation.String = r.FormValue("NoteDelegation")


	// Validate the length of title and description
    if len(note.Title) > MaxNoteLength || len(note.Description) > MaxNoteLength {
        
        
        http.SetCookie(w, &http.Cookie{
            Name:  "errorMessage",
            Value: "Update Error: Note title or description exceeds 256 characters.", // Set your error message
            Path:  "/list", // Set the path as needed
        })
        http.Redirect(w, r, "/list", http.StatusSeeOther)
        return
    }

    // Update the note in the database
    err := a.updateNoteInDatabase(note)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    // Redirect back to the list page or another appropriate page
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}

    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    noteID, _ := strconv.Atoi(r.FormValue("Id"))

    // Delete the note from the database
    err := a.deleteNoteFromDatabase(noteID)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    http.Redirect(w, r, "/list", http.StatusSeeOther)
}


func (a *App) shareHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Extract the shared user's username and privileges from the form
    sharedUsername := r.FormValue("SharedUsername")
    privileges := r.FormValue("Privileges")
    noteID, _ := strconv.Atoi(r.FormValue("Id"))

    // Share the note with the user in the database
    err := a.shareNoteWithUser(noteID, sharedUsername, privileges)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Provide feedback to the user (e.g., "Note shared successfully")

    // Redirect to an appropriate page
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}


func (a *App) removeSharedNoteHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse the note ID from the request
    noteID := r.FormValue("noteID")
	username := r.FormValue("username")

    // Implement the logic to remove the shared note from the user_shares table
    err := a.removeSharedNoteFromUser(username, noteID)
    if err != nil {
        // Handle the error appropriately (e.g., log it or show an error page)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Redirect the user to a success page or back to the list of shared notes
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) removeDelegationHandler(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
    }

	vars := mux.Vars(r)
    noteIDStr, ok := vars["noteID"]
    if !ok {
        http.Error(w, "Missing noteID in URL", http.StatusBadRequest)
        return
    }

    noteID, err := strconv.Atoi(noteIDStr)
    if err != nil {
        http.Error(w, "Invalid noteID", http.StatusBadRequest)
        return
    }

    // Call the database function to remove delegation
    if err := a.RemoveDelegation(noteID); // Replace with your actual DB function
    err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, map[string]string{"message": "Delegation removed successfully"})
}


func (a *App) updatePrivilegesHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}
    // Parse the POST data to retrieve the selected username and updated privileges
    r.ParseForm()
    selectedUsername := r.Form.Get("username")
    updatedPrivileges := r.Form.Get("privileges")
    noteID := r.Form.Get("noteID")

    // Perform the database update to change privileges for the selected user and noteID
    err := a.updateUserPrivileges(selectedUsername, updatedPrivileges, noteID)
    if err != nil {
        http.Error(w, "Failed to update privileges: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Redirect back to the list page after successfully updating privileges
	// Somehow add user feedback
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) findInNoteHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}
    vars := mux.Vars(r)
    noteIDStr, ok := vars["noteID"]
    if !ok {
        http.Error(w, "Missing noteID in URL", http.StatusBadRequest)
        return
    }

    noteID, err := strconv.Atoi(noteIDStr)
    if err != nil {
        http.Error(w, "Invalid noteID", http.StatusBadRequest)
        return
    }

	searchPattern := r.FormValue("searchInput")


    // Implement your search logic to find text in the note with the given noteID.
    // This could involve searching in your data store, such as a database, for the specified text pattern.
    searchResults, err := a.findTextInNote(noteID, searchPattern)
    if err != nil {
        http.Error(w, "Failed to find text in note: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Marshal the searchResults into JSON
    responseJSON, err := json.Marshal(searchResults)
    if err != nil {
        http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
        return
    }

    // Set the Content-Type header to indicate JSON response
    w.Header().Set("Content-Type", "application/json")

    // Write the JSON response to the HTTP response writer
    _, err = w.Write(responseJSON)
    if err != nil {
        http.Error(w, "Failed to write response", http.StatusInternalServerError)
        return
    }
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DISABLE_AUTH") != "1" {
        // Perform authentication checks only if the environment variable is not set
        a.isAuthenticated(w, r)
	}
	http.Redirect(w, r, "/list", http.StatusSeeOther)
}

