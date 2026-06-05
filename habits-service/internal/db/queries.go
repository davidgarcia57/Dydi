package db

import (
	"context"
	"time"

	"github.com/dydi/habits-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX allows passing either *pgxpool.Pool or pgx.Tx to query functions,
// enabling transactional and non-transactional callers to share the same code.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// CurrentWeekStart returns the Monday of the current ISO week in UTC.
func CurrentWeekStart() time.Time {
	now := time.Now().UTC()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
}

// ─── Habits ──────────────────────────────────────────────────────────────────

func ListHabits(ctx context.Context, pool *pgxpool.Pool) ([]model.Habit, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, name, description, icon_key, color FROM habits ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	habits := make([]model.Habit, 0)
	for rows.Next() {
		var h model.Habit
		if err := rows.Scan(&h.ID, &h.Name, &h.Description, &h.IconKey, &h.Color); err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}
	return habits, rows.Err()
}

// FindUserHabitID returns the user_habits.id for a user/group/habit triple.
func FindUserHabitID(ctx context.Context, pool *pgxpool.Pool, userID, groupID, habitID string) (string, error) {
	var id string
	err := pool.QueryRow(ctx,
		`SELECT id FROM user_habits WHERE user_id = $1 AND group_id = $2 AND habit_id = $3`,
		userID, groupID, habitID,
	).Scan(&id)
	return id, err
}

// BulkAssignHabit assigns a habit to every current member of a group.
// Called when an add_habit proposal is approved.
func BulkAssignHabit(ctx context.Context, pool *pgxpool.Pool, groupID, habitID string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO user_habits (user_id, group_id, habit_id)
		 SELECT user_id, $1, $2
		 FROM group_members
		 WHERE group_id = $1
		 ON CONFLICT (user_id, group_id, habit_id) DO NOTHING`,
		groupID, habitID,
	)
	return err
}

// BulkUnassignHabit removes a habit from every member of a group.
// Called when a remove_habit proposal is approved.
func BulkUnassignHabit(ctx context.Context, pool *pgxpool.Pool, groupID, habitID string) error {
	_, err := pool.Exec(ctx,
		`DELETE FROM user_habits WHERE group_id = $1 AND habit_id = $2`,
		groupID, habitID,
	)
	return err
}

// ─── Checkins ────────────────────────────────────────────────────────────────

// HasCheckinOnDate returns true if a check-in exists for the given user_habit and date.
// The date is provided by the caller (client's local date) to avoid UTC drift.
func HasCheckinOnDate(ctx context.Context, pool *pgxpool.Pool, userHabitID, date string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM checkins WHERE user_habit_id = $1 AND checked_on = $2::date)`,
		userHabitID, date,
	).Scan(&exists)
	return exists, err
}

