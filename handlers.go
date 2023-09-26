package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/icza/session"
)

type noteData struct {
	Username string
	Notes    []Note
}


func (a *App) listHandler(w http.ResponseWriter, r *http.Request) {
    a.isAuthenticated(w, r)

    sess := session.Get(r)
    username := "[guest]"
	

    if sess != nil {
        username = sess.CAttr("username").(string)
		fmt.Printf(username)
		
    }

    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Retrieve all notes
    notes, err := a.retrieveNotes(username)
    if err != nil {
        a.checkInternalServerError(err, w)
        return
    }

    var funcMap = template.FuncMap{
        "addOne": func(i int) int {
            return i + 1
        },
    }

    data := struct {
        Username string
        Notes    []Note
    }{
        Username: username,
        Notes:    notes,
    }

    t, err := template.New("list.html").Funcs(funcMap).ParseFiles("tmpl/list.html")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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


func (a *App) retrieveNotes(username string) ([]Note, error) {
	
	fmt.Printf("The username is %s", username)
    rows, err := a.db.Query("SELECT * FROM notes WHERE owner = $1 ORDER BY id", username)
    if err != nil {
        return nil, err
    }

    var notes []Note
    for rows.Next() {
        var note Note
        err := rows.Scan(
            &note.ID,
            &note.Title,
            &note.NoteType,
            &note.Description,
            &note.NoteCreated,
            &note.TaskCompletionTime,
            &note.TaskCompletionDate,
            &note.NoteStatus,
            &note.NoteDelegation,
            &note.Owner,
            &note.FTSText,
        )
        if err != nil {
            return nil, err
        }
        notes = append(notes, note)
    }

    return notes, nil
}



func (a *App) createHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)

	sess := session.Get(r)
	username := sess.CAttr("username").(string)

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var note Note
	note.Title = r.FormValue("Title")
	note.NoteType = r.FormValue("NoteType")
	note.Description = r.FormValue("Description")
	note.Owner = username // Set the owner ID to the logged-in user's ID (adjust as needed) !!! set to userID

	// Save to database
	_, err := a.db.Exec(`
		INSERT INTO notes (title, noteType, description, owner)
		VALUES($1, $2, $3, $4)
	`, note.Title, note.NoteType, note.Description, note.Owner)
	a.checkInternalServerError(err, w)

	

	http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) updateHandler(w http.ResponseWriter, r *http.Request) {
    a.isAuthenticated(w, r)

    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var note Note
    note.ID, _ = strconv.Atoi(r.FormValue("ID"))
    note.Title = r.FormValue("Title")
    note.NoteType = r.FormValue("NoteType")
    note.Description = r.FormValue("Description")

    // Update the database
    _, err := a.db.Exec(`
        UPDATE notes SET title=$1, noteType=$2, description=$3
        WHERE id=$4
    `, note.Title, note.NoteType, note.Description, note.ID)
    if err != nil {
        a.checkInternalServerError(err, w)
        return
    }

    // ... Render the template with updated notes data

    http.Redirect(w, r, "/list", http.StatusSeeOther)
}


func (a *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	noteID, _ := strconv.Atoi(r.FormValue("ID"))

	// Delete from the database
	_, err := a.db.Exec("DELETE FROM notes WHERE id=$1", noteID)
	a.checkInternalServerError(err, w)

	http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	http.Redirect(w, r, "/list", http.StatusSeeOther)
}