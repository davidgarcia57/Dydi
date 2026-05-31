package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dydi/groups-service/internal/db"
	"github.com/dydi/groups-service/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProposalHandler struct {
	pool *pgxpool.Pool
}

func NewProposalHandler(pool *pgxpool.Pool) *ProposalHandler {
	return &ProposalHandler{pool: pool}
}

func (h *ProposalHandler) CreateProposal(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")

	var body struct {
		Type    model.ProposalType `json:"type"`
		Payload json.RawMessage    `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	switch body.Type {
	case model.ProposalAddHabit, model.ProposalRemoveHabit, model.ProposalKickMember, model.ProposalDeleteGroup:
	default:
		writeError(w, http.StatusBadRequest, "invalid proposal type")
		return
	}

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	if body.Payload == nil {
		body.Payload = json.RawMessage(`{}`)
	}

	proposal, err := db.CreateProposal(r.Context(), h.pool, groupID, userID, body.Type, body.Payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create proposal")
		return
	}

	writeJSON(w, http.StatusCreated, proposal)
}

func (h *ProposalHandler) ListProposals(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")

	member, err := db.IsMember(r.Context(), h.pool, groupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	proposals, err := db.ListOpenProposals(r.Context(), h.pool, groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	if proposals == nil {
		proposals = []model.Proposal{}
	}
	writeJSON(w, http.StatusOK, proposals)
}

func (h *ProposalHandler) Vote(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	proposalID := chi.URLParam(r, "proposalID")

	proposal, err := db.GetProposal(r.Context(), h.pool, proposalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "proposal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	if proposal.Status != model.ProposalOpen {
		writeError(w, http.StatusConflict, "proposal is no longer open")
		return
	}

	member, err := db.IsMember(r.Context(), h.pool, proposal.GroupID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !member {
		writeError(w, http.StatusForbidden, "not a member of this group")
		return
	}

	already, err := db.HasVoted(r.Context(), h.pool, proposalID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if already {
		writeError(w, http.StatusConflict, "already voted")
		return
	}

	var body struct {
		Approved bool `json:"approved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := db.CastVote(r.Context(), h.pool, proposalID, userID, body.Approved); err != nil {
		writeError(w, http.StatusInternalServerError, "could not cast vote")
		return
	}

	approvals, members, err := db.CountApprovalVotes(r.Context(), h.pool, proposalID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// Auto-resolve: strict majority (>50%) approves
	if members > 0 && approvals*2 > members {
		if err := h.executeProposal(r, proposal); err != nil {
			writeError(w, http.StatusInternalServerError, "could not execute proposal")
			return
		}
		_ = db.SetProposalStatus(r.Context(), h.pool, proposalID, model.ProposalApproved)
	}

	w.WriteHeader(http.StatusNoContent)
}

// executeProposal applies the side-effects of an approved proposal.
func (h *ProposalHandler) executeProposal(r *http.Request, p *model.Proposal) error {
	ctx := r.Context()

	switch p.Type {
	case model.ProposalAddHabit, model.ProposalRemoveHabit:
		// Habit assignment is owned by habits-service; groups-service only tracks the vote.
		// The frontend calls habits-service directly after seeing status=approved.
		return nil

	case model.ProposalKickMember:
		var payload struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(p.Payload, &payload); err != nil || payload.UserID == "" {
			return nil
		}
		return db.RemoveMember(ctx, h.pool, p.GroupID, payload.UserID)

	case model.ProposalDeleteGroup:
		_, err := h.pool.Exec(ctx, `DELETE FROM groups WHERE id = $1`, p.GroupID)
		return err
	}

	return nil
}
