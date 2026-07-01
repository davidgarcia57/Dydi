package db

import (
	"context"
	"errors"
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

// BulkAssignHabit makes a group adopt a habit and assigns it to every active
// member. Called when an add_habit proposal is approved. Runs in a transaction
// because user_habits has a composite FK to group_habits — the group must adopt
// the habit before any member can hold it. addedBy is the proposer.
func BulkAssignHabit(ctx context.Context, pool *pgxpool.Pool, groupID, habitID, addedBy string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// The group adopts the habit (un-archive it if it had been removed before).
	if _, err := tx.Exec(ctx,
		`INSERT INTO group_habits (group_id, habit_id, added_by)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (group_id, habit_id) DO UPDATE SET archived_at = NULL`,
		groupID, habitID, addedBy,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO user_habits (user_id, group_id, habit_id)
		 SELECT user_id, $1, $2
		 FROM memberships
		 WHERE group_id = $1 AND status = 'active'
		 ON CONFLICT (group_id, user_id, habit_id) DO NOTHING`,
		groupID, habitID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// BulkUnassignHabit removes a habit from a group. Deleting the group_habit row
// cascades to every member's user_habits (and their checkins) via the FK.
// Called when a remove_habit proposal is approved.
func BulkUnassignHabit(ctx context.Context, pool *pgxpool.Pool, groupID, habitID string) error {
	_, err := pool.Exec(ctx,
		`DELETE FROM group_habits WHERE group_id = $1 AND habit_id = $2`,
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
		 FROM memberships gm
		 JOIN users       u  ON u.id = gm.user_id
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 JOIN habits      h  ON h.id = uh.habit_id
		 LEFT JOIN checkins c ON c.user_habit_id = uh.id AND c.checked_on = $2::date
		 WHERE gm.group_id = $1 AND gm.status = 'active'
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

// GetCheckinHistory returns every (member, habit, date) check-in for a group
// within [from, to]. The frontend groups these into per-member 7-day strips.
func GetCheckinHistory(ctx context.Context, pool *pgxpool.Pool, groupID, from, to string) ([]model.CheckinHistoryDay, error) {
	rows, err := pool.Query(ctx,
		`SELECT uh.user_id, uh.habit_id, to_char(c.checked_on, 'YYYY-MM-DD'), c.note
		 FROM memberships gm
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 JOIN checkins   c  ON c.user_habit_id = uh.id
		 WHERE gm.group_id = $1 AND gm.status = 'active'
		   AND c.checked_on >= $2::date AND c.checked_on <= $3::date`,
		groupID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]model.CheckinHistoryDay, 0)
	for rows.Next() {
		var d model.CheckinHistoryDay
		if err := rows.Scan(&d.UserID, &d.HabitID, &d.CheckedOn, &d.Note); err != nil {
			return nil, err
		}
		history = append(history, d)
	}
	return history, rows.Err()
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
		`SELECT EXISTS(SELECT 1 FROM memberships
		               WHERE group_id = $1 AND user_id = $2 AND status = 'active')`,
		groupID, userID,
	).Scan(&exists)
	return exists, err
}

// UsersShareGroup reports whether two users are active members of at least one
// common group. Used to authorize reading another user's streaks.
func UsersShareGroup(ctx context.Context, pool *pgxpool.Pool, userA, userB string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(
		     SELECT 1
		     FROM memberships m1
		     JOIN memberships m2 ON m1.group_id = m2.group_id
		     WHERE m1.user_id = $1 AND m2.user_id = $2
		       AND m1.status = 'active' AND m2.status = 'active'
		 )`,
		userA, userB,
	).Scan(&exists)
	return exists, err
}

// IsEligibleForRoulette reports whether a debtor missed at least one habit-day
// so far this week (Mon→yesterday) — the same rule GetEligibleMembers applies,
// scoped to a single user. A roulette must not be opened against someone who
// didn't actually fail. On Monday nobody is eligible yet.
func IsEligibleForRoulette(ctx context.Context, pool *pgxpool.Pool, groupID, debtorID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(
		     SELECT 1
		     FROM memberships gm
		     JOIN users       u  ON u.id = gm.user_id
		     JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		     -- "today" and the week's Monday in the member's OWN timezone, so a
		     -- late-night check-in counts for the right day instead of being
		     -- bumped a day forward in UTC.
		     CROSS JOIN LATERAL (
		         SELECT date_trunc('week', (now() AT TIME ZONE u.timezone))::date AS week_start,
		                (now() AT TIME ZONE u.timezone)::date                     AS today
		     ) d
		     WHERE gm.group_id = $1 AND gm.user_id = $2 AND gm.status = 'active'
		       AND (d.today - d.week_start) > 0
		       AND (
		           SELECT COUNT(*)
		           FROM checkins c
		           WHERE c.user_habit_id = uh.id
		             AND c.checked_on >= d.week_start
		             AND c.checked_on <  d.today
		       ) < (d.today - d.week_start)
		 )`,
		groupID, debtorID,
	).Scan(&exists)
	return exists, err
}

// GetEligibleMembers returns members who missed at least one habit Mon–yesterday.
// On Monday the list is always empty.
func GetEligibleMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.EligibleMember, error) {
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT gm.user_id, u.display_name
		 FROM memberships gm
		 JOIN users       u  ON u.id = gm.user_id
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 -- "today" and the week's Monday in each member's OWN timezone, so a
		 -- late-night check-in is not bumped a day forward in UTC.
		 CROSS JOIN LATERAL (
		     SELECT date_trunc('week', (now() AT TIME ZONE u.timezone))::date AS week_start,
		            (now() AT TIME ZONE u.timezone)::date                     AS today
		 ) d
		 WHERE gm.group_id = $1 AND gm.status = 'active'
		   AND (d.today - d.week_start) > 0
		   AND (
		       SELECT COUNT(*)
		       FROM checkins c
		       WHERE c.user_habit_id = uh.id
		         AND c.checked_on >= d.week_start
		         AND c.checked_on <  d.today
		   ) < (d.today - d.week_start)
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
// then returns it plus whether this call created it (so the caller can
// broadcast roulette_start exactly once). deadline sets when the suggestion
// window closes. Accepts DBTX so it can run inside a transaction.
func GetOrCreateRouletteEntry(ctx context.Context, dbtx DBTX, groupID, debtorID, weekStart string, deadline time.Time) (*model.RouletteEntry, bool, error) {
	entry, err := scanEntry(dbtx.QueryRow(ctx,
		`INSERT INTO roulette_entries (group_id, debtor_id, week_start, suggestion_deadline)
		 VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING
		 RETURNING id, group_id, debtor_id, week_start, suggestion_deadline, spun_at, created_at`,
		groupID, debtorID, weekStart, deadline,
	))
	if err == nil {
		return entry, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}
	entry, err = scanEntry(dbtx.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, suggestion_deadline, spun_at, created_at
		 FROM roulette_entries
		 WHERE group_id = $1 AND debtor_id = $2 AND week_start = $3`,
		groupID, debtorID, weekStart,
	))
	return entry, false, err
}

