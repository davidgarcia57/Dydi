//go:build integration

// Tests de integración de habits-service contra un Postgres REAL.
//
// Por qué existen: el resto de la suite pasa `nil` como pool, así que solo
// ejercita las guardas que retornan antes de tocar la BD. Nada del SQL se
// ejecuta en CI, y por eso un UPDATE con tipos inválidos podía vivir para
// siempre en verde. Aquí se corre el SQL de verdad.
//
// Se saltan solos si TEST_DATABASE_URL no está: `verify.sh` y el CI siguen
// pasando sin BD. Para correrlos: ./scripts/test-db.sh
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dydi/habits-service/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool abre el pool contra la BD efímera, o salta el test si no hay una.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL no está definida — corre ./scripts/test-db.sh")
	}
	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatalf("no se pudo abrir el pool: %v", err)
	}
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatalf("no se pudo conectar a %s: %v", dsn, err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// resetDB deja la BD limpia. auth.users va incluida porque el trigger
// on_auth_user_created cuelga public.users de ella.
func resetDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(t.Context(), `
		TRUNCATE debts, punishment_suggestions, roulette_entries, checkins,
		         user_habits, group_habits, proposal_votes,
		         proposal_eligible_voters, proposals, memberships, groups,
		         habits, users, auth.users
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("no se pudo limpiar la BD: %v", err)
	}
}

// fixture es un grupo con un miembro y un hábito ya asignado: el estado
// mínimo desde el que arranca cualquier prueba del core loop.
type fixture struct {
	GroupID     string
	UserID      string
	OtherUserID string
	HabitID     string
	UserHabitID string
}

// seedGroup crea auth.users (el trigger llena public.users), el grupo, las
// membresías, el hábito del catálogo, su adopción por el grupo y la
// asignación al miembro.
func seedGroup(t *testing.T, pool *pgxpool.Pool, tz string) fixture {
	t.Helper()
	ctx := t.Context()
	var f fixture

	insertUser := func(email, name string) string {
		var id string
		err := pool.QueryRow(ctx,
			`INSERT INTO auth.users (email, raw_user_meta_data)
			 VALUES ($1, jsonb_build_object('display_name', $2::text))
			 RETURNING id`, email, name).Scan(&id)
		if err != nil {
			t.Fatalf("insert auth.users(%s): %v", email, err)
		}
		if tz != "" {
			if _, err := pool.Exec(ctx, `UPDATE users SET timezone = $2 WHERE id = $1`, id, tz); err != nil {
				t.Fatalf("set timezone: %v", err)
			}
		}
		return id
	}

	f.UserID = insertUser("debtor@test.local", "Deudor")
	f.OtherUserID = insertUser("other@test.local", "Otro")

	if err := pool.QueryRow(ctx,
		`INSERT INTO groups (name, invite_code, created_by)
		 VALUES ('Grupo Test', 'TESTCODE', $1) RETURNING id`,
		f.UserID).Scan(&f.GroupID); err != nil {
		t.Fatalf("insert group: %v", err)
	}

	for _, u := range []struct {
		id   string
		role string
	}{{f.UserID, "owner"}, {f.OtherUserID, "member"}} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO memberships (group_id, user_id, role, status)
			 VALUES ($1, $2, $3, 'active')`, f.GroupID, u.id, u.role); err != nil {
			t.Fatalf("insert membership: %v", err)
		}
	}

	if err := pool.QueryRow(ctx,
		`INSERT INTO habits (name, description, icon_key, color)
		 VALUES ('Ejercicio', 'test', 'exercise', '#C9714A') RETURNING id`).
		Scan(&f.HabitID); err != nil {
		t.Fatalf("insert habit: %v", err)
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO group_habits (group_id, habit_id, added_by)
		 VALUES ($1, $2, $3)`, f.GroupID, f.HabitID, f.UserID); err != nil {
		t.Fatalf("insert group_habit: %v", err)
	}

	if err := pool.QueryRow(ctx,
		`INSERT INTO user_habits (user_id, group_id, habit_id)
		 VALUES ($1, $2, $3) RETURNING id`,
		f.UserID, f.GroupID, f.HabitID).Scan(&f.UserHabitID); err != nil {
		t.Fatalf("insert user_habit: %v", err)
	}

	return f
}

// seedSpinnableEntry crea una entrada de ruleta lista para girar: la fecha de
// creación va al pasado para poder poner el deadline también en el pasado sin
// violar chk_roulette_deadline_after_creation (deadline > created_at).
func seedSpinnableEntry(t *testing.T, pool *pgxpool.Pool, f fixture) string {
	t.Helper()
	var entryID string
	err := pool.QueryRow(t.Context(),
		`INSERT INTO roulette_entries
		     (group_id, debtor_id, week_start, suggestion_deadline, created_at)
		 VALUES ($1, $2, date_trunc('week', now())::date,
		         now() - interval '1 hour', now() - interval '3 hours')
		 RETURNING id`, f.GroupID, f.UserID).Scan(&entryID)
	if err != nil {
		t.Fatalf("insert roulette_entry: %v", err)
	}
	return entryID
}

func countDebts(t *testing.T, pool *pgxpool.Pool, entryID string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM debts WHERE roulette_entry_id = $1`, entryID).Scan(&n); err != nil {
		t.Fatalf("count debts: %v", err)
	}
	return n
}

