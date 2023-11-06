package main

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	// Import the package where your `App` type is defined
	// Import the package where your `Note` and `UserShare` types are defined
)
func TestRetrieveNotes_Success(t *testing.T) {
    // Create a new SQL mock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error occurred while opening a stub database connection: %v", err)
    }
    defer db.Close()

    // Create the App instance and set the database connection to the mock
    a := App{db: db}

    // Define the expected timestamp values
    
    noteCreatedTime := time.Date(2023, 11, 1, 15, 6, 20, 935951100, time.UTC)
    taskCompletionTime := sql.NullString{String: "12:00:00", Valid: true}
    taskCompletionDate := sql.NullString{String: "2023-11-01", Valid: true}

    // Define the expected rows to be returned by the mock
    rows := sqlmock.NewRows([]string{
        "id", "title", "noteType", "description", "noteCreated",
        "taskCompletionTime", "taskCompletionDate", "noteStatus", "noteDelegation", "owner",
        "FTSText", "username", "privileges",
    }).AddRow(
        1, "Test Note", "Type1", "Test Description", noteCreatedTime,
        taskCompletionTime,
        taskCompletionDate,
        sql.NullString{String: "Status1", Valid: true},
        sql.NullString{String: "Delegation1", Valid: true},
        "user1",
        sql.NullString{String: "Test FTSText", Valid: true},
        sql.NullString{String: "shared_user1", Valid: true},
        sql.NullString{String: "editor", Valid: true},
    ).AddRow(
        2, "Test Note 2", "Type2", "Test Description 2", noteCreatedTime,
        sql.NullString{String: "13:00:00", Valid: true},
        sql.NullString{String: "2023-11-02", Valid: true},
        sql.NullString{String: "Status2", Valid: true},
        sql.NullString{String: "Delegation2", Valid: true},
        "user2",
        sql.NullString{String: "Test FTSText 2", Valid: true},
        sql.NullString{String: "shared_user2", Valid: true},
        sql.NullString{String: "viewer", Valid: true},
    )

    // Expect the query with a specific username
    mock.ExpectPrepare("SELECT .* FROM notes n .*").ExpectQuery().
        WithArgs("user1").
        WillReturnRows(rows)

    // Call the function to retrieve notes
    notes, err := a.retrieveNotes("user1")
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Define the expected result
    expectedNotes := []Note{
        {
            ID:               1,
            Title:            "Test Note",
            NoteType:         "Type1",
            Description:      "Test Description",
            NoteCreated:      noteCreatedTime,
            TaskCompletionTime: taskCompletionTime,
            TaskCompletionDate: taskCompletionDate,
            NoteStatus:        sql.NullString{String: "Status1", Valid: true},
            NoteDelegation:    sql.NullString{String: "Delegation1", Valid: true},
            Owner:            "user1",
            FTSText:          sql.NullString{String: "Test FTSText", Valid: true},
            SharedUsers: []UserShare{
                {Username: sql.NullString{String: "shared_user1", Valid: true}, Privileges: sql.NullString{String: "editor", Valid: true}},
            },
        },
        {
            ID:               2,
            Title:            "Test Note 2",
            NoteType:         "Type2",
            Description:      "Test Description 2",
            NoteCreated:      noteCreatedTime,
            TaskCompletionTime: sql.NullString{String: "13:00:00", Valid: true},
            TaskCompletionDate: sql.NullString{String: "2023-11-02", Valid: true},
            NoteStatus:        sql.NullString{String: "Status2", Valid: true},
            NoteDelegation:    sql.NullString{String: "Delegation2", Valid: true},
            Owner:            "user2",
            FTSText:          sql.NullString{String: "Test FTSText 2", Valid: true},
            SharedUsers: []UserShare{
                {Username: sql.NullString{String: "shared_user2", Valid: true}, Privileges: sql.NullString{String: "viewer", Valid: true}},
            },
        },
    }

    // Check if the expected result matches the actual result
    if !reflect.DeepEqual(notes, expectedNotes) {
        t.Errorf("Expected notes to be %v, but got %v", expectedNotes, notes)
    }

    // Ensure all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %s", err)
    }
}

