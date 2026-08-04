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
	EventDebtUpdated    = "debt_updated"
	// EventHabitsChanged se emite cuando una propuesta aprobada ya se aplicó de
	// verdad (habits-service asignó o quitó el hábito). Es la única señal fiable
	// de que terminó: el voto responde antes de que la asignación ocurra.
	EventHabitsChanged = "habits_changed"
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