// GetOpenRouletteEntries returns the group's unspun roulette entries (newest
// first) with the debtor's display name, so clients can list open roulettes
// even after the debtor's current-week eligibility has expired.
func GetOpenRouletteEntries(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.OpenRouletteEntry, error) {
	rows, err := pool.Query(ctx,
		`SELECT e.id, e.group_id, e.debtor_id, e.week_start, e.suggestion_deadline,
		        e.spun_at, e.created_at, u.display_name
		 FROM roulette_entries e
		 JOIN users u ON u.id = e.debtor_id
		 WHERE e.group_id = $1 AND e.spun_at IS NULL
		 ORDER BY e.created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]model.OpenRouletteEntry, 0)
	for rows.Next() {
		var e model.OpenRouletteEntry
		if err := rows.Scan(
			&e.ID, &e.GroupID, &e.DebtorID, &e.WeekStart,
			&e.SuggestionDeadline, &e.SpunAt, &e.CreatedAt, &e.DebtorName,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func GetRouletteEntry(ctx context.Context, pool *pgxpool.Pool, entryID string) (*model.RouletteEntry, error) {
	return scanEntry(pool.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, suggestion_deadline, spun_at, created_at
		 FROM roulette_entries WHERE id = $1`,
		entryID,
	))
}

// GetRouletteEntryForUpdate locks the entry row (SELECT ... FOR UPDATE) inside a
// transaction. A concurrent second spin blocks here until the first commits,
// then sees spun_at set — preventing a double spin / double debt under a race.
func GetRouletteEntryForUpdate(ctx context.Context, dbtx DBTX, entryID string) (*model.RouletteEntry, error) {
	return scanEntry(dbtx.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, suggestion_deadline, spun_at, created_at
		 FROM roulette_entries WHERE id = $1 FOR UPDATE`,
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
		     (SELECT COUNT(*) FROM memberships
		      WHERE group_id = (SELECT group_id FROM roulette_entries WHERE id = $1)
		        AND status = 'active')`,
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
		&d.Scope, &d.Status, &d.CompletedAt, &d.ExpiresAt, &d.CreatedAt,
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
		      scope, expires_at)
		 VALUES ($1, $2, $3, $4::date, $5, $6, $7, 'individual', $4::date + INTERVAL '14 days')
		 RETURNING id, roulette_entry_id, group_id, debtor_id, week_start,
		           winning_suggestion_id, punishment_text, punishment_emoji,
		           scope, status, completed_at, expires_at, created_at`,
		entryID, groupID, debtorID, weekStart, suggestionID, text, emoji,
	))
}

// CompleteDebt marks the debtor's own pending debt as completed. Returns
// pgx.ErrNoRows when the debt does not exist, belongs to someone else or was
// already resolved.
func CompleteDebt(ctx context.Context, pool *pgxpool.Pool, debtID, debtorID string) (*model.Debt, error) {
	return scanDebt(pool.QueryRow(ctx,
		`UPDATE debts
		 SET status = 'completed', completed_at = NOW()
		 WHERE id = $1 AND debtor_id = $2 AND status = 'pending'
		 RETURNING id, roulette_entry_id, group_id, debtor_id, week_start,
		           winning_suggestion_id, punishment_text, punishment_emoji,
		           scope, status, completed_at, expires_at, created_at`,
		debtID, debtorID,
	))
}

// GetActiveDebts returns all non-expired debts for a group.
func GetActiveDebts(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Debt, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, roulette_entry_id, group_id, debtor_id, week_start,
		        winning_suggestion_id, punishment_text, punishment_emoji,
		        scope, status, completed_at, expires_at, created_at
		 FROM debts
		 WHERE group_id = $1 AND status = 'pending' AND expires_at > CURRENT_DATE
		 ORDER BY created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectDebts(rows)
}

// GetResolvedDebts returns the group's settled debts (newest first, capped at
// 50): completed, forgiven, and expired ones. A 'pending' debt past expires_at
// is presented as 'expired' — nothing flips the column, it just lapses.
func GetResolvedDebts(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Debt, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, roulette_entry_id, group_id, debtor_id, week_start,
		        winning_suggestion_id, punishment_text, punishment_emoji,
		        scope,
		        CASE WHEN status = 'pending' AND expires_at <= CURRENT_DATE THEN 'expired'
		             ELSE status END AS status,
		        completed_at, expires_at, created_at
		 FROM debts
		 WHERE group_id = $1 AND (status <> 'pending' OR expires_at <= CURRENT_DATE)
		 ORDER BY created_at DESC
		 LIMIT 50`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectDebts(rows)
}

func collectDebts(rows pgx.Rows) ([]model.Debt, error) {
	debts := make([]model.Debt, 0)
	for rows.Next() {
		var d model.Debt
		if err := rows.Scan(
			&d.ID, &d.RouletteEntryID, &d.GroupID, &d.DebtorID, &d.WeekStart,
			&d.WinningSuggestionID, &d.PunishmentText, &d.PunishmentEmoji,
			&d.Scope, &d.Status, &d.CompletedAt, &d.ExpiresAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, rows.Err()
}
