package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dydi/habits-service/internal/model"
)

var testCatalog = []model.Punishment{
	{ID: 1, Text: "Do 10 pushups", Emoji: "💪", Category: "exercise"},
}

func TestHealthEndpoint(t *testing.T) {
	r := setupRouter(nil, testCatalog)
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

func TestAssignHabit_MissingUserID(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	body := strings.NewReader(`{"group_id":"g1","habit_id":"h1"}`)
	req := httptest.NewRequest(http.MethodPost, "/habits/assign", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAssignHabit_MissingFields(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/habits/assign", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateCheckin_MissingUserID(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	body := strings.NewReader(`{"group_id":"g1","habit_id":"h1"}`)
	req := httptest.NewRequest(http.MethodPost, "/checkins", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateCheckin_MissingFields(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/checkins", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSpin_MissingUserID(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	body := strings.NewReader(`{"group_id":"g1","debtor_id":"u1"}`)
	req := httptest.NewRequest(http.MethodPost, "/penalties/spin", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSpin_MissingFields(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/penalties/spin", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetTodayCheckins_MissingUserID(t *testing.T) {
	r := setupRouter(nil, testCatalog)
	req := httptest.NewRequest(http.MethodGet, "/checkins/group-123/today", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
