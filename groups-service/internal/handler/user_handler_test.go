package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Sin la identidad estampada por el gateway, el handler debe rechazar antes de
// tocar la base (el pool es nil aquí: llegar a él haría panic — eso es el test).
func TestDeleteMeRequiresUserHeader(t *testing.T) {
	h := NewUserHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)

	h.DeleteMe(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
