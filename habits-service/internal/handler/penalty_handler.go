package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"time"

	"github.com/dydi/habits-service/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const suggestionWindowHours = 48

type PenaltyHandler struct {
	pool        *pgxpool.Pool
	realtimeURL string
}

func NewPenaltyHandler(pool *pgxpool.Pool, realtimeURL string) *PenaltyHandler {
	return &PenaltyHandler{pool: pool, realtimeURL: realtimeURL}
}

func (h *PenaltyHandler) GetEligible(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")

	member, err := db.IsMemberOfGroup(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	members, err := db.GetEligibleMembers(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// OpenRoulette creates a roulette entry for an eligible debtor and starts the
// suggestion window. Any group member can call this.
func (h *PenaltyHandler) OpenRoulette(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	var body struct {
		GroupID  string `json:"group_id"`
		DebtorID string `json:"debtor_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.GroupID == "" || body.DebtorID == "" {
		writeError(w, http.StatusBadRequest, "group_id and debtor_id are required")
		return
	}

	member, err := db.IsMemberOfGroup(r.Context(), h.pool, body.GroupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	weekStart := db.CurrentWeekStart().Format("2006-01-02")
	deadline := time.Now().UTC().Add(suggestionWindowHours * time.Hour)

	entry, err := db.GetOrCreateRouletteEntry(r.Context(), h.pool, body.GroupID, body.DebtorID, weekStart, deadline)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusCreated, entry)
}

// SubmitSuggestion adds a punishment suggestion to a roulette entry.
// Any group member (including the debtor) can submit one suggestion per entry.
func (h *PenaltyHandler) SubmitSuggestion(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	entryID := chi.URLParam(r, "entryID")

	entry, err := db.GetRouletteEntry(r.Context(), h.pool, entryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "roulette entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	member, err := db.IsMemberOfGroup(r.Context(), h.pool, entry.GroupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	if time.Now().After(entry.SuggestionDeadline) {
		writeError(w, http.StatusConflict, "suggestion window has closed")
		return
	}

	already, err := db.HasSuggested(r.Context(), h.pool, entryID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if already {
		writeError(w, http.StatusConflict, "already submitted a suggestion for this entry")
		return
	}

	var body struct {
		Text  string  `json:"text"`
		Emoji *string `json:"emoji,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	suggestion, err := db.CreateSuggestion(r.Context(), h.pool, entryID, entry.GroupID, userID, body.Text, body.Emoji)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not submit suggestion")
		return
	}

	writeJSON(w, http.StatusCreated, suggestion)
}

// GetSuggestions returns all suggestions for a roulette entry.
func (h *PenaltyHandler) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	entryID := chi.URLParam(r, "entryID")

	entry, err := db.GetRouletteEntry(r.Context(), h.pool, entryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "roulette entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	member, err := db.IsMemberOfGroup(r.Context(), h.pool, entry.GroupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	suggestions, err := db.GetSuggestionsForEntry(r.Context(), h.pool, entryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, suggestions)
}

// Spin picks a random suggestion and assigns the debt. Only the debtor themselves
// can spin, and only after the suggestion deadline has passed.
// If no suggestions were submitted the whole group receives a collective debt.
func (h *PenaltyHandler) Spin(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	entryID := chi.URLParam(r, "entryID")

	entry, err := db.GetRouletteEntry(r.Context(), h.pool, entryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "roulette entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	if entry.DebtorID != userID {
		writeError(w, http.StatusForbidden, "only the debtor can spin their own roulette")
		return
	}
	if entry.SpunAt != nil {
		writeError(w, http.StatusConflict, "already spun for this entry")
		return
	}
	if time.Now().Before(entry.SuggestionDeadline) {
		writeError(w, http.StatusConflict, "suggestion window has not closed yet")
		return
	}

	suggestions, err := db.GetSuggestionsForEntry(r.Context(), h.pool, entryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	weekStart := entry.WeekStart.Format("2006-01-02")

	if len(suggestions) == 0 {
		// No suggestions → collective punishment for all group members.
		debts, err := db.CreateCollectiveDebts(r.Context(), tx, entryID, entry.GroupID, weekStart)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create collective debts")
			return
		}
		if err := db.MarkEntrySpun(r.Context(), tx, entryID); err != nil {
			writeError(w, http.StatusInternalServerError, "could not finalize spin")
			return
		}
		if err := tx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, "db error")
			return
		}
		go h.notifyRealtime(entry.GroupID, "collective_punishment", debts)
		writeJSON(w, http.StatusCreated, debts)
		return
	}

	// Pick a random suggestion.
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(suggestions))))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not select suggestion")
		return
	}
	winner := suggestions[idx.Int64()]

	debt, err := db.CreateDebt(r.Context(), tx, entryID, entry.GroupID, entry.DebtorID, weekStart, &winner.ID, winner.Text, winner.Emoji)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create debt")
		return
	}
	if err := db.MarkEntrySpun(r.Context(), tx, entryID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not finalize spin")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	go h.notifyRealtime(entry.GroupID, "roulette_result", debt)
	writeJSON(w, http.StatusCreated, debt)
}

func (h *PenaltyHandler) GetActiveDebts(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")

	member, err := db.IsMemberOfGroup(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	debts, err := db.GetActiveDebts(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, debts)
}

// notifyRealtime fires a broadcast to realtime-service. Errors are ignored
// so a realtime outage never fails the spin.
func (h *PenaltyHandler) notifyRealtime(groupID, eventType string, data any) {
	if h.realtimeURL == "" {
		return
	}

	payload, err := json.Marshal(map[string]any{
		"type":    eventType,
		"groupID": groupID,
		"payload": data,
	})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.realtimeURL+"/internal/broadcast", bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req) //nolint:errcheck
}
