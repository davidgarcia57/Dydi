package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToStripsAPIPrefix(t *testing.T) {
	var receivedPath string
	var receivedQuery string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/users/sync?source=test", nil)
	w := httptest.NewRecorder()

	To(upstream.URL).ServeHTTP(w, req)

	if receivedPath != "/users/sync" {
		t.Fatalf("expected /users/sync, got %q", receivedPath)
	}
	if receivedQuery != "source=test" {
		t.Fatalf("expected query to be preserved, got %q", receivedQuery)
	}
}
