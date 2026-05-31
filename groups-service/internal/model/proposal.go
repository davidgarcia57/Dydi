package model

import (
	"encoding/json"
	"time"
)

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

type Proposal struct {
	ID          string          `json:"id"`
	GroupID     string          `json:"group_id"`
	ProposerID  string          `json:"proposer_id"`
	Type        ProposalType    `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      ProposalStatus  `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	VoteCount   int             `json:"vote_count"`
	MemberCount int             `json:"member_count"`
}

type ProposalVote struct {
	ProposalID string    `json:"proposal_id"`
	VoterID    string    `json:"voter_id"`
	Approved   bool      `json:"approved"`
	VotedAt    time.Time `json:"voted_at"`
}
