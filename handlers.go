package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strconv"

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
        a.username = username
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

    var funcMap = template.FuncMap{
        "addOne": func(i int) int {
            return i + 1
        },
		"filterSharedUsers": func(sharedUsers []User) []User {
			
			var filteredUsers []User
			for _, user := range allUsers {
				found := false
				for _, sharedUser := range sharedUsers {
					if user.Username == sharedUser.Username {
						found = true
						break
					}
				}
				if !found {
					filteredUsers = append(filteredUsers, user)
				}
			}
			return filteredUsers
		},
		
		
    }

    // Pass the shared notes with privileges to the template
    data := struct {
        Username      string
        Notes         []Note
        AllUsers      []User
        SharedNotes   []Note
    }{
        Username:      username,
        Notes:         notes,
        AllUsers:      allUsers,
        SharedNotes:   sharedNotes,
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


func (a *App) getSharedUsersForNote(noteID int) ([]UserShare, error) {
    // Initialize a slice to store the shared users
    var sharedUsers []UserShare

    // Perform a database query to fetch shared users and their privileges for the given noteID
    rows, err := a.db.Query("SELECT username, privileges FROM user_shares WHERE note_id = $1", noteID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var username sql.NullString
        var privileges sql.NullString
        // Populate the username and privileges from the database result
        if err := rows.Scan(&username, &privileges); err != nil {
            return nil, err
        }

        // Create a UserShare struct with the username and privileges
        user := UserShare{
            Username: username,
            Privileges: privileges,
            // Add other user fields if needed
        }

        // Append the user to the sharedUsers slice
        sharedUsers = append(sharedUsers, user)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return sharedUsers, nil
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

func (a *App) getUnsharedUsersForNote(noteID int, username string) ([]User, error) {
    // Initialize a slice to store unshared users
    var unsharedUsers []User

    

    // Perform a database query to fetch users who have not been shared with the note
    // and are not the current user
    rows, err := a.db.Query("SELECT username FROM users WHERE username NOT IN (SELECT username FROM user_shares WHERE note_id = $1) AND username != $2", noteID, username)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var username string
        // Populate the username from the database result
        if err := rows.Scan(&username); err != nil {
            return nil, err
        }

        // Create a User struct with the username
        user := User{
            Username: username,
            // Add other user fields if needed
        }

        // Append the user to the unsharedUsers slice
        unsharedUsers = append(unsharedUsers, user)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return unsharedUsers, nil
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





// Add the searchNotesHandler function
func (a *App) searchNotesHandler(w http.ResponseWriter, r *http.Request) {
    searchQuery := r.FormValue("searchQuery")

    // Query your database using FTS to search for notes based on searchQuery
    results, err := a.searchNotesInDatabase(searchQuery, a.username)
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
    searchData := struct {
        SearchResults []Note
    }{
        SearchResults: results,
    }

    // Render the search results in your template
    t, err := template.New("search_results.html").ParseFiles("tmpl/search_results.html")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Pass the search results to the template
    if err := t.Execute(w, searchData); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
}


func (a *App) searchNotesInDatabase(searchQuery string, username string) ([]Note, error) {
    query := `
        SELECT notes.id, notes.title, notes.notetype, notes.description, notes.notecreated,
               notes.taskcompletiondate, notes.taskcompletiontime, notes.notestatus, notes.notedelegation,
               user_shares.username AS shared_username
        FROM notes
        LEFT JOIN user_shares ON notes.id = user_shares.note_id
        WHERE (notes.fts_text @@ plainto_tsquery('english', $1) AND notes.owner = $2)
           OR (user_shares.username = $1)
    `
    rows, err := a.db.Query(query, searchQuery, username)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var notes []Note
    noteMap := make(map[int]*Note)

    for rows.Next() {
        var note Note
        var sharedUsername sql.NullString

        if err := rows.Scan(&note.ID, &note.Title, &note.NoteType, &note.Description, &note.NoteCreated,
            &note.TaskCompletionDate.String, &note.TaskCompletionTime.String, &note.NoteStatus.String, &note.NoteDelegation.String, &sharedUsername); err != nil {
            return nil, err
        }

        // Check if the note is already in the notes slice
        existingNote, exists := noteMap[note.ID]
        if !exists {
            // If it doesn't exist, add it to the map and slice
            notes = append(notes, note)
            noteMap[note.ID] = &notes[len(notes)-1]
            existingNote = noteMap[note.ID]
        }

        // If sharedUsername is not null, add it to the note's shared users
        if sharedUsername.Valid {
            sharedUser := UserShare{
                Username: sharedUsername,
            }
            existingNote.SharedUsers = append(existingNote.SharedUsers, sharedUser)
        }
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return notes, nil
}




func (a *App) retrieveNotes(username string) ([]Note, error) {
    // Query to fetch notes and shared users' data
    query := `
        SELECT
            n.*, u.username, us.privileges
        FROM
            notes n
        LEFT JOIN
            user_shares us ON n.id = us.note_id
        LEFT JOIN
            users u ON us.username = u.username
        WHERE
            n.owner = $1
        ORDER BY
            n.id
    `

    rows, err := a.db.Query(query, username)
    if err != nil {
        return nil, err
    }

    defer rows.Close()

    notesMap := make(map[int]Note)
    for rows.Next() {
        var note Note
        var sharedUser UserShare

        err := rows.Scan(
            &note.ID, &note.Title, &note.NoteType, &note.Description, &note.NoteCreated,
            &note.TaskCompletionTime, &note.TaskCompletionDate, &note.NoteStatus,
            &note.NoteDelegation, &note.Owner, &note.FTSText,
            &sharedUser.Username, &sharedUser.Privileges,
        )
        if err != nil {
            return nil, err
        }

        // Append shared users to the note
        note.SharedUsers = append(note.SharedUsers, sharedUser)

        // Store the note in a map using its ID as the key
        notesMap[note.ID] = note
    }

    // Convert the map of notes to a slice
    notes := make([]Note, 0, len(notesMap))
    for _, note := range notesMap {
        notes = append(notes, note)
    }

    return notes, nil
}




func (a *App) retrieveSharedNotesWithPrivileges(username string) ([]Note, error) {
    rows, err := a.db.Query(`
        SELECT n.*, us.privileges
        FROM notes n
        INNER JOIN user_shares us ON n.id = us.note_id
        WHERE us.username = $1
        ORDER BY n.id
    `, username)
    if err != nil {
        return nil, err
    }

    var sharedNotes []Note
    for rows.Next() {
        var sharedNote Note
		
        
        err := rows.Scan(
            &sharedNote.ID,
            &sharedNote.Title,
            &sharedNote.NoteType,
            &sharedNote.Description,
            &sharedNote.NoteCreated,
            &sharedNote.TaskCompletionTime,
            &sharedNote.TaskCompletionDate,
            &sharedNote.NoteStatus,
            &sharedNote.NoteDelegation,
            &sharedNote.Owner,
            &sharedNote.FTSText,
            &sharedNote.Privileges, // Retrieve the 'privileges' field
        )
        if err != nil {
            return nil, err
        }
        
        // Set the 'Privileges' field in the Note struct
        

        sharedNotes = append(sharedNotes, sharedNote)
    }

    return sharedNotes, nil
}




func (a *App) getAllUsers(ownerUsername string) ([]User, error) {
    var users []User

    rows, err := a.db.Query("SELECT username FROM users WHERE username != $1", ownerUsername)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var user User
        if err := rows.Scan(&user.Username); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
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
    note.TaskCompletionDate.String = r.FormValue("TaskCompletionDate")
    note.TaskCompletionTime.String = r.FormValue("TaskCompletionTime")
    note.NoteStatus.String = r.FormValue("NoteStatus")
    note.NoteDelegation.String = r.FormValue("NoteDelegation")


	// Save to database
	_, err := a.db.Exec(`
		INSERT INTO notes (title, noteType, description, TaskCompletionDate, TaskCompletionTime, NoteStatus, NoteDelegation, owner, fts_text)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, to_tsvector('english', $1 || ' ' || $2 || ' ' || $3 || ' ' || $4 || ' ' || $5 || ' ' || $6 || ' ' || $7 || ' ' || $8))
	`, note.Title, note.NoteType, note.Description, note.TaskCompletionDate.String, note.TaskCompletionTime.String, note.NoteStatus.String, note.NoteDelegation.String, note.Owner)
	checkInternalServerError(err, w)

	

	http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) updateHandler(w http.ResponseWriter, r *http.Request) {
    a.isAuthenticated(w, r)

    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    var note Note
    note.ID, _ = strconv.Atoi(r.FormValue("Id")) // Given ID
    note.Title = r.FormValue("Title") 
    note.NoteType = r.FormValue("NoteType")
    note.Description = r.FormValue("Description")
	note.TaskCompletionTime.String = r.FormValue("TaskCompletionTime")
    note.TaskCompletionDate.String = r.FormValue("TaskCompletionDate")
    note.NoteStatus.String = r.FormValue("NoteStatus")
    note.NoteDelegation.String = r.FormValue("NoteDelegation")

    // Update the database
    _, err := a.db.Exec(`
        UPDATE notes SET title=$1, noteType=$2, description=$3,
        taskcompletiontime=$4, taskcompletiondate=$5, notestatus=$6, notedelegation=$7
        WHERE id=$8
    `, note.Title, note.NoteType, note.Description, note.TaskCompletionTime.String,
    note.TaskCompletionDate.String, note.NoteStatus.String, note.NoteDelegation.String, note.ID)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    // Redirect back to the list page or another appropriate page
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}


func (a *App) deleteHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	noteID, _ := strconv.Atoi(r.FormValue("Id"))
	// Delete from the database
	_, err := a.db.Exec("DELETE FROM notes WHERE id=$1", noteID)
	checkInternalServerError(err, w)

	http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) shareHandler(w http.ResponseWriter, r *http.Request) {
    a.isAuthenticated(w, r)

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Extract the shared user's username and privileges from the form
    sharedUsername := r.FormValue("SharedUsername")
    privileges := r.FormValue("Privileges")
    noteID := r.FormValue("Id")
	
	

    // Check if the shared user exists in the users table by username
    var sharedUserID string 
    err := a.db.QueryRow("SELECT username FROM users WHERE username = $1", sharedUsername).Scan(&sharedUserID)
    if err != nil {
        // Handle the case where the shared user does not exist
        // You can display an error message or redirect as needed
        http.Error(w, "Invalid shared user", http.StatusBadRequest)
        return
    }

    // Check if the note with the given ID exists
    var noteExists bool
    err = a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM notes WHERE id = $1)", noteID).Scan(&noteExists)
	
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    if !noteExists {
        http.Error(w, "Note does not exist", http.StatusBadRequest)
        return
    }

    // Check if there is already an existing entry for the given note and shared user
    var existingShareID int
    err = a.db.QueryRow("SELECT note_id FROM user_shares WHERE note_id = $1 AND username = $2", noteID, sharedUsername).Scan(&existingShareID)
    if err == nil {
        // An existing entry was found, which means the note is already shared with the user
        // You can handle this case by displaying an error message to the user
        http.Error(w, "Note is already shared with this user", http.StatusBadRequest)
        return
    } else if err != sql.ErrNoRows {
        // Handle any other errors that may have occurred during the query
        checkInternalServerError(err, w)
        return
    }

    // If no existing entry was found, proceed with sharing the note
    _, err = a.db.Exec(`
        INSERT INTO user_shares (note_id, username, privileges)
        VALUES ($1, $2, $3)
    `, noteID, sharedUsername, privileges)
    if err != nil {
        checkInternalServerError(err, w)
        return
    }

    // Provide feedback to the user (e.g., "Note shared successfully")

    // Redirect to an appropriate page
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}

func (a *App) removeSharedNoteHandler(w http.ResponseWriter, r *http.Request) {
    a.isAuthenticated(w, r)

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

func (a *App) removeSharedNoteFromUser(username string, noteID string) error {
    _, err := a.db.Exec(`
        DELETE FROM user_shares
        WHERE username = $1 AND note_id = $2
    `, username, noteID)
    if err != nil {
        return err
    }

    return nil
}


func (a *App) updatePrivilegesHandler(w http.ResponseWriter, r *http.Request) {
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

func (a *App) updateUserPrivileges(selectedUsername, updatedPrivileges, noteID string) error {
    // Prepare the SQL statement to update privileges for the selected user and noteID
    query := "UPDATE user_shares SET privileges = $1 WHERE username = $2 AND note_id = $3"

    _, err := a.db.Exec(query, updatedPrivileges, selectedUsername, noteID)
    if err != nil {
        return err
    }

    return nil
}



/*func (a *App) getFilteredUsers(ownerUsername string, noteID int) ([]User, error) {
	fmt.Print("hi")
    var users []User
	fmt.Printf("%d", noteID)

    // Use a subquery to select users who are not in the user_shares table for the given noteID
    query := `
        SELECT username FROM users
        WHERE username != $1
        AND username NOT IN (
            SELECT username FROM user_shares WHERE note_id = $2
        )
    `

    rows, err := a.db.Query(query, ownerUsername, noteID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var user User
        if err := rows.Scan(&user.Username); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}




func (a *App) shareNoteHandler(w http.ResponseWriter, r *http.Request) {
    noteID := r.FormValue("Id") // Get the noteID from the form data
	noteIDInt, err := strconv.Atoi(noteID)


	sess := session.Get(r)
    username := "[guest]"

    if sess != nil {
        username = sess.CAttr("username").(string)
    }

    // Query your database to get the list of non-shared users for the specified note.
    nonSharedUsers, err := a.getFilteredUsers(username, noteIDInt)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
	fmt.Printf("%v", nonSharedUsers)

    // Redirect the user to a success page or back to the list of shared notes
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}
*/









func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	a.isAuthenticated(w, r)
	http.Redirect(w, r, "/list", http.StatusSeeOther)
}

