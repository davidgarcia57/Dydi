package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/dydi/habits-service/internal/db"
	"github.com/dydi/habits-service/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PenaltyHandler struct {
	pool        *pgxpool.Pool
	catalog     []model.Punishment
	realtimeURL string
}

func NewPenaltyHandler(pool *pgxpool.Pool, catalog []model.Punishment, realtimeURL string) *PenaltyHandler {
	return &PenaltyHandler{pool: pool, catalog: catalog, realtimeURL: realtimeURL}
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

func (h *PenaltyHandler) Spin(w http.ResponseWriter, r *http.Request) {
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

	punishment, err := h.randomPunishment()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not select punishment")
		return
	}
	var emoji *string
	if punishment.Emoji != "" {
		emoji = &punishment.Emoji
	}

	// Wrap the three writes in a transaction so a mid-flight failure cannot
	// leave a debt without a completed entry (or vice-versa).
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck — rollback on any early return

	draw, err := db.GetOrCreateRouletteDraw(r.Context(), tx, body.GroupID, body.DebtorID, weekStart)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if draw.SpunAt != nil {
		writeError(w, http.StatusConflict, "already spun for this member this week")
		return
	}

	debt, err := db.CreateDebt(r.Context(), tx, draw.ID, punishment.Text, emoji)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create debt")
		return
	}

	if err := db.MarkDrawCompleted(r.Context(), tx, draw.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not finalize spin")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	go h.notifyRealtime(body.GroupID, debt)
	writeJSON(w, http.StatusCreated, debt)
}

func (h *PenaltyHandler) GetPendingDebts(w http.ResponseWriter, r *http.Request) {
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

	debts, err := db.GetPendingDebts(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, debts)
}

func (h *PenaltyHandler) ResolveDebt(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	debtID := chi.URLParam(r, "id")

	if err := db.ResolveDebt(r.Context(), h.pool, debtID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not resolve debt")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PenaltyHandler) randomPunishment() (model.Punishment, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(h.catalog))))
	if err != nil {
		return model.Punishment{}, err
	}
	return h.catalog[n.Int64()], nil
}

// notifyRealtime fires a broadcast to realtime-service. Errors are ignored
// so a realtime outage never fails the spin.
func (h *PenaltyHandler) notifyRealtime(groupID string, debt *model.Debt) {
	if h.realtimeURL == "" {
		return
	}

	payload, err := json.Marshal(map[string]any{
		"group_id": groupID,
		"event":    "roulette_result",
		"payload":  debt,
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
	http.DefaultClient.Do(req)
}