func CreateCheckin(ctx context.Context, pool *pgxpool.Pool, userHabitID, checkedOn string, note *string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO checkins (user_habit_id, checked_on, note) VALUES ($1, $2, $3)`,
		userHabitID, checkedOn, note,
	)
	return err
}

func GetTodayCheckinsByGroup(ctx context.Context, pool *pgxpool.Pool, groupID, date string) ([]model.TodayCheckin, error) {
	rows, err := pool.Query(ctx,
		`SELECT
		     u.id,
		     u.display_name,
		     h.id,
		     h.name,
		     h.icon_key,
		     h.color,
		     uh.scheduled_time::text,
		     CASE WHEN c.id IS NOT NULL THEN 'done' ELSE 'pending' END AS status,
		     c.note
		 FROM group_members gm
		 JOIN users       u  ON u.id = gm.user_id
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 JOIN habits      h  ON h.id = uh.habit_id
		 LEFT JOIN checkins c ON c.user_habit_id = uh.id AND c.checked_on = $2::date
		 WHERE gm.group_id = $1
		 ORDER BY u.display_name, h.name`,
		groupID, date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.TodayCheckin, 0)
	for rows.Next() {
		var tc model.TodayCheckin
		if err := rows.Scan(
			&tc.UserID, &tc.DisplayName,
			&tc.HabitID, &tc.HabitName, &tc.IconKey, &tc.Color,
			&tc.ScheduledTime, &tc.Status, &tc.Note,
		); err != nil {
			return nil, err
		}
		result = append(result, tc)
	}
	return result, rows.Err()
}

// ─── Streaks ─────────────────────────────────────────────────────────────────

func GetStreaksForUser(ctx context.Context, pool *pgxpool.Pool, userID string) ([]model.Streak, error) {
	rows, err := pool.Query(ctx,
		`SELECT uh.id, uh.habit_id, h.name, uh.group_id, c.checked_on
		 FROM user_habits uh
		 JOIN habits h ON h.id = uh.habit_id
		 LEFT JOIN checkins c ON c.user_habit_id = uh.id
		 WHERE uh.user_id = $1
		 ORDER BY uh.id, c.checked_on DESC NULLS LAST`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type accum struct {
		habitID   string
		habitName string
		groupID   string
		dates     []time.Time
	}
	byUserHabit := map[string]*accum{}
	order := []string{}

	for rows.Next() {
		var uhID, habitID, habitName, groupID string
		var checkedOn *time.Time
		if err := rows.Scan(&uhID, &habitID, &habitName, &groupID, &checkedOn); err != nil {
			return nil, err
		}
		if _, ok := byUserHabit[uhID]; !ok {
			byUserHabit[uhID] = &accum{habitID: habitID, habitName: habitName, groupID: groupID}
			order = append(order, uhID)
		}
		if checkedOn != nil {
			byUserHabit[uhID].dates = append(byUserHabit[uhID].dates, *checkedOn)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	streaks := make([]model.Streak, 0, len(order))
	for _, uhID := range order {
		a := byUserHabit[uhID]
		streaks = append(streaks, model.Streak{
			UserHabitID: uhID,
			HabitID:     a.habitID,
			HabitName:   a.habitName,
			GroupID:     a.groupID,
			Current:     calculateStreak(a.dates),
		})
	}
	return streaks, nil
}

// calculateStreak counts consecutive days ending on today or yesterday.
// dates must be sorted DESC (most recent first).
func calculateStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	latest := time.Date(dates[0].Year(), dates[0].Month(), dates[0].Day(), 0, 0, 0, 0, time.UTC)
	if !latest.Equal(today) && !latest.Equal(yesterday) {
		return 0
	}

	streak := 1
	prev := latest
	for i := 1; i < len(dates); i++ {
		curr := time.Date(dates[i].Year(), dates[i].Month(), dates[i].Day(), 0, 0, 0, 0, time.UTC)
		if curr.Equal(prev.AddDate(0, 0, -1)) {
			streak++
			prev = curr
		} else {
			break
		}
	}
	return streak
}

// ─── Penalties ───────────────────────────────────────────────────────────────

func IsMemberOfGroup(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)`,
		groupID, userID,
	).Scan(&exists)
	return exists, err
}