func TestRetrieveSharedNotesWithPrivileges_Success(t *testing.T) {
    // Create a new SQL mock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error occurred while opening a stub database connection: %v", err)
    }
    defer db.Close()

    // Create the App instance and set the database connection to the mock
    a := App{db: db}

    // Define the expected timestamp values
    noteCreatedTime := time.Date(2023, 11, 1, 15, 6, 20, 935951100, time.UTC)
    taskCompletionTime := sql.NullString{String: "12:00:00", Valid: true}
    taskCompletionDate := sql.NullString{String: "2023-11-01", Valid: true}

    // Define the expected rows to be returned by the mock
    rows := sqlmock.NewRows([]string{
        "id", "title", "noteType", "description", "noteCreated",
        "taskCompletionTime", "taskCompletionDate", "noteStatus", "noteDelegation", "owner",
        "FTSText", "privileges",
    }).AddRow(
        1, "Test Note", "Type1", "Test Description", noteCreatedTime,
        taskCompletionTime,
        taskCompletionDate,
        sql.NullString{String: "Status1", Valid: true},
        sql.NullString{String: "Delegation1", Valid: true},
        "user1",
        sql.NullString{String: "Test FTSText", Valid: true},
        "editor", // Privileges is a string
    ).AddRow(
        2, "Test Note 2", "Type2", "Test Description 2", noteCreatedTime,
        sql.NullString{String: "13:00:00", Valid: true},
        sql.NullString{String: "2023-11-02", Valid: true},
        sql.NullString{String: "Status2", Valid: true},
        sql.NullString{String: "Delegation2", Valid: true},
        "user2",
        sql.NullString{String: "Test FTSText 2", Valid: true},
        "viewer", // Privileges is a string
    )

    // Expect the query with a specific username
    mock.ExpectPrepare("SELECT n.*, us.privileges FROM notes n INNER JOIN user_shares us .*").ExpectQuery().
        WithArgs("user1").
        WillReturnRows(rows)

    // Call the function to retrieve shared notes
    sharedNotes, err := a.retrieveSharedNotesWithPrivileges("user1")
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Define the expected result
    expectedNotes := []Note{
        {
            ID:               1,
            Title:            "Test Note",
            NoteType:         "Type1",
            Description:      "Test Description",
            NoteCreated:      noteCreatedTime,
            TaskCompletionTime: taskCompletionTime,
            TaskCompletionDate: taskCompletionDate,
            NoteStatus:        sql.NullString{String: "Status1", Valid: true},
            NoteDelegation:    sql.NullString{String: "Delegation1", Valid: true},
            Owner:            "user1",
            FTSText:          sql.NullString{String: "Test FTSText", Valid: true},
            Privileges:       "editor", // Privileges is a string
        },
        {
            ID:               2,
            Title:            "Test Note 2",
            NoteType:         "Type2",
            Description:      "Test Description 2",
            NoteCreated:      noteCreatedTime,
            TaskCompletionTime: sql.NullString{String: "13:00:00", Valid: true},
            TaskCompletionDate: sql.NullString{String: "2023-11-02", Valid: true},
            NoteStatus:        sql.NullString{String: "Status2", Valid: true},
            NoteDelegation:    sql.NullString{String: "Delegation2", Valid: true},
            Owner:            "user2",
            FTSText:          sql.NullString{String: "Test FTSText 2", Valid: true},
            Privileges:       "viewer", // Privileges is a string
        },
    }

    // Check if the expected result matches the actual result
    if !reflect.DeepEqual(sharedNotes, expectedNotes) {
        t.Errorf("Expected sharedNotes to be %v, but got %v", expectedNotes, sharedNotes)
    }

    // Ensure all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %s", err)
    }
}

func TestGetAllUsers_Success(t *testing.T) {
    // Create a new SQL mock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error occurred while opening a stub database connection: %v", err)
    }
    defer db.Close()

    // Create the App instance and set the database connection to the mock
    a := App{db: db}

    // Define the expected rows to be returned by the mock
    rows := sqlmock.NewRows([]string{"username"}).
        AddRow("user1").
        AddRow("user2").
        AddRow("user3")

    // Expect the query with a specific ownerUsername
    mock.ExpectPrepare("SELECT username FROM users WHERE username != ?").ExpectQuery().
        WithArgs("owner").
        WillReturnRows(rows)

    // Call the function to get all users
    users, err := a.getAllUsers("owner")
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Define the expected result
    expectedUsers := []User{
        {Username: "user1"},
        {Username: "user2"},
        {Username: "user3"},
    }

    // Check if the expected result matches the actual result
    if !reflect.DeepEqual(users, expectedUsers) {
        t.Errorf("Expected users to be %v, but got %v", expectedUsers, users)
    }

    // Ensure all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %s", err)
    }
}

func TestShareNoteWithUser(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    noteID := 1
    sharedUsername := "testuser"
    privileges := "read"

    // Define the expected SQL queries and their results using sqlmock
    mock.ExpectQuery("SELECT username").
        WithArgs(sharedUsername).
        WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow(sharedUsername))

    mock.ExpectQuery("SELECT EXISTS").
        WithArgs(noteID).
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

    mock.ExpectQuery("SELECT note_id").
        WithArgs(noteID, sharedUsername).
        WillReturnError(sql.ErrNoRows)

    mock.ExpectExec("INSERT INTO user_shares").
        WithArgs(noteID, sharedUsername, privileges).
        WillReturnResult(sqlmock.NewResult(1, 1))

    err = app.shareNoteWithUser(noteID, sharedUsername, privileges)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }
}

