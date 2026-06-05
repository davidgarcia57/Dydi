package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dydi/habits-service/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HabitHandler struct {
	pool *pgxpool.Pool
}

func NewHabitHandler(pool *pgxpool.Pool) *HabitHandler {
	return &HabitHandler{pool: pool}
}

func (h *HabitHandler) ListHabits(w http.ResponseWriter, r *http.Request) {
	habits, err := db.ListHabits(r.Context(), h.pool)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, habits)
}

func (h *HabitHandler) CreateCheckin(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		GroupID   string  `json:"group_id"`
		HabitID   string  `json:"habit_id"`
		Note      *string `json:"note,omitempty"`
		CheckedOn string  `json:"checked_on"` // "YYYY-MM-DD", local date from client
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == "" || body.HabitID == "" || body.CheckedOn == "" {
		writeError(w, http.StatusBadRequest, "group_id, habit_id and checked_on are required")
		return
	}

	userHabitID, err := db.FindUserHabitID(r.Context(), h.pool, userID, body.GroupID, body.HabitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "habit not assigned to this user in this group")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	already, err := db.HasCheckinOnDate(r.Context(), h.pool, userHabitID, body.CheckedOn)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if already {
		writeError(w, http.StatusConflict, "already checked in today")
		return
	}

	if err := db.CreateCheckin(r.Context(), h.pool, userHabitID, body.CheckedOn, body.Note); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create checkin")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *HabitHandler) GetTodayCheckins(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")
	date := r.URL.Query().Get("date") // "YYYY-MM-DD", local date from client
	if date == "" {
		writeError(w, http.StatusBadRequest, "date query param is required")
		return
	}

	checkins, err := db.GetTodayCheckinsByGroup(r.Context(), h.pool, groupID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, checkins)
}

func (h *HabitHandler) GetStreaks(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	streaks, err := db.GetStreaksForUser(r.Context(), h.pool, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, streaks)
}

// ApplyProposal is an internal endpoint called by groups-service when a proposal
// is approved. It bulk-assigns or bulk-removes a habit for all group members.
// Not exposed through api-gateway.
func (h *HabitHandler) ApplyProposal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GroupID string `json:"group_id"`
		HabitID string `json:"habit_id"`
		Action  string `json:"action"` // "add" | "remove"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == "" || body.HabitID == "" {
		writeError(w, http.StatusBadRequest, "group_id, habit_id and action are required")
		return
	}

	switch body.Action {
	case "add":
		if err := db.BulkAssignHabit(r.Context(), h.pool, body.GroupID, body.HabitID); err != nil {
			writeError(w, http.StatusInternalServerError, "could not assign habit")
			return
		}
	case "remove":
		if err := db.BulkUnassignHabit(r.Context(), h.pool, body.GroupID, body.HabitID); err != nil {
			writeError(w, http.StatusInternalServerError, "could not remove habit")
			return
		}
	default:
		writeError(w, http.StatusBadRequest, "action must be 'add' or 'remove'")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