// TestSpinConcurrentCreatesSingleDebt es la prueba que respalda la afirmación
// de concurrencia del paper: dos spins simultáneos sobre la misma entrada
// deben producir UNA sola deuda. El segundo se bloquea en el
// SELECT ... FOR UPDATE hasta que el primero commitea, y entonces ve spun_at
// ya puesto y responde 409 en vez de crear una deuda duplicada.
func TestSpinConcurrentCreatesSingleDebt(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "")
	entryID := seedSpinnableEntry(t, pool, f)

	r := setupRouter(pool)

	const attempts = 2
	start := make(chan struct{})
	codes := make([]int, attempts)
	var wg sync.WaitGroup

	for i := range attempts {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost,
				"/penalties/roulette/"+entryID+"/spin", nil)
			req.Header.Set("X-User-ID", f.UserID)
			w := httptest.NewRecorder()
			<-start // barrera: los dos arrancan lo más junto posible
			r.ServeHTTP(w, req)
			codes[idx] = w.Code
		}(i)
	}
	close(start)
	wg.Wait()

	if got := countDebts(t, pool, entryID); got != 1 {
		t.Fatalf("se esperaba exactamente 1 deuda, se crearon %d (códigos: %v)", got, codes)
	}

	var created, conflict int
	for _, c := range codes {
		switch c {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			conflict++
		default:
			t.Errorf("código inesperado: %d", c)
		}
	}
	if created != 1 || conflict != 1 {
		t.Fatalf("se esperaba 1×201 y 1×409, se obtuvo %v", codes)
	}
}

// TestSpinTwiceSequentiallyIsRejected cubre el camino no concurrente: una vez
// girada, la entrada queda cerrada para siempre.
func TestSpinTwiceSequentiallyIsRejected(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "")
	entryID := seedSpinnableEntry(t, pool, f)

	r := setupRouter(pool)
	spin := func() int {
		req := httptest.NewRequest(http.MethodPost,
			"/penalties/roulette/"+entryID+"/spin", nil)
		req.Header.Set("X-User-ID", f.UserID)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := spin(); code != http.StatusCreated {
		t.Fatalf("primer spin: se esperaba 201, got %d", code)
	}
	if code := spin(); code != http.StatusConflict {
		t.Fatalf("segundo spin: se esperaba 409, got %d", code)
	}
	if got := countDebts(t, pool, entryID); got != 1 {
		t.Fatalf("se esperaba 1 deuda, hay %d", got)
	}
}

// TestSpinBeforeDeadlineIsRejected: la ventana de sugerencias tiene que haber
// cerrado. Es un chequeo en runtime, no un CHECK — un CHECK se evalúa al
// escribir y no se vuelve falso solo porque pase el tiempo.
func TestSpinBeforeDeadlineIsRejected(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "")

	var entryID string
	err := pool.QueryRow(t.Context(),
		`INSERT INTO roulette_entries
		     (group_id, debtor_id, week_start, suggestion_deadline)
		 VALUES ($1, $2, date_trunc('week', now())::date, now() + interval '2 hours')
		 RETURNING id`, f.GroupID, f.UserID).Scan(&entryID)
	if err != nil {
		t.Fatalf("insert roulette_entry: %v", err)
	}

	r := setupRouter(pool)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette/"+entryID+"/spin", nil)
	req.Header.Set("X-User-ID", f.UserID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("se esperaba 409 antes del deadline, got %d", w.Code)
	}
	if got := countDebts(t, pool, entryID); got != 0 {
		t.Fatalf("no debía crearse deuda, hay %d", got)
	}
}

