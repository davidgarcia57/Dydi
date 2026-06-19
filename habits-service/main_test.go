package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	r := setupRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("expected 'ok', got %q", w.Body.String())
	}
}

func TestCreateCheckin_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"group_id":"g1","habit_id":"h1","checked_on":"2026-06-04"}`)
	req := httptest.NewRequest(http.MethodPost, "/habits/checkins", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateCheckin_MissingFields(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/habits/checkins", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetTodayCheckins_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/habits/checkins/group-123/today", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOpenRoulette_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"group_id":"g1","debtor_id":"u1"}`)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOpenRoulette_MissingFields(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSpin_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette/entry-123/spin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestRequireInternalToken pins the gateway↔services trust boundary: with the
// secret configured, application routes are unreachable without it.
func TestRequireInternalToken(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "secret")
	r := setupRouter(nil)

	// No token → rejected at the gate.
	req := httptest.NewRequest(http.MethodPost, "/habits/checkins", strings.NewReader(`{}`))
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without internal token, got %d", w.Code)
	}

	// Correct token → reaches the handler (400 for missing fields, no DB hit).
	req = httptest.NewRequest(http.MethodPost, "/habits/checkins", strings.NewReader(`{}`))
	req.Header.Set("X-User-ID", "user-123")
	req.Header.Set("X-Internal-Token", "secret")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (handler reached) with correct token, got %d", w.Code)
	}
}

func TestSubmitSuggestion_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"text":"do 20 pushups"}`)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette/entry-123/suggestions", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
