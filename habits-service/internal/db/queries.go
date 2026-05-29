package db

import (
	"context"
	"time"

	"github.com/dydi/habits-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// currentWeekStart returns the Monday of the current ISO week in UTC.
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

func AssignHabit(ctx context.Context, pool *pgxpool.Pool, userID, groupID, habitID string) (*model.UserHabit, error) {
	uh := &model.UserHabit{}
	err := pool.QueryRow(ctx,
		`INSERT INTO user_habits (user_id, group_id, habit_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, group_id, habit_id) DO NOTHING
		 RETURNING id, user_id, group_id, habit_id, created_at`,
		userID, groupID, habitID,
	).Scan(&uh.ID, &uh.UserID, &uh.GroupID, &uh.HabitID, &uh.CreatedAt)
	return uh, err
}

func FindUserHabitID(ctx context.Context, pool *pgxpool.Pool, userID, groupID, habitID string) (string, error) {
	var id string
	err := pool.QueryRow(ctx,
		`SELECT id FROM user_habits WHERE user_id = $1 AND group_id = $2 AND habit_id = $3`,
		userID, groupID, habitID,
	).Scan(&id)
	return id, err
}

// ─── Checkins ────────────────────────────────────────────────────────────────

func HasCheckinToday(ctx context.Context, pool *pgxpool.Pool, userHabitID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM checkins WHERE user_habit_id = $1 AND checked_on = CURRENT_DATE)`,
		userHabitID,
	).Scan(&exists)
	return exists, err
}

func CreateCheckin(ctx context.Context, pool *pgxpool.Pool, userHabitID string, note *string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO checkins (user_habit_id, note) VALUES ($1, $2)`,
		userHabitID, note,
	)
	return err
}

func GetTodayCheckinsByGroup(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.TodayCheckin, error) {
	rows, err := pool.Query(ctx,
		`SELECT
		    u.id, u.display_name,
		    h.id, h.name, h.icon_key, h.color,
		    CASE WHEN c.id IS NOT NULL THEN true ELSE false END,
		    c.note
		 FROM group_members gm
		 JOIN users u ON u.id = gm.user_id
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 JOIN habits h ON h.id = uh.habit_id
		 LEFT JOIN checkins c ON c.user_habit_id = uh.id AND c.checked_on = CURRENT_DATE
		 WHERE gm.group_id = $1
		 ORDER BY u.display_name, h.name`,
		groupID,
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
			&tc.Checked, &tc.Note,
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
	byHabit := map[string]*accum{}
	order := []string{}

	for rows.Next() {
		var uhID, habitID, habitName, groupID string
		var checkedOn *time.Time
		if err := rows.Scan(&uhID, &habitID, &habitName, &groupID, &checkedOn); err != nil {
			return nil, err
		}
		if _, ok := byHabit[uhID]; !ok {
			byHabit[uhID] = &accum{habitID: habitID, habitName: habitName, groupID: groupID}
			order = append(order, uhID)
		}
		if checkedOn != nil {
			byHabit[uhID].dates = append(byHabit[uhID].dates, *checkedOn)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	streaks := make([]model.Streak, 0, len(order))
	for _, uhID := range order {
		a := byHabit[uhID]
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

func GetEligibleMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.EligibleMember, error) {
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT gm.user_id, u.display_name
		 FROM group_members gm
		 JOIN users u ON u.id = gm.user_id
		 JOIN user_habits uh ON uh.user_id = gm.user_id AND uh.group_id = $1
		 WHERE gm.group_id = $1
		   AND DATE_TRUNC('week', CURRENT_DATE) < CURRENT_DATE
		   AND EXISTS (
		       SELECT 1
		       FROM generate_series(
		           DATE_TRUNC('week', CURRENT_DATE)::date,
		           (CURRENT_DATE - INTERVAL '1 day')::date,
		           '1 day'::interval
		       ) AS day(d)
		       WHERE NOT EXISTS (
		           SELECT 1 FROM checkins c
		           WHERE c.user_habit_id = uh.id AND c.checked_on = day.d
		       )
		   )
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

func GetOrCreateRouletteEntry(ctx context.Context, pool *pgxpool.Pool, groupID, debtorID, weekStart string) (*model.RouletteEntry, error) {
	pool.Exec(ctx,
		`INSERT INTO roulette_entries (group_id, debtor_id, week_start)
		 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		groupID, debtorID, weekStart,
	)

	e := &model.RouletteEntry{}
	err := pool.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, status, spun_at, created_at
		 FROM roulette_entries
		 WHERE group_id = $1 AND debtor_id = $2 AND week_start = $3`,
		groupID, debtorID, weekStart,
	).Scan(&e.ID, &e.GroupID, &e.DebtorID, &e.WeekStart, &e.Status, &e.SpunAt, &e.CreatedAt)
	return e, err
}

func MarkEntryCompleted(ctx context.Context, pool *pgxpool.Pool, entryID string) error {
	_, err := pool.Exec(ctx,
		`UPDATE roulette_entries SET status = 'completed', spun_at = NOW() WHERE id = $1`,
		entryID,
	)
	return err
}

func CreateDebt(ctx context.Context, pool *pgxpool.Pool, groupID, debtorID, entryID, text string, emoji *string, weekStart string) (*model.Debt, error) {
	d := &model.Debt{}
	err := pool.QueryRow(ctx,
		`INSERT INTO debts (group_id, debtor_id, entry_id, punishment_text, punishment_emoji, week_start)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, group_id, debtor_id, entry_id, punishment_text, punishment_emoji,
		           week_start, resolved, created_at, resolved_at`,
		groupID, debtorID, entryID, text, emoji, weekStart,
	).Scan(
		&d.ID, &d.GroupID, &d.DebtorID, &d.EntryID,
		&d.PunishmentText, &d.PunishmentEmoji,
		&d.WeekStart, &d.Resolved, &d.CreatedAt, &d.ResolvedAt,
	)
	return d, err
}

func GetPendingDebts(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Debt, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, group_id, debtor_id, entry_id, punishment_text, punishment_emoji,
		        week_start, resolved, created_at, resolved_at
		 FROM debts
		 WHERE group_id = $1 AND resolved = false
		   AND week_start >= CURRENT_DATE - INTERVAL '7 days'
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
			&d.ID, &d.GroupID, &d.DebtorID, &d.EntryID,
			&d.PunishmentText, &d.PunishmentEmoji,
			&d.WeekStart, &d.Resolved, &d.CreatedAt, &d.ResolvedAt,
		); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, rows.Err()
}

func ResolveDebt(ctx context.Context, pool *pgxpool.Pool, debtID string) error {
	_, err := pool.Exec(ctx,
		`UPDATE debts SET resolved = true, resolved_at = NOW() WHERE id = $1`,
		debtID,
	)
	return err
}