// TestSpinPicksAmongSubmittedSuggestions: con sugerencias, la deuda tiene que
// salir de una de ellas (y quedar ligada por winning_suggestion_id), no del
// catálogo por defecto.
func TestSpinPicksAmongSubmittedSuggestions(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "")
	entryID := seedSpinnableEntry(t, pool, f)

	const suggestion = "Cantar en el camión"
	if _, err := pool.Exec(t.Context(),
		`INSERT INTO punishment_suggestions
		     (roulette_entry_id, group_id, suggester_id, text)
		 VALUES ($1, $2, $3, $4)`,
		entryID, f.GroupID, f.OtherUserID, suggestion); err != nil {
		t.Fatalf("insert suggestion: %v", err)
	}

	r := setupRouter(pool)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette/"+entryID+"/spin", nil)
	req.Header.Set("X-User-ID", f.UserID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("se esperaba 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var got struct {
		PunishmentText      string  `json:"punishment_text"`
		WinningSuggestionID *string `json:"winning_suggestion_id"`
	}
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode respuesta: %v", err)
	}
	if got.PunishmentText != suggestion {
		t.Errorf("con una sola sugerencia debía ganar %q, ganó %q", suggestion, got.PunishmentText)
	}
	if got.WinningSuggestionID == nil {
		t.Error("winning_suggestion_id no debía ser nulo cuando hubo sugerencias")
	}
}

// TestSpinWithoutSuggestionsDrawsFromCatalog: sin sugerencias se sortea del
// catálogo y la deuda queda sin winning_suggestion_id.
func TestSpinWithoutSuggestionsDrawsFromCatalog(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "")
	entryID := seedSpinnableEntry(t, pool, f)

	r := setupRouter(pool)
	req := httptest.NewRequest(http.MethodPost, "/penalties/roulette/"+entryID+"/spin", nil)
	req.Header.Set("X-User-ID", f.UserID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("se esperaba 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var got struct {
		PunishmentText      string  `json:"punishment_text"`
		WinningSuggestionID *string `json:"winning_suggestion_id"`
	}
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode respuesta: %v", err)
	}
	if strings.TrimSpace(got.PunishmentText) == "" {
		t.Error("la deuda del catálogo debía traer texto")
	}
	if got.WinningSuggestionID != nil {
		t.Errorf("sin sugerencias, winning_suggestion_id debía ser nulo, fue %q", *got.WinningSuggestionID)
	}
}

// TestEligibilityMemberWhoCheckedInEveryDayIsNotEligible es el invariante que
// no depende del día: si cumpliste lunes..ayer, no entras a la ruleta.
func TestEligibilityMemberWhoCheckedInEveryDayIsNotEligible(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "UTC")
	ctx := t.Context()

	// El hábito lleva 30 días asignado: si no, la ventana de fallos sería vacía
	// y el test pasaría sin probar nada.
	backdateHabitAssignment(t, pool, f, 30)

	// Un check-in por cada día desde el lunes hasta ayer inclusive.
	if _, err := pool.Exec(ctx,
		`INSERT INTO checkins (user_habit_id, checked_on)
		 SELECT $1, d::date
		 FROM generate_series(
		     date_trunc('week', (now() AT TIME ZONE 'UTC'))::date,
		     (now() AT TIME ZONE 'UTC')::date - 1,
		     interval '1 day'
		 ) d`, f.UserHabitID); err != nil {
		t.Fatalf("insert checkins: %v", err)
	}

	eligible, err := db.IsEligibleForRoulette(ctx, pool, f.GroupID, f.UserID)
	if err != nil {
		t.Fatalf("IsEligibleForRoulette: %v", err)
	}
	if eligible {
		t.Error("quien cumplió todos los días no debía ser elegible para la ruleta")
	}
}

// backdateHabitAssignment mueve la fecha de asignación del hábito al pasado. Sin
// esto, seedGroup lo asigna HOY y —correctamente— el miembro nunca es elegible,
// porque la ventana de fallos arranca el día de la asignación.
func backdateHabitAssignment(t *testing.T, pool *pgxpool.Pool, f fixture, days int) {
	t.Helper()
	// make_interval en vez de concatenar texto: pgx no puede codificar un int
	// como text para el operador ||.
	_, err := pool.Exec(t.Context(),
		`UPDATE user_habits SET created_at = now() - make_interval(days => $2) WHERE id = $1`,
		f.UserHabitID, days)
	if err != nil {
		t.Fatalf("backdate user_habit: %v", err)
	}
}

// TestEligibilityMemberWithNoCheckinsIsEligible: cero check-ins → elegible, pero
// solo si el hábito ya llevaba tiempo asignado. Se salta el lunes: por diseño la
// ventana está vacía ese día y el test no probaría nada.
func TestEligibilityMemberWithNoCheckinsIsEligible(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "UTC")
	ctx := t.Context()

	var today time.Time
	if err := pool.QueryRow(ctx, `SELECT (now() AT TIME ZONE 'UTC')::date`).Scan(&today); err != nil {
		t.Fatalf("leer fecha del servidor: %v", err)
	}
	if today.Weekday() == time.Monday {
		t.Skip("en lunes la ventana está vacía y nadie es elegible por diseño")
	}

	// El hábito lleva 30 días asignado: la ventana es toda la semana.
	backdateHabitAssignment(t, pool, f, 30)

	eligible, err := db.IsEligibleForRoulette(ctx, pool, f.GroupID, f.UserID)
	if err != nil {
		t.Fatalf("IsEligibleForRoulette: %v", err)
	}
	if !eligible {
		t.Error("sin check-ins en la semana el miembro debía ser elegible")
	}
}

