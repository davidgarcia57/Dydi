package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dydi/realtime-service/internal/domain"
	"github.com/dydi/realtime-service/internal/usecase"
)

func TestHealthEndpoint(t *testing.T) {
	h := usecase.NewHubUseCase()
	go h.Run()
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]int
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if _, ok := body["active_connections"]; !ok {
		t.Fatal("expected 'active_connections' field in response")
	}
}

func TestBroadcastEndpoint(t *testing.T) {
	h := usecase.NewHubUseCase()
	go h.Run()
	r := setupRouter(h)

	ev := domain.Event{
		Type:    domain.EventCheckin,
		GroupID: "group-123",
		UserID:  "user-456",
	}
	body, _ := json.Marshal(ev)

	req := httptest.NewRequest(http.MethodPost, "/internal/broadcast", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestBroadcastEndpointInvalidBody(t *testing.T) {
	h := usecase.NewHubUseCase()
	go h.Run()
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/internal/broadcast", bytes.NewReader([]byte("not-json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
