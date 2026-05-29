package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/dydi/groups-service/internal/db"
	"github.com/dydi/groups-service/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupHandler struct {
	pool         *pgxpool.Pool
	maxGroupSize int
}

func NewGroupHandler(pool *pgxpool.Pool) *GroupHandler {
	max := 8
	if v := os.Getenv("MAX_GROUP_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			max = n
		}
	}
	return &GroupHandler{pool: pool, maxGroupSize: max}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	code, err := db.GenerateInviteCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate invite code")
		return
	}

	group, err := db.CreateGroup(r.Context(), h.pool, body.Name, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create group")
		return
	}

	if err := db.AddMember(r.Context(), h.pool, group.ID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not add creator to group")
		return
	}

	writeJSON(w, http.StatusCreated, group)
}

func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	group, err := db.GetGroupByID(r.Context(), h.pool, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	members, err := db.GetMembers(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusOK, model.GroupWithMembers{Group: *group, Members: members})
}

func (h *GroupHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	var body struct {
		InviteCode string `json:"invite_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.InviteCode == "" {
		writeError(w, http.StatusBadRequest, "invite_code is required")
		return
	}

	group, err := db.GetGroupByID(r.Context(), h.pool, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// Respond with 404 instead of 403 to avoid leaking group existence
	if group.InviteCode != body.InviteCode {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}

	already, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if already {
		writeError(w, http.StatusConflict, "already a member")
		return
	}

	count, err := db.CountMembers(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if count >= h.maxGroupSize {
		writeError(w, http.StatusConflict, "group is full")
		return
	}

	if err := db.AddMember(r.Context(), h.pool, groupID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not join group")
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func (h *GroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	members, err := db.GetMembers(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "id")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	if err := db.RemoveMember(r.Context(), h.pool, groupID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not leave group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
