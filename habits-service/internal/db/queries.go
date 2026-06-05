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

func AssignHabit(ctx context.Context, pool *pgxpool.Pool, userID, groupID, habitID string, scheduledTime *string) (*model.HabitAssignment, error) {
	ha := &model.HabitAssignment{}
	err := pool.QueryRow(ctx,
		`INSERT INTO habit_assignments (user_id, group_id, habit_id, scheduled_time)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, group_id, habit_id) DO NOTHING
		 RETURNING id, user_id, group_id, habit_id, scheduled_time::text, created_at`,
		userID, groupID, habitID, scheduledTime,
	).Scan(&ha.ID, &ha.UserID, &ha.GroupID, &ha.HabitID, &ha.ScheduledTime, &ha.CreatedAt)
	return ha, err
}

func FindHabitAssignmentID(ctx context.Context, pool *pgxpool.Pool, userID, groupID, habitID string) (string, error) {
	var id string
	err := pool.QueryRow(ctx,
		`SELECT id FROM habit_assignments WHERE user_id = $1 AND group_id = $2 AND habit_id = $3`,
		userID, groupID, habitID,
	).Scan(&id)
	return id, err
}

// ─── Checkins ────────────────────────────────────────────────────────────────

// HasCheckinOnDate checks if a check-in exists for the given date ("YYYY-MM-DD").
// The date is provided by the caller (derived from the client's local timezone)
// to avoid CURRENT_DATE UTC drift for users in non-UTC timezones.
func HasCheckinOnDate(ctx context.Context, pool *pgxpool.Pool, habitAssignmentID, date string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM checkins WHERE habit_assignment_id = $1 AND checked_on = $2::date)`,
		habitAssignmentID, date,
	).Scan(&exists)
	return exists, err
}

func CreateCheckin(ctx context.Context, pool *pgxpool.Pool, habitAssignmentID string, checkedOn string, note *string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO checkins (habit_assignment_id, checked_on, note) VALUES ($1, $2, $3)`,
		habitAssignmentID, checkedOn, note,
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
		    ha.scheduled_time::text,
		    CASE WHEN c.id IS NOT NULL THEN 'done' ELSE 'pending' END AS status,
		    c.note
		 FROM user_groups gm
		 JOIN users u           ON u.id  = gm.user_id
		 JOIN habit_assignments ha ON ha.user_id = gm.user_id AND ha.group_id = $1
		 JOIN habits h          ON h.id  = ha.habit_id
		 LEFT JOIN checkins c   ON c.habit_assignment_id = ha.id AND c.checked_on = $2::date
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
			&tc.ScheduledTime,
			&tc.Status,
			&tc.Note,
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
		`SELECT ha.id, ha.habit_id, h.name, ha.group_id, c.checked_on
		 FROM habit_assignments ha
		 JOIN habits h ON h.id = ha.habit_id
		 LEFT JOIN checkins c ON c.habit_assignment_id = ha.id
		 WHERE ha.user_id = $1
		 ORDER BY ha.id, c.checked_on DESC NULLS LAST`,
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
	byAssignment := map[string]*accum{}
	order := []string{}

	for rows.Next() {
		var haID, habitID, habitName, groupID string
		var checkedOn *time.Time
		if err := rows.Scan(&haID, &habitID, &habitName, &groupID, &checkedOn); err != nil {
			return nil, err
		}
		if _, ok := byAssignment[haID]; !ok {
			byAssignment[haID] = &accum{habitID: habitID, habitName: habitName, groupID: groupID}
			order = append(order, haID)
		}
		if checkedOn != nil {
			byAssignment[haID].dates = append(byAssignment[haID].dates, *checkedOn)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	streaks := make([]model.Streak, 0, len(order))
	for _, haID := range order {
		a := byAssignment[haID]
		streaks = append(streaks, model.Streak{
			HabitAssignmentID: haID,
			HabitID:           a.habitID,
			HabitName:         a.habitName,
			GroupID:           a.groupID,
			Current:           calculateStreak(a.dates),
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
		`SELECT EXISTS(SELECT 1 FROM user_groups WHERE group_id = $1 AND user_id = $2)`,
		groupID, userID,
	).Scan(&exists)
	return exists, err
}

// GetEligibleMembers returns members who missed at least one habit on any day
// this week before today. On Monday the list is always empty.
func GetEligibleMembers(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.EligibleMember, error) {
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT gm.user_id, u.display_name
		 FROM user_groups gm
		 JOIN users u              ON u.id = gm.user_id
		 JOIN habit_assignments ha ON ha.user_id = gm.user_id AND ha.group_id = $1
		 WHERE gm.group_id = $1
		   AND (CURRENT_DATE - DATE_TRUNC('week', CURRENT_DATE)::date) > 0
		   AND (
		       SELECT COUNT(*)
		       FROM checkins c
		       WHERE c.habit_assignment_id = ha.id
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

// GetOrCreateRouletteDraw accepts DBTX so it can run inside a transaction.
// Uses spun_at IS NOT NULL to determine if already spun (no status column).
func GetOrCreateRouletteDraw(ctx context.Context, dbtx DBTX, groupID, debtorID, weekStart string) (*model.RouletteDraw, error) {
	dbtx.Exec(ctx,
		`INSERT INTO roulette_draws (group_id, debtor_id, week_start)
		 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		groupID, debtorID, weekStart,
	)

	d := &model.RouletteDraw{}
	err := dbtx.QueryRow(ctx,
		`SELECT id, group_id, debtor_id, week_start, spun_at, created_at
		 FROM roulette_draws
		 WHERE group_id = $1 AND debtor_id = $2 AND week_start = $3`,
		groupID, debtorID, weekStart,
	).Scan(&d.ID, &d.GroupID, &d.DebtorID, &d.WeekStart, &d.SpunAt, &d.CreatedAt)
	return d, err
}

// MarkDrawCompleted accepts DBTX so it can run inside a transaction.
func MarkDrawCompleted(ctx context.Context, dbtx DBTX, drawID string) error {
	_, err := dbtx.Exec(ctx,
		`UPDATE roulette_draws SET spun_at = NOW() WHERE id = $1`,
		drawID,
	)
	return err
}

// CreateDebt inserts a debt and returns it enriched with group/debtor/week
// from roulette_draws via JOIN. Accepts DBTX so it can run inside a transaction.
func CreateDebt(ctx context.Context, dbtx DBTX, drawID, text string, emoji *string) (*model.Debt, error) {
	d := &model.Debt{}
	err := dbtx.QueryRow(ctx,
		`WITH ins AS (
		     INSERT INTO debts (draw_id, punishment_text, punishment_emoji)
		     VALUES ($1, $2, $3)
		     RETURNING *
		 )
		 SELECT ins.id, ins.draw_id,
		        ins.punishment_text, ins.punishment_emoji,
		        ins.resolved, ins.created_at, ins.resolved_at,
		        rd.group_id, rd.debtor_id, rd.week_start
		 FROM ins
		 JOIN roulette_draws rd ON rd.id = ins.draw_id`,
		drawID, text, emoji,
	).Scan(
		&d.ID, &d.DrawID,
		&d.PunishmentText, &d.PunishmentEmoji,
		&d.Resolved, &d.CreatedAt, &d.ResolvedAt,
		&d.GroupID, &d.DebtorID, &d.WeekStart,
	)
	return d, err
}

func GetPendingDebts(ctx context.Context, pool *pgxpool.Pool, groupID string) ([]model.Debt, error) {
	rows, err := pool.Query(ctx,
		`SELECT d.id, d.draw_id,
		        d.punishment_text, d.punishment_emoji,
		        d.resolved, d.created_at, d.resolved_at,
		        rd.group_id, rd.debtor_id, rd.week_start
		 FROM debts d
		 JOIN roulette_draws rd ON rd.id = d.draw_id
		 WHERE rd.group_id = $1 AND d.resolved = false
		 ORDER BY d.created_at DESC`,
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
			&d.ID, &d.DrawID,
			&d.PunishmentText, &d.PunishmentEmoji,
			&d.Resolved, &d.CreatedAt, &d.ResolvedAt,
			&d.GroupID, &d.DebtorID, &d.WeekStart,
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
