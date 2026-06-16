package model

import "time"

type Habit struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IconKey     string  `json:"icon_key"`
	Color       string  `json:"color"`
}

// UserHabit is one habit assigned to one member in one group.
// Rows are created/deleted by the proposals system.
type UserHabit struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	GroupID       string    `json:"group_id"`
	HabitID       string    `json:"habit_id"`
	ScheduledTime *string   `json:"scheduled_time,omitempty"` // "HH:MM", nullable
	CreatedAt     time.Time `json:"created_at"`
}

// TodayCheckin is returned by GET /checkins/:groupID/today.
// Status is derived server-side: "done" | "pending".
type TodayCheckin struct {
	UserID        string  `json:"user_id"`
	DisplayName   string  `json:"display_name"`
	HabitID       string  `json:"habit_id"`
	HabitName     string  `json:"habit_name"`
	IconKey       string  `json:"icon_key"`
	Color         string  `json:"color"`
	ScheduledTime *string `json:"scheduled_time,omitempty"`
	Status        string  `json:"status"` // "done" | "pending"
	Note          *string `json:"note,omitempty"`
}

// CheckinHistoryDay is one (member, habit, day-they-checked-in) tuple, used to
// build the real 7-day strips in the squad view instead of faking past days.
type CheckinHistoryDay struct {
	UserID    string `json:"user_id"`
	HabitID   string `json:"habit_id"`
	CheckedOn string `json:"checked_on"` // YYYY-MM-DD
}

type Streak struct {
	UserHabitID string `json:"user_habit_id"`
	HabitID     string `json:"habit_id"`
	HabitName   string `json:"habit_name"`
	GroupID     string `json:"group_id"`
	Current     int    `json:"current"`
}

// RouletteEntry is opened for an eligible member (missed habits Mon-Fri).
// SuggestionDeadline: group members can submit suggestions until this time.
// After the deadline the offender can spin with whatever suggestions exist.
// If no suggestions exist at deadline, collective debt is issued instead.
// SpunAt: nil = not yet spun.
type RouletteEntry struct {
	ID                 string     `json:"id"`
	GroupID            string     `json:"group_id"`
	DebtorID           string     `json:"debtor_id"`
	WeekStart          time.Time  `json:"week_start"`
	SuggestionDeadline time.Time  `json:"suggestion_deadline"`
	SpunAt             *time.Time `json:"spun_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// PunishmentSuggestion is submitted by a group member for a specific RouletteEntry.
// One suggestion per member per entry (enforced by DB unique constraint).
type PunishmentSuggestion struct {
	ID               string    `json:"id"`
	RouletteEntryID  string    `json:"roulette_entry_id"`
	GroupID          string    `json:"group_id"`
	SuggesterID      string    `json:"suggester_id"`
	Text             string    `json:"text"`
	Emoji            *string   `json:"emoji,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Debt is created after a spin or as a collective punishment.
//
// Individual (Scope = "individual"):
//   - One debt for the offender.
//   - WinningSuggestionID is set to the selected suggestion.
//
// Collective (Scope = "collective"):
//   - Triggered when suggestion_deadline passes with zero suggestions.
//   - One debt per active member (including the offender).
//   - WinningSuggestionID is nil; PunishmentText is a default message.
//
// Debts auto-expire at week_start + 14 days (expires_at). Status tracks the
// lifecycle: pending → completed / expired / forgiven.
type Debt struct {
	ID                  string     `json:"id"`
	RouletteEntryID     string     `json:"roulette_entry_id"`
	GroupID             string     `json:"group_id"`
	DebtorID            string     `json:"debtor_id"`
	WeekStart           time.Time  `json:"week_start"`
	WinningSuggestionID *string    `json:"winning_suggestion_id,omitempty"`
	PunishmentText      string     `json:"punishment_text"`
	PunishmentEmoji     *string    `json:"punishment_emoji,omitempty"`
	Scope               string     `json:"scope"`  // "individual" | "collective"
	Status              string     `json:"status"` // "pending" | "completed" | "expired" | "forgiven"
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	ExpiresAt           time.Time  `json:"expires_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

type EligibleMember struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}
