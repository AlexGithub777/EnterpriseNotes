package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func (a *App) retrieveNotes(username string) ([]Note, error) {
	// Prepare the SQL statement for fetching notes and shared users' data
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

	stmt, err := a.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
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
	// Prepare the SQL statement for fetching shared notes with privileges
	query := `
		SELECT n.*, us.privileges
		FROM notes n
		INNER JOIN user_shares us ON n.id = us.note_id
		WHERE us.username = $1
		ORDER BY n.id
	`

	stmt, err := a.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
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

		sharedNotes = append(sharedNotes, sharedNote)
	}

	return sharedNotes, nil
}

func (a *App) getAllUsers(ownerUsername string) ([]User, error) {
	// Prepare the SQL statement for fetching all users except the owner
	query := "SELECT username FROM users WHERE username != $1"

	stmt, err := a.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(ownerUsername)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
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

    // Prepare the SQL statement for fetching shared users and their privileges for the given noteID
    query := "SELECT username, privileges FROM user_shares WHERE note_id = $1"

    stmt, err := a.db.Prepare(query)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    rows, err := stmt.Query(noteID)
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
            Username:   username,
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
	// Prepare the SQL statement for updating note fields
	updateQuery := `
        UPDATE notes
        SET title = $1, noteType = $2, description = $3,
        taskcompletiontime = $4, taskcompletiondate = $5, notestatus = $6, notedelegation = $7
        WHERE id = $8
    `

	updateStmt, err := a.db.Prepare(updateQuery)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	_, err = updateStmt.Exec(
		note.Title,
		note.NoteType,
		note.Description,
		note.TaskCompletionTime.String,
		note.TaskCompletionDate.String,
		note.NoteStatus.String,
		note.NoteDelegation.String,
		note.ID,
	)
	if err != nil {
		return err
	}

	// Prepare the SQL statement for recalculating the fts_text field
	recalculateQuery := `
        UPDATE notes
        SET fts_text = to_tsvector('english', title || ' ' || noteType || ' ' || description || ' ' || taskcompletiontime || ' ' || taskcompletiondate || ' ' || notestatus || ' ' || notedelegation)
        WHERE id = $1
    `

	recalculateStmt, err := a.db.Prepare(recalculateQuery)
	if err != nil {
		return err
	}
	defer recalculateStmt.Close()

	_, err = recalculateStmt.Exec(note.ID)
	if err != nil {
		return err
	}

	return nil
}


func (a *App) insertNoteIntoDatabase(note Note) error {
	// Prepare the SQL statement for inserting a new note
	insertQuery := `
        INSERT INTO notes (title, noteType, description, TaskCompletionDate, TaskCompletionTime, NoteStatus, NoteDelegation, owner, fts_text)
		VALUES (
			$1::text, $2::text, $3::text, $4::text, $5::text, $6::text, $7::text, $8::text,
			to_tsvector('english', $1::text || ' ' || $2::text || ' ' || $3::text || ' ' || $4::text || ' ' || $5::text || ' ' || $6::text || ' ' || $7::text || ' ' || $8::text)
		)
		`

	insertStmt, err := a.db.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(
		note.Title,
		note.NoteType,
		note.Description,
		note.TaskCompletionDate.String,
		note.TaskCompletionTime.String,
		note.NoteStatus.String,
		note.NoteDelegation.String,
		note.Owner,
	)
	if err != nil {
		return err
	}

	return nil
}


