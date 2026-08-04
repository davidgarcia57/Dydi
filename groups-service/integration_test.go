//go:build integration

// Tests de integración de groups-service contra un Postgres REAL.
//
// El resto de la suite pasa `nil` como pool y solo cubre las guardas de
// entrada, así que ningún SQL se ejecuta en CI. Ese hueco dejó vivir un bug
// real: SetProposalStatus fallaba SIEMPRE por un desajuste uuid/text y nadie
// se enteraba, porque el handler solo lo loguea y responde 204 igual.
// TestSetProposalStatusPersists existe para que eso no vuelva a pasar.
//
// Se saltan solos si TEST_DATABASE_URL no está. Para correrlos:
// ./scripts/test-db.sh
package main

import (
	"errors"
	"os"
	"testing"

	"github.com/dydi/groups-service/internal/db"
	"github.com/dydi/groups-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		t.Fatalf("no se pudo conectar: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

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

type fixture struct {
	GroupID string
	OwnerID string
	MateID  string
	HabitID string
}

func seedGroup(t *testing.T, pool *pgxpool.Pool) fixture {
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
		return id
	}

	f.OwnerID = insertUser("owner@test.local", "Owner")
	f.MateID = insertUser("mate@test.local", "Mate")

	g, err := db.CreateGroupWithOwner(ctx, pool, "Grupo Test", "TESTCODE", f.OwnerID)
	if err != nil {
		t.Fatalf("CreateGroupWithOwner: %v", err)
	}
	f.GroupID = g.ID

	if err := db.AddMember(ctx, pool, f.GroupID, f.MateID, "member"); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	if err := pool.QueryRow(ctx,
		`INSERT INTO habits (name, description, icon_key, color)
		 VALUES ('Ejercicio', 'test', 'exercise', '#C9714A') RETURNING id`).
		Scan(&f.HabitID); err != nil {
		t.Fatalf("insert habit: %v", err)
	}

	return f
}

// TestSetProposalStatusPersists es el test de regresión del bug de tipos.
// Antes del cast a ::uuid, este UPDATE reventaba con "column resolved_by is of
// type uuid but expression is of type text" y la propuesta se quedaba en
// 'open' para siempre, sin que la API lo delatara.
func TestSetProposalStatusPersists(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	p, err := db.CreateProposal(ctx, pool, f.GroupID, f.OwnerID,
		model.ProposalAddHabit, &f.HabitID, nil)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}

	// El error tiene que salir por el valor de retorno, no solo por un log.
	if err := db.SetProposalStatus(ctx, pool, p.ID, model.ProposalApproved, &f.OwnerID); err != nil {
		t.Fatalf("SetProposalStatus devolvió error: %v", err)
	}

	var status string
	var resolvedBy *string
	var resolvedAt *string
	if err := pool.QueryRow(ctx,
		`SELECT status, resolved_by::text, resolved_at::text FROM proposals WHERE id = $1`,
		p.ID).Scan(&status, &resolvedBy, &resolvedAt); err != nil {
		t.Fatalf("leer propuesta: %v", err)
	}

	if status != string(model.ProposalApproved) {
		t.Errorf("status: se esperaba %q, quedó %q", model.ProposalApproved, status)
	}
	if resolvedAt == nil {
		t.Error("resolved_at debía quedar estampado al salir de 'open'")
	}
	if resolvedBy == nil || *resolvedBy != f.OwnerID {
		t.Errorf("resolved_by: se esperaba %q, quedó %v", f.OwnerID, resolvedBy)
	}
}

// TestRotateInviteCodeRevokesTheOldOne es el test del mecanismo de revocación: lo
// que importa no es que el código cambie, sino que el ANTERIOR deje de servir.
// Sin esto, un código filtrado valía para siempre.
func TestRotateInviteCodeRevokesTheOldOne(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	before, err := db.GetGroupByID(ctx, pool, f.GroupID)
	if err != nil {
		t.Fatalf("leer grupo: %v", err)
	}

	nuevo, err := db.GenerateInviteCode()
	if err != nil {
		t.Fatalf("GenerateInviteCode: %v", err)
	}

	after, err := db.RotateInviteCode(ctx, pool, f.GroupID, nuevo)
	if err != nil {
		t.Fatalf("RotateInviteCode: %v", err)
	}

	if after.InviteCode == before.InviteCode {
		t.Error("el código no cambió")
	}
	if after.InviteCode != nuevo {
		t.Errorf("se esperaba %q, quedó %q", nuevo, after.InviteCode)
	}
	// Lo importante: el viejo ya no resuelve a ningún grupo.
	if _, err := db.GetGroupByInviteCode(ctx, pool, before.InviteCode); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("el código viejo sigue sirviendo: err=%v", err)
	}
	// Y el nuevo sí, que es lo que usa /groups/join-by-code.
	got, err := db.GetGroupByInviteCode(ctx, pool, nuevo)
	if err != nil {
		t.Fatalf("el código nuevo no resuelve: %v", err)
	}
	if got.ID != f.GroupID {
		t.Errorf("resolvió al grupo %q en vez de %q", got.ID, f.GroupID)
	}
}