func TestRemoveSharedNoteFromUser(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    username := "testuser"
    noteID := "1"

    // Define the expected SQL query and result using sqlmock
    expectedQuery := "DELETE FROM user_shares"
    mock.ExpectPrepare(expectedQuery).ExpectExec().
        WithArgs(username, noteID).
        WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

    err = app.removeSharedNoteFromUser(username, noteID)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }
}

func TestRemoveDelegation(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    noteID := 123 // Replace with the appropriate noteID

    // Define the expected SQL query and result using sqlmock
    expectedQuery := "UPDATE notes SET noteDelegation = NULL WHERE id = ?"
    mock.ExpectPrepare(expectedQuery).ExpectExec().
        WithArgs(noteID).
        WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

    err = app.RemoveDelegation(noteID)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }
}

func TestGetUnsharedUsersForNote(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    noteID := 123 // Replace with the appropriate noteID
    username := "testuser" // Replace with the appropriate username

    // Define the expected SQL query and result using sqlmock
    expectedQuery := "SELECT username FROM users"
    mock.ExpectPrepare(expectedQuery).ExpectQuery().
        WithArgs(noteID, username).
        WillReturnRows(sqlmock.NewRows([]string{"username"}).
            AddRow("user1").
            AddRow("user2"))

    // Call the getUnsharedUsersForNote function
    unsharedUsers, err := app.getUnsharedUsersForNote(noteID, username)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Check the length and content of the unsharedUsers slice
    if len(unsharedUsers) != 2 {
        t.Errorf("Expected 2 unshared users, but got %d", len(unsharedUsers))
    }
    if unsharedUsers[0].Username != "user1" || unsharedUsers[1].Username != "user2" {
        t.Errorf("Unexpected unshared users: got %v", unsharedUsers)
    }
}

func TestDeleteNoteFromDatabase(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    noteID := 123 // Replace with the appropriate noteID

    // Define the expected SQL query and result using sqlmock
    expectedQuery := "DELETE FROM notes"
    mock.ExpectPrepare(expectedQuery).ExpectExec().
        WithArgs(noteID).
        WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

    err = app.deleteNoteFromDatabase(noteID)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }
}

func TestUpdateUserPrivileges(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    selectedUsername := "testuser" // Replace with the appropriate username
    updatedPrivileges := "read"   // Replace with the updated privileges
    noteID := "123"               // Replace with the appropriate noteID

    // Define the expected SQL query and result using sqlmock
    expectedQuery := "UPDATE user_shares SET privileges"
    mock.ExpectPrepare(expectedQuery).ExpectExec().
        WithArgs(updatedPrivileges, selectedUsername, noteID).
        WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

    err = app.updateUserPrivileges(selectedUsername, updatedPrivileges, noteID)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }
}

func TestGetNoteByID(t *testing.T) {
    // Create a new database connection with sqlmock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Create an instance of your App with the mock database
    app := &App{db: db}

    // Define the expected SQL query and result using sqlmock
    expectedQuery := "SELECT id, title, description, noteType, taskCompletionTime, taskCompletionDate, noteStatus, noteDelegation, owner FROM notes WHERE id = ?"
    expectedNoteID := 123 // Replace with the appropriate noteID
    mock.ExpectQuery(expectedQuery).
        WithArgs(expectedNoteID).
        WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "noteType", "taskCompletionTime", "taskCompletionDate", "noteStatus", "noteDelegation", "owner"}).
            AddRow(123, "Sample Title", "Sample Description", "Type", "2023-11-01", "2023-11-02", "Status", "Delegation", "Owner"),
        )

    // Call the getNoteByID function
    retrievedNote, err := app.getNoteByID(expectedNoteID)

    // Check if there are any expectations that were not met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("there were unfulfilled expectations: %s", err)
    }

    // Check if there was an error
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Check the retrieved note
    if retrievedNote == nil {
        t.Errorf("Expected a non-nil note, but got nil")
    }
    if retrievedNote.ID != 123 || retrievedNote.Title != "Sample Title" || retrievedNote.Description != "Sample Description" {
        t.Errorf("Unexpected note: got %v", retrievedNote)
    }
}

func TestCountOccurrences(t *testing.T) {
    // Test counting occurrences in a text
    text := "This is a sample text with sample words. Sample is a keyword."
    searchPattern := "sample"
    count := countOccurrences(text, searchPattern)

    // Check the count
    if count != 3 {
        t.Errorf("Expected 3 occurrences, but got %d", count)
    }

    // Test with a text that doesn't contain the pattern
    text = "No matching pattern here."
    count = countOccurrences(text, searchPattern)

    // Check the count
    if count != 0 {
        t.Errorf("Expected 0 occurrences, but got %d", count)
    }
}

