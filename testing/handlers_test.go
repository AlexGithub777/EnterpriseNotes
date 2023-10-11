package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	// Import other necessary packages for your tests
)




func TestListHandler_Success(t *testing.T) {
    // Create a new instance of your application
    a := App{}
    a.Initialize()

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
