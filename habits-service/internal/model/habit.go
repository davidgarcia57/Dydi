package model

import "time"

type Habit struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IconKey     string  `json:"icon_key"`
	Color       string  `json:"color"`
}

type UserHabit struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	GroupID       string    `json:"group_id"`
	HabitID       string    `json:"habit_id"`
	ScheduledTime *string   `json:"scheduled_time,omitempty"` // "06:00", nullable
	CreatedAt     time.Time `json:"created_at"`
}

// TodayCheckin is returned by GET /checkins/:groupID/today.
// Status is derived server-side: "done" | "pending".
// "missed" is only relevant for historical queries (7-day strip).
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

type Streak struct {
	UserHabitID string `json:"user_habit_id"`
	HabitID     string `json:"habit_id"`
	HabitName   string `json:"habit_name"`
	GroupID     string `json:"group_id"`
	Current     int    `json:"current"`
}

type RouletteEntry struct {
	ID        string     `json:"id"`
	GroupID   string     `json:"group_id"`
	DebtorID  string     `json:"debtor_id"`
	WeekStart time.Time  `json:"week_start"`
	Status    string     `json:"status"`
	SpunAt    *time.Time `json:"spun_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Debt.GroupID, DebtorID, WeekStart are read from roulette_entries via JOIN —
// they are not stored in the debts table.
type Debt struct {
	ID              string     `json:"id"`
	EntryID         string     `json:"entry_id"`
	GroupID         string     `json:"group_id"`
	DebtorID        string     `json:"debtor_id"`
	WeekStart       time.Time  `json:"week_start"`
	PunishmentText  string     `json:"punishment_text"`
	PunishmentEmoji *string    `json:"punishment_emoji,omitempty"`
	Resolved        bool       `json:"resolved"`
	CreatedAt       time.Time  `json:"created_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

type EligibleMember struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type Punishment struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Emoji    string `json:"emoji"`
	Category string `json:"category"`
}

type PunishmentSuggestion struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	WeekStart   time.Time `json:"week_start"`
	SuggesterID string    `json:"suggester_id"`
	Text        string    `json:"text"`
	Emoji       *string   `json:"emoji,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
