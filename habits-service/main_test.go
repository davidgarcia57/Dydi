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
