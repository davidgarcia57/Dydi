package model

import "time"

type ProposalType string

const (
	ProposalAddHabit    ProposalType = "add_habit"
	ProposalRemoveHabit ProposalType = "remove_habit"
	ProposalKickMember  ProposalType = "kick_member"
	ProposalDeleteGroup ProposalType = "delete_group"
)

type ProposalStatus string

const (
	ProposalOpen     ProposalStatus = "open"
	ProposalApproved ProposalStatus = "approved"
	ProposalRejected ProposalStatus = "rejected"
	ProposalExpired  ProposalStatus = "expired"
)

// Proposal uses typed columns instead of a generic JSONB payload.
// Only the fields relevant to each type will be non-nil:
//   add_habit / remove_habit → HabitID is set
//   kick_member              → TargetUserID is set
//   delete_group             → neither is set
//
// VoteCount and MemberCount are computed at query time, not stored.
type Proposal struct {
	ID            string         `json:"id"`
	GroupID       string         `json:"group_id"`
	ProposerID    string         `json:"proposer_id"`
	Type          ProposalType   `json:"type"`
	HabitID       *string        `json:"habit_id,omitempty"`
	TargetUserID  *string        `json:"target_user_id,omitempty"`
	Status        ProposalStatus `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	VoteCount     int            `json:"vote_count"`
	MemberCount   int            `json:"member_count"`
}

type ProposalVote struct {
	ProposalID string    `json:"proposal_id"`
	VoterID    string    `json:"voter_id"`
	Approved   bool      `json:"approved"`
	VotedAt    time.Time `json:"voted_at"`
}