// GetEligibleMembers returns members who missed at least one habit Mon–yesterday.
// On Monday the list is always empty.
func GetEligibleMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.EligibleMember, error) {
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT gm.user_id, u.display_name
		 FROM group_members gm
		 JOIN users       u  ON u.id = gm.user_id
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 WHERE gm.group_id = $1
		   AND (CURRENT_DATE - DATE_TRUNC('week', CURRENT_DATE)::date) > 0
		   AND (
		       SELECT COUNT(*)
		       FROM checkins c
		       WHERE c.user_habit_id = uh.id
		         AND c.checked_on >= DATE_TRUNC('week', CURRENT_DATE)::date
		         AND c.checked_on < CURRENT_DATE
		   ) < (CURRENT_DATE - DATE_TRUNC('week', CURRENT_DATE)::date)
		 ORDER BY u.display_name`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]model.EligibleMember, 0)
	for rows.Next() {
		var m model.EligibleMember
		if err := rows.Scan(&m.UserID, &m.DisplayName); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// ─── Roulette entries ─────────────────────────────────────────────────────────

func scanEntry(row pgx.Row) (*model.RouletteEntry, error) {
	e := &model.RouletteEntry{}
	err := row.Scan(
		&e.ID, &e.GroupID, &e.DebtorID, &e.WeekStart,
		&e.SuggestionDeadline, &e.SpunAt, &e.CreatedAt,
	)
	return e, err
}

// GetOrCreateRouletteEntry inserts a roulette_entry if it does not exist yet,
// then returns it. deadline sets when the suggestion window closes.
// Accepts DBTX so it can run inside a transaction.
func GetOrCreateRouletteEntry(ctx context.Context, dbtx DBTX, groupID, debtorID, weekStart string, deadline time.Time) (*model.RouletteEntry, error) {
	dbtx.Exec(ctx,
		`INSERT INTO roulette_entries (group_id, debtor_id, week_start, suggestion_deadline)
		 VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
		groupID, debtorID, weekStart, deadline,
	)
	return scanEntry(dbtx.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, suggestion_deadline, spun_at, created_at
		 FROM roulette_entries
		 WHERE group_id = $1 AND debtor_id = $2 AND week_start = $3`,
		groupID, debtorID, weekStart,
	))
}

func GetRouletteEntry(ctx context.Context, pool *pgxpool.Pool, entryID string) (*model.RouletteEntry, error) {
	return scanEntry(pool.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, suggestion_deadline, spun_at, created_at
		 FROM roulette_entries WHERE id = $1`,
		entryID,
	))
}

func MarkEntrySpun(ctx context.Context, dbtx DBTX, entryID string) error {
	_, err := dbtx.Exec(ctx,
		`UPDATE roulette_entries SET spun_at = NOW() WHERE id = $1`,
		entryID,
	)
	return err
}

// ─── Suggestions ─────────────────────────────────────────────────────────────

func HasSuggested(ctx context.Context, pool *pgxpool.Pool, entryID, suggesterID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(
		     SELECT 1 FROM punishment_suggestions
		     WHERE roulette_entry_id = $1 AND suggester_id = $2
		 )`,
		entryID, suggesterID,
	).Scan(&exists)
	return exists, err
}

func CreateSuggestion(ctx context.Context, pool *pgxpool.Pool, entryID, groupID, suggesterID, text string, emoji *string) (*model.PunishmentSuggestion, error) {
	s := &model.PunishmentSuggestion{}
	err := pool.QueryRow(ctx,
		`INSERT INTO punishment_suggestions (roulette_entry_id, group_id, suggester_id, text, emoji)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, roulette_entry_id, group_id, suggester_id, text, emoji, created_at`,
		entryID, groupID, suggesterID, text, emoji,
	).Scan(&s.ID, &s.RouletteEntryID, &s.GroupID, &s.SuggesterID, &s.Text, &s.Emoji, &s.CreatedAt)
	return s, err
}

func GetSuggestionsForEntry(ctx context.Context, pool *pgxpool.Pool, entryID string) ([]model.PunishmentSuggestion, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, roulette_entry_id, group_id, suggester_id, text, emoji, created_at
		 FROM punishment_suggestions
		 WHERE roulette_entry_id = $1
		 ORDER BY created_at`,
		entryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suggestions := make([]model.PunishmentSuggestion, 0)
	for rows.Next() {
		var s model.PunishmentSuggestion
		if err := rows.Scan(&s.ID, &s.RouletteEntryID, &s.GroupID, &s.SuggesterID, &s.Text, &s.Emoji, &s.CreatedAt); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, s)
	}
	return suggestions, rows.Err()
}

// CountSuggestionsAndMembers returns how many suggestions exist for an entry
// and how many members the group currently has.
func CountSuggestionsAndMembers(ctx context.Context, pool *pgxpool.Pool, entryID string) (suggestions, members int, err error) {
	err = pool.QueryRow(ctx,
		`SELECT
		     (SELECT COUNT(*) FROM punishment_suggestions  WHERE roulette_entry_id = $1),
		     (SELECT COUNT(*) FROM group_members
		      WHERE group_id = (SELECT group_id FROM roulette_entries WHERE id = $1))`,
		entryID,
	).Scan(&suggestions, &members)
	return
}

