package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	// Import other necessary packages for your tests
)




func TestListHandler(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

	// Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock HTTP request with GET method (simulating a successful request)
    req := httptest.NewRequest("GET", "/list", nil)

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the listHandler function
    a.listHandler(rr, req)

    // Check the HTTP status code (200 OK for success)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // You can add more assertions to check the response body, headers, or other aspects.
}

func TestGetSharedUsersForNote_Success(t *testing.T) {
    // Create a new SQL mock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error occurred while opening a stub database connection: %v", err)
    }
    defer db.Close()

    // Create the App instance and set the database connection to the mock
    a := App{db: db}

    // Define the expected rows to be returned by the mock
    rows := sqlmock.NewRows([]string{"username", "privileges"}).
        AddRow("user1", "read").
        AddRow("user2", "write")

    // Expect the query with a specific noteID
    mock.ExpectQuery("SELECT username, privileges FROM user_shares WHERE note_id = ?").
        WithArgs(123).WillReturnRows(rows)

    // Call the function with the noteID
    sharedUsers, err := a.getSharedUsersForNote(123)
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Define the expected result
    expectedUsers := []UserShare{
        {Username: sql.NullString{String: "user1", Valid: true}, Privileges: sql.NullString{String: "read", Valid: true}},
        {Username: sql.NullString{String: "user2", Valid: true}, Privileges: sql.NullString{String: "write", Valid: true}},
    }

    // Check if the expected result matches the actual result
    if !reflect.DeepEqual(sharedUsers, expectedUsers) {
        t.Errorf("Expected sharedUsers to be %v, but got %v", expectedUsers, sharedUsers)
    }

    // Ensure all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %s", err)
    }
}

func TestGetSharedUsersForNoteHandler_Success(t *testing.T) {
    // Create a new App instance
    a := App{
        // You can set up a mock database connection or use an in-memory database for this test
        // db: yourMockDB,
    }

    // Create a request with a valid noteID
    req, err := http.NewRequest("GET", "/shared-users/123", nil)
    if err != nil {
        t.Fatalf("Failed to create a request: %v", err)
    }

    // Create a ResponseRecorder to record the response
    rr := httptest.NewRecorder()

    // Define the expected shared users
    expectedSharedUsers := []UserShare{
        {Username: sql.NullString{String: "user1", Valid: true}, Privileges: sql.NullString{String: "read", Valid: true}},
        {Username: sql.NullString{String: "user2", Valid: true}, Privileges: sql.NullString{String: "write", Valid: true}},
    }

    // Mock the getSharedUsersForNote function
    a.getSharedUsersForNote = func(noteID int) ([]UserShare, error) {
        if noteID != 123 {
            return nil, errors.New("Invalid noteID")
        }
        return expectedSharedUsers, nil
    }

    // Serve the request using the handler
    a.Router.ServeHTTP(rr, req)

    // Check the response status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Define the expected JSON response
    expectedResponse := `[{"Username":{"String":"user1","Valid":true},"Privileges":{"String":"read","Valid":true}},{"Username":{"String":"user2","Valid":true},"Privileges":{"String":"write","Valid":true}}]`

    // Check if the response body matches the expected JSON
    if rr.Body.String() != expectedResponse {
        t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expectedResponse)
    }
}
