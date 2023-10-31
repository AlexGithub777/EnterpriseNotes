package main

import (
	"database/sql"
	"fmt"
	"strings"
)

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
        n.owner = $1 OR n.noteDelegation = $1
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

func (a *App) updateNoteInDatabase(note Note) error {
    _, err := a.db.Exec(`
        UPDATE notes SET title=$1, noteType=$2, description=$3,
        taskcompletiontime=$4, taskcompletiondate=$5, notestatus=$6, notedelegation=$7
        WHERE id=$8
    `, note.Title, note.NoteType, note.Description, note.TaskCompletionTime.String,
        note.TaskCompletionDate.String, note.NoteStatus.String, note.NoteDelegation.String, note.ID)

    if err != nil {
        return err
    }

    // Recalculate the fts_text field
    _, err = a.db.Exec(`
        UPDATE notes SET fts_text = to_tsvector('english', title || ' ' || noteType || ' ' || description || ' ' || taskcompletiontime || ' ' || taskcompletiondate || ' ' || notestatus || ' ' || notedelegation)
        WHERE id = $1
    `, note.ID)
    if err != nil {
        return err
    }

    return nil
}

func (a *App) insertNoteIntoDatabase(note Note) error {
    _, err := a.db.Exec(`
        INSERT INTO notes (title, noteType, description, TaskCompletionDate, TaskCompletionTime, NoteStatus, NoteDelegation, owner, fts_text)
        VALUES($1, $2, $3, $4, $5, $6, $7, $8, to_tsvector('english', $1 || ' ' || $2 || ' ' || $3 || ' ' || $4 || ' ' || $5 || ' ' || $6 || ' ' || $7 || ' ' || $8))
    `, note.Title, note.NoteType, note.Description, note.TaskCompletionDate.String, note.TaskCompletionTime.String, note.NoteStatus.String, note.NoteDelegation.String, note.Owner)

    return err
}



func (a *App) searchNotesInDatabase(searchQuery string, username string) ([]Note, error) {
    query := `
        SELECT notes.id, notes.title, notes.noteType, notes.description, notes.noteCreated,
               notes.taskCompletionDate, notes.taskCompletionTime, notes.noteStatus, notes.noteDelegation,
               user_shares.username AS shared_username
        FROM notes
        LEFT JOIN user_shares ON notes.id = user_shares.note_id
        WHERE (notes.fts_text @@ plainto_tsquery('english', $1) AND notes.owner = $2)
           OR (user_shares.username ILIKE $1)
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

// RemoveDelegation removes delegation for a note with the given noteID.
func (a *App) RemoveDelegation(noteID int) error {
    // Use a SQL statement to set the noteDelegation to NULL for the specified noteID
    query := "UPDATE notes SET noteDelegation = NULL WHERE id = $1"
    _, err := a.db.Exec(query, noteID)

    if err != nil {
        return fmt.Errorf("Failed to remove delegation: %v", err)
    }

    return nil
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

func (a *App) deleteNoteFromDatabase(noteID int) error {
    _, err := a.db.Exec("DELETE FROM notes WHERE id=$1", noteID)
    return err
}

func (a *App) shareNoteWithUser(noteID int, sharedUsername string, privileges string) error {
    // Check if the shared user exists in the users table by username
    var sharedUserID string
    err := a.db.QueryRow("SELECT username FROM users WHERE username = $1", sharedUsername).Scan(&sharedUserID)
    if err != nil {
        // Handle the case where the shared user does not exist
        return err
    }

    // Check if the note with the given ID exists
    var noteExists bool
    err = a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM notes WHERE id = $1)", noteID).Scan(&noteExists)
    if err != nil {
        return err
    }

    if !noteExists {
        return fmt.Errorf("Note does not exist")
    }

    // Check if there is already an existing entry for the given note and shared user
    var existingShareID int
    err = a.db.QueryRow("SELECT note_id FROM user_shares WHERE note_id = $1 AND username = $2", noteID, sharedUsername).Scan(&existingShareID)
    if err == nil {
        return fmt.Errorf("Note is already shared with this user")
    } else if err != sql.ErrNoRows {
        return err
    }

    // If no existing entry was found, proceed with sharing the note
    _, err = a.db.Exec(`
        INSERT INTO user_shares (note_id, username, privileges)
        VALUES ($1, $2, $3)
    `, noteID, sharedUsername, privileges)
    return err
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

func (a *App) updateUserPrivileges(selectedUsername, updatedPrivileges, noteID string) error {
    // Prepare the SQL statement to update privileges for the selected user and noteID
    query := "UPDATE user_shares SET privileges = $1 WHERE username = $2 AND note_id = $3"

    _, err := a.db.Exec(query, updatedPrivileges, selectedUsername, noteID)
    if err != nil {
        return err
    }

    return nil
}

type SearchResult struct {
    Count       int
    Description string
}

func (a *App) findTextInNote(noteID int, searchPattern string) ([]SearchResult, error) {
    // Fetch the note with the given ID to access the title and description
    note, err := a.getNoteByID(noteID)
    if err != nil {
        return nil, err
    }

    // Count occurrences in the title and description
    titleOccurrences := countOccurrences(note.Title, searchPattern)
    descriptionOccurrences := countOccurrences(note.Description, searchPattern)

    results := []SearchResult{}

    if titleOccurrences > 0 {
        results = append(results, SearchResult{
            Count:       titleOccurrences,
            Description: "Title",
        })
    }

    if descriptionOccurrences > 0 {
        results = append(results, SearchResult{
            Count:       descriptionOccurrences,
            Description: "Description",
        })
    }

    return results, nil
}



// countOccurrences counts the number of occurrences of a searchPattern in a text.
func countOccurrences(text, searchPattern string) int {
    // Implement your logic to count occurrences. You can use strings.Count or regular expressions.
    // Here's a simple example using strings.Count:
    count := strings.Count(strings.ToLower(text), strings.ToLower(searchPattern))
    return count
}


func (a *App) getNoteByID(noteID int) (*Note, error) {
    // Query the database to retrieve the note.
    var note Note
    err := a.db.QueryRow("SELECT id, title, description, noteType, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner FROM notes WHERE id = $1", noteID).Scan(&note.ID, &note.Title, &note.Description, &note.NoteType, &note.TaskCompletionTime, &note.TaskCompletionDate, &note.NoteStatus, &note.NoteDelegation, &note.Owner)
    if err != nil {
        return nil, err
    }

    // Return the retrieved note.
    return &note, nil
}

