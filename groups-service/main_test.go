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

func TestCreateGroup_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"name":"los champions"}`)
	req := httptest.NewRequest(http.MethodPost, "/groups", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateGroup_MissingName(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/groups", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestJoinGroup_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"invite_code":"ABC12345"}`)
	req := httptest.NewRequest(http.MethodPost, "/groups/some-id/join", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestJoinGroup_MissingInviteCode(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/groups/some-id/join", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetGroup_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/groups/some-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLeaveGroup_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	req := httptest.NewRequest(http.MethodDelete, "/groups/some-id/leave", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncUser_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"display_name":"David"}`)
	req := httptest.NewRequest(http.MethodPost, "/users/sync", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSyncUser_MissingDisplayName(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/users/sync", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListMyGroups_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateProposal_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"type":"add_habit","payload":{"habit_id":"abc"}}`)
	req := httptest.NewRequest(http.MethodPost, "/groups/some-group/proposals", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestVote_MissingUserID(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"approved":true}`)
	req := httptest.NewRequest(http.MethodPost, "/proposals/some-proposal/vote", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestRequireInternalToken pins the gateway↔services trust boundary: when the
// secret is configured, application routes are unreachable without it (so an
// internet caller can't forge X-User-ID by hitting the service directly).
func TestRequireInternalToken(t *testing.T) {
	t.Setenv("INTERNAL_TOKEN", "secret")
	r := setupRouter(nil)

	// No token → rejected at the gate.
	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without internal token, got %d", w.Code)
	}

	// Wrong token → rejected.
	req = httptest.NewRequest(http.MethodGet, "/groups", nil)
	req.Header.Set("X-User-ID", "user-123")
	req.Header.Set("X-Internal-Token", "nope")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong internal token, got %d", w.Code)
	}

	// Correct token → reaches the handler (400 for the missing name, no DB hit).
	req = httptest.NewRequest(http.MethodPost, "/groups", strings.NewReader(`{}`))
	req.Header.Set("X-User-ID", "user-123")
	req.Header.Set("X-Internal-Token", "secret")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (handler reached) with correct token, got %d", w.Code)
	}

	// Health stays public even with the secret configured.
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /health to stay public (200), got %d", w.Code)
	}
}

func TestCreateProposal_InvalidType(t *testing.T) {
	r := setupRouter(nil)
	body := strings.NewReader(`{"type":"ban_user","payload":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/groups/some-group/proposals", body)
	req.Header.Set("X-User-ID", "user-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