// TestRotateInviteCodeOnMissingGroup: el UPDATE ... RETURNING de un grupo que no
// existe devuelve ErrNoRows, que es lo que el handler traduce a 404.
func TestRotateInviteCodeOnMissingGroup(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	ctx := t.Context()

	const ausente = "3f2504e0-4f89-11d3-9a0c-0305e82c3302"
	if _, err := db.RotateInviteCode(ctx, pool, ausente, "ABCD2345"); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("se esperaba ErrNoRows, salió: %v", err)
	}
}

// TestSetProposalStatusNoRowsIsAnError cubre el otro modo de fallo silencioso: un
// UPDATE que no encaja con ninguna fila. Postgres lo reporta como sentencia
// exitosa con 0 filas, así que antes de comprobar RowsAffected el handler
// registraba éxito mientras la propuesta se quedaba en 'open'.
func TestSetProposalStatusNoRowsIsAnError(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	ctx := t.Context()

	// UUID válido que no existe: el UPDATE es legal pero no toca nada.
	const ausente = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"

	err := db.SetProposalStatus(ctx, pool, ausente, model.ProposalApproved, nil)
	if err == nil {
		t.Fatal("SetProposalStatus devolvió nil con 0 filas afectadas")
	}
	if !errors.Is(err, db.ErrProposalNotUpdated) {
		t.Errorf("se esperaba ErrProposalNotUpdated, salió: %v", err)
	}
}

// TestSetProposalStatusExpiredWithoutResolver cubre la ruta del sistema, donde
// resolvedBy va nil (caducidad perezosa, sin usuario que la dispare).
func TestSetProposalStatusExpiredWithoutResolver(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	p, err := db.CreateProposal(ctx, pool, f.GroupID, f.OwnerID,
		model.ProposalAddHabit, &f.HabitID, nil)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}

	if err := db.SetProposalStatus(ctx, pool, p.ID, model.ProposalExpired, nil); err != nil {
		t.Fatalf("SetProposalStatus(nil resolver): %v", err)
	}

	var status string
	var resolvedBy *string
	if err := pool.QueryRow(ctx,
		`SELECT status, resolved_by::text FROM proposals WHERE id = $1`,
		p.ID).Scan(&status, &resolvedBy); err != nil {
		t.Fatalf("leer propuesta: %v", err)
	}
	if status != string(model.ProposalExpired) {
		t.Errorf("status: se esperaba %q, quedó %q", model.ProposalExpired, status)
	}
	if resolvedBy != nil {
		t.Errorf("resolved_by debía quedar nulo, quedó %q", *resolvedBy)
	}
}

