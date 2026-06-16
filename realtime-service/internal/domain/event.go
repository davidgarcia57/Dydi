package domain

import "time"

const (
	EventCheckin        = "checkin"
	EventStreakUpdate   = "streak_update"
	EventMemberOnline   = "member_online"
	EventMemberOffline  = "member_offline"
	EventRouletteStart  = "roulette_start"
	EventRouletteResult = "roulette_result"
	EventDebtCreated    = "debt_created"
)

type Event struct {
	Type    string      `json:"type"`
	GroupID string      `json:"groupID"`
	UserID  string      `json:"userID"`
	Payload interface{} `json:"payload"`
	// EmittedAt is stamped by the hub when the event is enqueued. The client can
	// compute delivery latency as (receiveTime - emittedAt) for the paper.
	EmittedAt time.Time `json:"emittedAt"`
}