// ─── Debts ───────────────────────────────────────────────────────────────────

func scanDebt(row pgx.Row) (*model.Debt, error) {
	d := &model.Debt{}
	err := row.Scan(
		&d.ID, &d.RouletteEntryID, &d.GroupID, &d.DebtorID, &d.WeekStart,
		&d.WinningSuggestionID, &d.PunishmentText, &d.PunishmentEmoji,
		&d.IsCollective, &d.ExpiresAt, &d.CreatedAt,
	)
	return d, err
}

// CreateDebt records the result of a spin.
// suggestionID is the winning suggestion; text and emoji are snapshotted from it.
// Accepts DBTX so it can run inside a transaction.
func CreateDebt(ctx context.Context, dbtx DBTX, entryID, groupID, debtorID, weekStart string, suggestionID *string, text string, emoji *string) (*model.Debt, error) {
	return scanDebt(dbtx.QueryRow(ctx,
		`INSERT INTO debts
		     (roulette_entry_id, group_id, debtor_id, week_start,
		      winning_suggestion_id, punishment_text, punishment_emoji,
		      is_collective, expires_at)
		 VALUES ($1, $2, $3, $4::date, $5, $6, $7, false, $4::date + INTERVAL '14 days')
		 RETURNING id, roulette_entry_id, group_id, debtor_id, week_start,
		           winning_suggestion_id, punishment_text, punishment_emoji,
		           is_collective, expires_at, created_at`,
		entryID, groupID, debtorID, weekStart, suggestionID, text, emoji,
	))
}

// CreateCollectiveDebts creates one debt per group member when no suggestions
// were submitted before the deadline. Accepts DBTX for transactional safety.
func CreateCollectiveDebts(ctx context.Context, dbtx DBTX, entryID, groupID, weekStart string) ([]model.Debt, error) {
	rows, err := dbtx.Query(ctx,
		`INSERT INTO debts
		     (roulette_entry_id, group_id, debtor_id, week_start,
		      winning_suggestion_id, punishment_text, is_collective, expires_at)
		 SELECT $1, $2, user_id, $3::date,
		        NULL, 'Nadie propuso una penitencia — todos pagan.', true,
		        $3::date + INTERVAL '14 days'
		 FROM group_members
		 WHERE group_id = $2
		 ON CONFLICT (roulette_entry_id, debtor_id) DO NOTHING
		 RETURNING id, roulette_entry_id, group_id, debtor_id, week_start,
		           winning_suggestion_id, punishment_text, punishment_emoji,
		           is_collective, expires_at, created_at`,
		entryID, groupID, weekStart,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	debts := make([]model.Debt, 0)
	for rows.Next() {
		var d model.Debt
		if err := rows.Scan(
			&d.ID, &d.RouletteEntryID, &d.GroupID, &d.DebtorID, &d.WeekStart,
			&d.WinningSuggestionID, &d.PunishmentText, &d.PunishmentEmoji,
			&d.IsCollective, &d.ExpiresAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, rows.Err()
}

// GetActiveDebts returns all non-expired debts for a group.
func GetActiveDebts(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Debt, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, roulette_entry_id, group_id, debtor_id, week_start,
		        winning_suggestion_id, punishment_text, punishment_emoji,
		        is_collective, expires_at, created_at
		 FROM debts
		 WHERE group_id = $1 AND expires_at > CURRENT_DATE
		 ORDER BY created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	debts := make([]model.Debt, 0)
	for rows.Next() {
		var d model.Debt
		if err := rows.Scan(
			&d.ID, &d.RouletteEntryID, &d.GroupID, &d.DebtorID, &d.WeekStart,
			&d.WinningSuggestionID, &d.PunishmentText, &d.PunishmentEmoji,
			&d.IsCollective, &d.ExpiresAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, rows.Err()
}
