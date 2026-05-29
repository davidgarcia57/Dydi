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

func (h *HabitHandler) AssignHabit(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		GroupID string `json:"group_id"`
		HabitID string `json:"habit_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == "" || body.HabitID == "" {
		writeError(w, http.StatusBadRequest, "group_id and habit_id are required")
		return
	}

	uh, err := db.AssignHabit(r.Context(), h.pool, userID, body.GroupID, body.HabitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT DO NOTHING returned nothing — already assigned
			writeError(w, http.StatusConflict, "habit already assigned")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not assign habit")
		return
	}
	writeJSON(w, http.StatusCreated, uh)
}

func (h *HabitHandler) CreateCheckin(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		GroupID string  `json:"group_id"`
		HabitID string  `json:"habit_id"`
		Note    *string `json:"note,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == "" || body.HabitID == "" {
		writeError(w, http.StatusBadRequest, "group_id and habit_id are required")
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

	already, err := db.HasCheckinToday(r.Context(), h.pool, userHabitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if already {
		writeError(w, http.StatusConflict, "already checked in today")
		return
	}

	if err := db.CreateCheckin(r.Context(), h.pool, userHabitID, body.Note); err != nil {
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

	checkins, err := db.GetTodayCheckinsByGroup(r.Context(), h.pool, groupID)
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
