// authentication_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAuthenticated(t *testing.T) {
	app := &application{
		sessions: make(map[string]int),
	}

	// Test authenticated
	req := httptest.NewRequest("GET", "/", nil)
	sessionID := "test_session"
	app.sessions[sessionID] = 1
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

	if !app.isAuthenticated(req) {
		t.Error("Expected true for authenticated user, got false")
	}

	// Test not authenticated
	req = httptest.NewRequest("GET", "/", nil)
	if app.isAuthenticated(req) {
		t.Error("Expected false for unauthenticated user, got true")
	}
}
