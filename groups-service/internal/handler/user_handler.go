package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dydi/groups-service/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	pool *pgxpool.Pool
}

func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	return &UserHandler{pool: pool}
}

func (h *UserHandler) SyncUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		DisplayName string  `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "display_name is required")
		return
	}

	user, err := db.UpsertUser(r.Context(), h.pool, userID, body.DisplayName, body.AvatarURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not sync user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	if err := db.DeleteUser(r.Context(), h.pool, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
