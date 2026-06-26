package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dydi/habits-service/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// internalClient is used for service-to-service calls (realtime broadcast).
// It has a bounded timeout so a hung downstream never blocks a goroutine forever.
var internalClient = &http.Client{Timeout: 10 * time.Second}

type HabitHandler struct {
	pool        *pgxpool.Pool
	realtimeURL string
}

func NewHabitHandler(pool *pgxpool.Pool, realtimeURL string) *HabitHandler {
	return &HabitHandler{pool: pool, realtimeURL: realtimeURL}
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

	// checked_on is the client's local date (to avoid UTC drift), but it must
	// stay within a day of the server's date. Otherwise a member could backfill
	// arbitrary past dates to fake streaks or dodge roulette eligibility.
	checkedOn, perr := time.Parse("2006-01-02", body.CheckedOn)
	if perr != nil {
		writeError(w, http.StatusBadRequest, "checked_on must be a valid YYYY-MM-DD date")
		return
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if delta := checkedOn.Sub(today); delta < -24*time.Hour || delta > 24*time.Hour {
		writeError(w, http.StatusBadRequest, "checked_on is outside the allowed range")
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

	go h.notifyRealtime(body.GroupID, userID, "checkin", map[string]any{
		"user_id":  userID,
		"habit_id": body.HabitID,
		"status":   "done",
	})

	w.WriteHeader(http.StatusCreated)
}

func (h *HabitHandler) notifyRealtime(groupID, userID, eventType string, data any) {
	if h.realtimeURL == "" {
		return
	}
	payload, err := json.Marshal(map[string]any{
		"type":    eventType,
		"groupID": groupID,
		"userID":  userID,
		"payload": data,
	})
	if err != nil {
		log.Printf("notifyRealtime: marshal error: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.realtimeURL+"/internal/broadcast", bytes.NewReader(payload))
	if err != nil {
		log.Printf("notifyRealtime: request build error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if tok := os.Getenv("INTERNAL_TOKEN"); tok != "" {
		req.Header.Set("X-Internal-Token", tok)
	}
	resp, err := internalClient.Do(req)
	if err != nil {
		log.Printf("notifyRealtime: broadcast to %s failed: %v", h.realtimeURL, err)
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("notifyRealtime: broadcast returned %d", resp.StatusCode)
	}
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

	// Only members may read a group's check-ins (matches GetHistory / debts).
	member, err := db.IsMemberOfGroup(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	checkins, err := db.GetTodayCheckinsByGroup(r.Context(), h.pool, groupID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, checkins)
}

// GetHistory returns check-in history for a group within [from, to] so the
// frontend can render real 7-day strips. Caller must be a member of the group.
func (h *HabitHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "from and to query params are required")
		return
	}

	member, err := db.IsMemberOfGroup(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	history, err := db.GetCheckinHistory(r.Context(), h.pool, groupID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (h *HabitHandler) GetStreaks(w http.ResponseWriter, r *http.Request) {
	callerID := r.Header.Get("X-User-ID")
	if callerID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	targetID := chi.URLParam(r, "userID")

	// A user may read their own streaks, or those of someone they share a group
	// with (the squad/today views show co-members' streaks). Anything else is a
	// cross-user data leak.
	if callerID != targetID {
		shares, err := db.UsersShareGroup(r.Context(), h.pool, callerID, targetID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "db error")
			return
		}
		if !shares {
			writeError(w, http.StatusForbidden, "not allowed to view this user's streaks")
			return
		}
	}

	streaks, err := db.GetStreaksForUser(r.Context(), h.pool, targetID)
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
	// Service-to-service auth is enforced by the requireInternalToken middleware
	// on this route (the endpoint is internet-reachable on Render).
	var body struct {
		GroupID string `json:"group_id"`
		HabitID string `json:"habit_id"`
		Action  string `json:"action"`   // "add" | "remove"
		AddedBy string `json:"added_by"` // proposer; recorded as group_habits.added_by on add
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == "" || body.HabitID == "" {
		writeError(w, http.StatusBadRequest, "group_id, habit_id and action are required")
		return
	}

	switch body.Action {
	case "add":
		if body.AddedBy == "" {
			writeError(w, http.StatusBadRequest, "added_by is required for add")
			return
		}
		if err := db.BulkAssignHabit(r.Context(), h.pool, body.GroupID, body.HabitID, body.AddedBy); err != nil {
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
