package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dydi/groups-service/internal/db"
	"github.com/dydi/groups-service/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProposalHandler struct {
	pool            *pgxpool.Pool
	habitsServiceURL string
}

func NewProposalHandler(pool *pgxpool.Pool, habitsServiceURL string) *ProposalHandler {
	return &ProposalHandler{pool: pool, habitsServiceURL: habitsServiceURL}
}

func (h *ProposalHandler) CreateProposal(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "missing X-User-ID")
		return
	}

	groupID := chi.URLParam(r, "groupID")

	var body struct {
		Type         model.ProposalType `json:"type"`
		HabitID      *string            `json:"habit_id,omitempty"`
		TargetUserID *string            `json:"target_user_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	switch body.Type {
	case model.ProposalAddHabit, model.ProposalRemoveHabit:
		if body.HabitID == nil || *body.HabitID == "" {
			writeError(w, http.StatusBadRequest, "habit_id is required for this proposal type")
			return
		}
	case model.ProposalKickMember:
		if body.TargetUserID == nil || *body.TargetUserID == "" {
			writeError(w, http.StatusBadRequest, "target_user_id is required for kick_member proposals")
			return
		}
	case model.ProposalDeleteGroup:
		// no extra fields required
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

	proposal, err := db.CreateProposal(r.Context(), h.pool, groupID, userID, body.Type, body.HabitID, body.TargetUserID)
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

	// Quorum: yes_votes * 2 >= member_count (≥50% of all current members).
	if members > 0 && approvals*2 >= members {
		go h.executeProposal(proposal)
		_ = db.SetProposalStatus(r.Context(), h.pool, proposalID, model.ProposalApproved)
	}

	w.WriteHeader(http.StatusNoContent)
}

// executeProposal applies the side-effects of an approved proposal in a goroutine.
// Errors are logged but do not affect the vote response — the vote is already committed.
func (h *ProposalHandler) executeProposal(p *model.Proposal) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch p.Type {
	case model.ProposalAddHabit:
		h.callHabitsService(ctx, p.GroupID, *p.HabitID, "add")

	case model.ProposalRemoveHabit:
		h.callHabitsService(ctx, p.GroupID, *p.HabitID, "remove")

	case model.ProposalKickMember:
		if p.TargetUserID == nil {
			return
		}
		db.RemoveMember(ctx, h.pool, p.GroupID, *p.TargetUserID) //nolint:errcheck

	case model.ProposalDeleteGroup:
		h.pool.Exec(ctx, `DELETE FROM groups WHERE id = $1`, p.GroupID) //nolint:errcheck
	}
}

// callHabitsService notifies habits-service to bulk-assign or bulk-remove a habit.
// Runs inside executeProposal's goroutine; a habits-service outage must not
// block or fail the vote response.
func (h *ProposalHandler) callHabitsService(ctx context.Context, groupID, habitID, action string) {
	if h.habitsServiceURL == "" {
		return
	}

	body, err := json.Marshal(map[string]string{
		"group_id": groupID,
		"habit_id": habitID,
		"action":   action,
	})
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		h.habitsServiceURL+"/internal/proposals/apply", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req) //nolint:errcheck
}
