package main

import (
	"database/sql"

	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	// Import other necessary packages for your tests
)

// Define a mock session struct that embeds the mock.Mock struct
type MockSession struct {
    mock.Mock
}

// Implement the Get method on the mock session
func (m *MockSession) Get(key string) interface{} {
    args := m.Called(key)
    return args.Get(0)
}


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
        AddRow("user1", "editor").
        AddRow("user2", "viewer")

    // Expect the query with a specific noteID
    mock.ExpectPrepare("SELECT username, privileges FROM user_shares WHERE note_id = ?").ExpectQuery().
        WithArgs(123).WillReturnRows(rows)

    // Call the function with the noteID
    sharedUsers, err := a.getSharedUsersForNote(123)
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Define the expected result
    expectedUsers := []UserShare{
        {Username: sql.NullString{String: "user1", Valid: true}, Privileges: sql.NullString{String: "editor", Valid: true}},
        {Username: sql.NullString{String: "user2", Valid: true}, Privileges: sql.NullString{String: "viewer", Valid: true}},
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

func TestGetUnsharedUsersForNoteHandler(t *testing.T) {
    

    // Create a new HTTP request for testing
    req := httptest.NewRequest("GET", "/your-endpoint/{noteID}", nil)
    req = mux.SetURLVars(req, map[string]string{"noteID": "123"})

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()



    // Create a new SQL mock
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("An error occurred while opening a stub database connection: %v", err)
    }
    defer db.Close()

    // Create the App instance and set the database connection to the mock
    a := App{db: db}

    

    // Expect the query with a specific noteID
    rows := sqlmock.NewRows([]string{"username"}).
        AddRow("user1").
        AddRow("user2")

    mock.ExpectPrepare("SELECT username").ExpectQuery().
        WithArgs(123, "testuser").
        WillReturnRows(rows)


    // Fetch the unshared users for the given noteID
    unsharedUsers, err := a.getUnsharedUsersForNote(123, "testuser")
    if err != nil {
        t.Errorf("Expected no error, but got %v", err)
    }

    // Check the HTTP status code (e.g., 200 OK for success)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

	
    // Define the expected result
    expectedUsers := []User{
        {Username: "user1"},
        {Username: "user2"},
    }

    // Check if the expected result matches the actual result
    if !reflect.DeepEqual(unsharedUsers, expectedUsers) {
        t.Errorf("Expected sharedUsers to be %v, but got %v", expectedUsers, unsharedUsers)
    }

    // Ensure all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %s", err)
    }

    

}

func TestSearchNotesHandler(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

    // Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock HTTP request with GET method (simulating a successful request)
    req := httptest.NewRequest("GET", "/search", nil)

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the searchNotesHandler function
    a.searchNotesHandler(rr, req)

    // Check the HTTP status code (200 OK for success)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // You can add more assertions to check the response body, headers, or other aspects.
}

func TestCreateHandler(t *testing.T) {
    
    // Create the App instance and set the database connection to the mock
    a := App{}
	a.Initialize()

    // Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock HTTP request with POST method (simulating a successful request)
    form := url.Values{}
    form.Add("Title", "Test Title")
    form.Add("NoteType", "Test Type")
    form.Add("Description", "Test Description")
    form.Add("TaskCompletionDate", "2023-11-01")
    form.Add("TaskCompletionTime", "13:00:00")
    form.Add("NoteStatus", "Test Status")
    form.Add("NoteDelegation", "Test Delegation")
	
    req := httptest.NewRequest("POST", "/create", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the createHandler function
    a.createHandler(rr, req)

    // Check the HTTP status code (e.g., 302 Found for a redirect to "/list")
    if status := rr.Code; status != http.StatusSeeOther {
        t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusSeeOther)
    }

}

func TestUpdateHandler(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

    // Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock HTTP request with POST method (simulating a successful request)
    form := url.Values{}
    form.Add("Id", "1")
    form.Add("Title", "Updated Title")
    form.Add("NoteType", "Updated Type")
    form.Add("Description", "Updated Description")
    form.Add("TaskCompletionDate", "2023-11-02")
    form.Add("TaskCompletionTime", "14:00:00")
    form.Add("NoteStatus", "Updated Status")
    form.Add("NoteDelegation", "Updated Delegation")

    req := httptest.NewRequest("POST", "/update", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the updateHandler function
    a.updateHandler(rr, req)

    // Check the HTTP status code (e.g., 302 Found for a redirect to "/list")
    if status := rr.Code; status != http.StatusSeeOther {
        t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusSeeOther)
    }

    // You can add more assertions to check the behavior after the redirect, such as checking the updated note in the database.

}

func TestDeleteHandler(t *testing.T) {
	// Create a new instance of your application
	a := App{}
	a.Initialize()

	// Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
	os.Setenv("DISABLE_AUTH", "1")

	// Create a mock HTTP request with POST method (simulating a successful request)
	form := url.Values{}
	form.Add("Id", "1")

	req := httptest.NewRequest("POST", "/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the deleteHandler function
	a.deleteHandler(rr, req)

	// Check the HTTP status code (e.g., 302 Found for a redirect to "/list")
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusSeeOther)
	}

	// You can add more assertions to check the behavior after the redirect, such as checking the deleted note in the database.

}

func TestUpdatePrivilegesHandler(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

	// Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock HTTP request with POST method and form data
    form := url.Values{
        "username":   {"testuser"},   // Replace with the appropriate username
        "privileges": {"read"},       // Replace with the updated privileges
        "noteID":     {"1"},          // Replace with the appropriate noteID
    }
    body := strings.NewReader(form.Encode())
    req := httptest.NewRequest("POST", "/update-privileges", body)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the updatePrivilegesHandler function
    a.updatePrivilegesHandler(rr, req)

    // Check the HTTP status code (302 Found for a redirect)
    if status := rr.Code; status != http.StatusSeeOther {
        t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusSeeOther)
    }

    // You can add more assertions to check the response body, headers, or other aspects.
}

func TestFindInNoteHandler(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

	// Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock HTTP request with a valid noteID and search pattern
    req := httptest.NewRequest("GET", "/find-note/1?searchInput=searchtext", nil)
    req = mux.SetURLVars(req, map[string]string{"noteID": "1"})

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the findInNoteHandler function
    a.findInNoteHandler(rr, req)

    // Check the HTTP status code (200 OK for success)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
    }

    // You can add assertions to check the response body, headers, or other aspects.
}

func TestIndexHandler(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

	// Set the DISABLE_AUTH environment variable to "1" to disable authentication checks
    os.Setenv("DISABLE_AUTH", "1")

    // Create a mock authenticated request
    req := httptest.NewRequest("GET", "/index", nil)
    req.AddCookie(&http.Cookie{Name: "authCookie", Value: "authToken"}) // Replace with your authentication method

    // Create a ResponseRecorder to capture the response
    rr := httptest.NewRecorder()

    // Call the indexHandler function
    a.indexHandler(rr, req)

    // Check the HTTP status code (303 See Other for redirect)
    if status := rr.Code; status != http.StatusSeeOther {
        t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusSeeOther)
    }

    // You can add more assertions to check the redirect location and other aspects.
}