// TestElectorateIsFrozenAtCreation: quien entra al grupo DESPUÉS de abrirse la
// propuesta no vota. Es lo que hace que el quórum tenga un denominador estable.
func TestElectorateIsFrozenAtCreation(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	p, err := db.CreateProposal(ctx, pool, f.GroupID, f.OwnerID,
		model.ProposalAddHabit, &f.HabitID, nil)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}

	// Tercer miembro, ya abierta la propuesta.
	var lateID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO auth.users (email, raw_user_meta_data)
		 VALUES ('late@test.local', '{"display_name":"Late"}'::jsonb) RETURNING id`).
		Scan(&lateID); err != nil {
		t.Fatalf("insert late user: %v", err)
	}
	if err := db.AddMember(ctx, pool, f.GroupID, lateID, "member"); err != nil {
		t.Fatalf("AddMember(late): %v", err)
	}

	eligible, err := db.IsEligibleVoter(ctx, pool, p.ID, lateID)
	if err != nil {
		t.Fatalf("IsEligibleVoter: %v", err)
	}
	if eligible {
		t.Error("un miembro que llegó después no debía estar en el electorado congelado")
	}

	_, members, err := db.CountApprovalVotes(ctx, pool, p.ID)
	if err != nil {
		t.Fatalf("CountApprovalVotes: %v", err)
	}
	if members != 2 {
		t.Errorf("el electorado debía quedar en 2 (los de la creación), quedó %d", members)
	}
}

// TestQuorumReachedWithMajorityOfFrozenElectorate: 2 de 2 aprueban → quórum.
func TestQuorumReachedWithMajorityOfFrozenElectorate(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	p, err := db.CreateProposal(ctx, pool, f.GroupID, f.OwnerID,
		model.ProposalAddHabit, &f.HabitID, nil)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}

	if err := db.CastVote(ctx, pool, p.ID, f.OwnerID, true); err != nil {
		t.Fatalf("CastVote(owner): %v", err)
	}
	approvals, members, err := db.CountApprovalVotes(ctx, pool, p.ID)
	if err != nil {
		t.Fatalf("CountApprovalVotes: %v", err)
	}
	if approvals != 1 || members != 2 {
		t.Fatalf("tras 1 voto: se esperaba 1/2, got %d/%d", approvals, members)
	}

	if err := db.CastVote(ctx, pool, p.ID, f.MateID, true); err != nil {
		t.Fatalf("CastVote(mate): %v", err)
	}
	approvals, members, err = db.CountApprovalVotes(ctx, pool, p.ID)
	if err != nil {
		t.Fatalf("CountApprovalVotes: %v", err)
	}
	if approvals != 2 || members != 2 {
		t.Fatalf("tras 2 votos: se esperaba 2/2, got %d/%d", approvals, members)
	}
	// La aritmética del quórum vive en internal/handler y ya tiene su test
	// unitario (TestQuorumReached). Aquí solo verificamos que el conteo que la
	// alimenta salga correcto de la BD.
}

// TestRejectionVoteDoesNotCountAsApproval: un voto en contra no acerca al
// quórum. Es el escenario que me hizo tropezar probando en vivo: mandar
// {"vote":true} en vez de {"approved":true} decodifica a `false`.
func TestRejectionVoteDoesNotCountAsApproval(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	p, err := db.CreateProposal(ctx, pool, f.GroupID, f.OwnerID,
		model.ProposalAddHabit, &f.HabitID, nil)
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}
	if err := db.CastVote(ctx, pool, p.ID, f.OwnerID, false); err != nil {
		t.Fatalf("CastVote: %v", err)
	}

	approvals, members, err := db.CountApprovalVotes(ctx, pool, p.ID)
	if err != nil {
		t.Fatalf("CountApprovalVotes: %v", err)
	}
	if approvals != 0 {
		t.Errorf("un voto en contra no debía contar como aprobación, got %d", approvals)
	}
	if members != 2 {
		t.Errorf("el electorado debía seguir en 2, quedó %d", members)
	}
}

// TestGroupSizeLimitIsEnforcedByTrigger: el techo de 8 activos lo pone la BD,
// no la app. Si alguien mete un INSERT por otra vía, el trigger igual lo para.
func TestGroupSizeLimitIsEnforcedByTrigger(t *testing.T) {
	pool := testPool(t)
	resetDB(t, pool)
	f := seedGroup(t, pool)
	ctx := t.Context()

	// Ya hay 2 miembros; metemos 6 más para llegar a 8, y el noveno debe fallar.
	for i := range 7 {
		var id string
		if err := pool.QueryRow(ctx,
			`INSERT INTO auth.users (email, raw_user_meta_data)
			 VALUES ($1, '{"display_name":"Extra"}'::jsonb) RETURNING id`,
			"extra"+string(rune('a'+i))+"@test.local").Scan(&id); err != nil {
			t.Fatalf("insert extra user %d: %v", i, err)
		}
		err := db.AddMember(ctx, pool, f.GroupID, id, "member")
		if i < 6 {
			if err != nil {
				t.Fatalf("AddMember %d debía pasar (aún hay lugar): %v", i, err)
			}
			continue
		}
		// El noveno.
		if err == nil {
			t.Fatal("el noveno miembro activo debía ser rechazado por el trigger de tamaño")
		}
	}

	var active int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM memberships WHERE group_id = $1 AND status = 'active'`,
		f.GroupID).Scan(&active); err != nil {
		t.Fatalf("count active: %v", err)
	}
	if active != 8 {
		t.Errorf("se esperaban 8 miembros activos, hay %d", active)
	}
}