func (a *App) searchNotesInDatabase(searchQuery string, username string) ([]Note, error) {
    // Attempted to validate search query, would alter search results so didn't keep
	/*if !isValidSearchQuery(searchQuery) {
        fmt.Printf("Invalid search query")
        return []Note{}, nil
    }*/
	
	// Prepare the SQL statement for searching notes
    query := `
        SELECT notes.id, notes.title, notes.noteType, notes.description, notes.noteCreated,
               notes.taskCompletionDate, notes.taskCompletionTime, notes.noteStatus, notes.noteDelegation, notes.owner,
               user_shares.username AS shared_username
        FROM notes
        LEFT JOIN user_shares ON notes.id = user_shares.note_id
        WHERE (notes.fts_text @@ plainto_tsquery('english', $1) AND (notes.owner = $2 OR notes.noteDelegation = $2))
        OR (user_shares.username ILIKE $1)
    `

    stmt, err := a.db.Prepare(query)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    rows, err := stmt.Query(searchQuery, username)
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
            &note.TaskCompletionDate.String, &note.TaskCompletionTime.String, &note.NoteStatus.String, &note.NoteDelegation.String, &note.Owner, &sharedUsername); err != nil {
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

// Attempted to validate search query, would alter search results so didn't keep
/*
func isValidSearchQuery(textPattern string) bool {
    // Define regular expressions for the valid patterns
    sentencePattern := `^.*\bprefix\b.*\bsuffix\b.*$`
    phoneNumberPattern := `^\d{3}-\d{7}$`
    partialEmailPattern := `^[a-zA-Z0-9._%+-]+@.*$`
    keywordsPattern := `(?i)\b(meeting|minutes|agenda|action|attendees|apologies)\b`
    allCapsWordPattern := `^[A-Z]{3,}.*$`

    // Check if the searchQuery matches any of the valid patterns
    return regexp.MustCompile(sentencePattern).MatchString(textPattern) ||
        regexp.MustCompile(phoneNumberPattern).MatchString(textPattern) ||
        regexp.MustCompile(partialEmailPattern).MatchString(textPattern) ||
        regexp.MustCompile(keywordsPattern).MatchString(textPattern) ||
        regexp.MustCompile(allCapsWordPattern).MatchString(textPattern)
}
*/

func (a *App) RemoveDelegation(noteID int) error {
	// Prepare the SQL statement for removing delegation
	query := "UPDATE notes SET noteDelegation = NULL, noteStatus = NULL WHERE id = $1"

	stmt, err := a.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(noteID)
	if err != nil {
		return fmt.Errorf("Failed to remove delegation: %v", err)
	}

	return nil
}


func (a *App) getUnsharedUsersForNote(noteID int, username string) ([]User, error) {
	// Initialize a slice to store unshared users
	var unsharedUsers []User

	// Prepare the SQL statement for fetching unshared users
	query := `
        SELECT username FROM users
        WHERE username NOT IN (SELECT username FROM user_shares WHERE note_id = $1) AND username != $2
    `

	stmt, err := a.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(noteID, username)
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
    // Prepare the SQL statement for deleting a note by ID
    query := "DELETE FROM notes WHERE id = $1"
    
    stmt, err := a.db.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(noteID)
    if err != nil {
        return err
    }

    return nil
}

func (a *App) shareNoteWithUser(noteID int, sharedUsername string, privileges string) error {
    // Prepare the SQL statement for checking if the shared user exists
    checkUserQuery := "SELECT username FROM users WHERE username = $1"

    // Prepare the SQL statement for checking if the note exists
    checkNoteQuery := "SELECT EXISTS(SELECT 1 FROM notes WHERE id = $1)"

    // Prepare the SQL statement for checking if there is an existing entry for the note and shared user
    checkExistingShareQuery := "SELECT note_id FROM user_shares WHERE note_id = $1 AND username = $2"

    // Prepare the SQL statement for inserting the user share
    insertUserShareQuery := `
        INSERT INTO user_shares (note_id, username, privileges)
        VALUES ($1, $2, $3)
    `

    // Check if the shared user exists
    var sharedUserID string
    err := a.db.QueryRow(checkUserQuery, sharedUsername).Scan(&sharedUserID)
    if err != nil {
        // Handle the case where the shared user does not exist
        return err
    }

    // Check if the note with the given ID exists
    var noteExists bool
    err = a.db.QueryRow(checkNoteQuery, noteID).Scan(&noteExists)
    if err != nil {
        return err
    }

    if !noteExists {
        return fmt.Errorf("Note does not exist")
    }

    // Check if there is already an existing entry for the given note and shared user
    var existingShareID int
    err = a.db.QueryRow(checkExistingShareQuery, noteID, sharedUsername).Scan(&existingShareID)
    if err == nil {
        return fmt.Errorf("Note is already shared with this user")
    } else if err != sql.ErrNoRows {
        return err
    }

    // If no existing entry was found, proceed with sharing the note
    _, err = a.db.Exec(insertUserShareQuery, noteID, sharedUsername, privileges)
    return err
}

func (a *App) removeSharedNoteFromUser(username string, noteID string) error {
    // Prepare the SQL statement for removing the shared note from a user
    query := "DELETE FROM user_shares WHERE username = $1 AND note_id = $2"

    stmt, err := a.db.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(username, noteID)
    if err != nil {
        return err
    }

    return nil
}

func (a *App) updateUserPrivileges(selectedUsername, updatedPrivileges, noteID string) error {
    // Prepare the SQL statement for updating user privileges
    query := "UPDATE user_shares SET privileges = $1 WHERE username = $2 AND note_id = $3"

    stmt, err := a.db.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(updatedPrivileges, selectedUsername, noteID)
    if err != nil {
        return err
    }

    return nil
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
    count := strings.Count(strings.ToLower(text), strings.ToLower(searchPattern))
    return count
}


func (a *App) getNoteByID(noteID int) (*Note, error) {
    query := "SELECT id, title, description, noteType, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner FROM notes WHERE id = $1"
    row := a.db.QueryRow(query, noteID)

    var note Note
    err := row.Scan(&note.ID, &note.Title, &note.Description, &note.NoteType, &note.TaskCompletionTime, &note.TaskCompletionDate, &note.NoteStatus, &note.NoteDelegation, &note.Owner)
    if err != nil {
        return nil, err
    }

    return &note, nil
}