// TestEligibilityIgnoresDaysBeforeTheHabitExisted es la regresión del bug que se
// veía en la web: entrar al grupo un domingo pintaba los seis días previos en
// rojo y metía al recién llegado directo a la ruleta. Nadie puede fallar un
// hábito que todavía no tenía.
func TestEligibilityIgnoresDaysBeforeTheHabitExisted(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "UTC")
	ctx := t.Context()

	// seedGroup asigna el hábito HOY, que es justo el caso del bug.
	eligible, err := db.IsEligibleForRoulette(ctx, pool, f.GroupID, f.UserID)
	if err != nil {
		t.Fatalf("IsEligibleForRoulette: %v", err)
	}
	if eligible {
		t.Error("con el hábito asignado hoy no hay dia previo que fallar: no debía ser elegible")
	}

	// Y no debe aparecer en la lista del grupo.
	members, err := db.GetEligibleMembers(ctx, pool, f.GroupID)
	if err != nil {
		t.Fatalf("GetEligibleMembers: %v", err)
	}
	for _, m := range members {
		if m.UserID == f.UserID {
			t.Error("tampoco debía salir en GetEligibleMembers")
		}
	}
}

// TestEligibilityCountsOnlyFromAssignmentDay: con el hábito asignado ayer y sin
// check-ins, hoy ya hay exactamente un día fallado → elegible. Fija el borde.
func TestEligibilityCountsOnlyFromAssignmentDay(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "UTC")
	ctx := t.Context()

	var today time.Time
	if err := pool.QueryRow(ctx, `SELECT (now() AT TIME ZONE 'UTC')::date`).Scan(&today); err != nil {
		t.Fatalf("leer fecha: %v", err)
	}
	if today.Weekday() == time.Monday {
		t.Skip("en lunes la ventana de la semana no alcanza a incluir ayer")
	}

	backdateHabitAssignment(t, pool, f, 1)

	eligible, err := db.IsEligibleForRoulette(ctx, pool, f.GroupID, f.UserID)
	if err != nil {
		t.Fatalf("IsEligibleForRoulette: %v", err)
	}
	if !eligible {
		t.Error("hábito asignado ayer y sin check-in: ese día cuenta como fallado")
	}
}

// TestStreakCountsConsecutiveDays: la racha cuenta días consecutivos hacia
// atrás desde hoy.
func TestStreakCountsConsecutiveDays(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "UTC")
	ctx := t.Context()

	if _, err := pool.Exec(ctx,
		`INSERT INTO checkins (user_habit_id, checked_on)
		 SELECT $1, ((now() AT TIME ZONE 'UTC')::date - g)
		 FROM generate_series(0, 2) g`, f.UserHabitID); err != nil {
		t.Fatalf("insert checkins: %v", err)
	}

	streaks, err := db.GetStreaksForUser(ctx, pool, f.UserID)
	if err != nil {
		t.Fatalf("GetStreaksForUser: %v", err)
	}
	if len(streaks) != 1 {
		t.Fatalf("se esperaba 1 racha, hay %d", len(streaks))
	}
	if streaks[0].Current != 3 {
		t.Errorf("se esperaba racha de 3 días consecutivos, got %d", streaks[0].Current)
	}
}

// TestCheckinIsIdempotentPerDay: la UNIQUE (user_habit_id, checked_on) impide
// dos check-ins del mismo hábito el mismo día.
func TestCheckinIsIdempotentPerDay(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool, "UTC")
	ctx := t.Context()

	var today string
	if err := pool.QueryRow(ctx, `SELECT (now() AT TIME ZONE 'UTC')::date::text`).Scan(&today); err != nil {
		t.Fatalf("leer fecha: %v", err)
	}

	if err := db.CreateCheckin(ctx, pool, f.UserHabitID, today, nil); err != nil {
		t.Fatalf("primer check-in: %v", err)
	}
	// El segundo no debe duplicar la fila, ya sea porque el upsert lo absorbe
	// o porque la constraint lo rechaza. Lo que importa es el conteo final.
	_ = db.CreateCheckin(ctx, pool, f.UserHabitID, today, nil)

	var n int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM checkins WHERE user_habit_id = $1 AND checked_on = $2::date`,
		f.UserHabitID, today).Scan(&n); err != nil {
		t.Fatalf("count checkins: %v", err)
	}
	if n != 1 {
		t.Errorf("se esperaba 1 check-in para el día, hay %d", n)
	}
}
