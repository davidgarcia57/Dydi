#!/usr/bin/env bash
# seed.sh — siembra los grupos de carga para la prueba k6 (tesis), vía Docker (psql).
#
# El handshake WS valida membresía (fail-closed), así que k6 necesita grupos REALES
# donde el usuario del TOKEN sea miembro activo. Este script crea N grupos
# 'loadtest-*' (invite_code determinista 'LOADTEST-<n>', idempotente), mete al
# usuario de prueba como owner activo y exporta los UUIDs a loadtest_groups.json
# para que k6 los reparta entre VUs (tope: 8 conexiones por grupo).
#
# ⚠️ A diferencia de scripts/q.sh, este SÍ escribe en la BD (Supabase cloud).
#    Solo toca filas con invite_code 'LOADTEST-%'; 'down' borra únicamente esas.
#
# En Windows, córrelo desde la distro WSL:
#   wsl -d ubuntu bash -lc './load-tests/seed.sh up -u <uuid> -n 1000'
#
# Uso:
#   ./load-tests/seed.sh up -u <uuid-usuario> [-n 1000] [--dry-run]
#   ./load-tests/seed.sh export        # regenera loadtest_groups.json
#   ./load-tests/seed.sh down          # borra los grupos loadtest (cascade)
#
# El usuario también puede venir por env: DYDI_USER=<uuid> ./load-tests/seed.sh up
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/load-tests/loadtest_groups.json"

# Lee DATABASE_URL del .env sin sourcearlo (la contraseña puede traer caracteres raros).
DB="$(grep -E '^DATABASE_URL=' "$ROOT/.env" 2>/dev/null | head -1 | cut -d= -f2-)"
[ -n "${DB:-}" ] || { echo "✗ Falta DATABASE_URL en .env (copia .env.example)"; exit 1; }

psql_run() { docker run --rm -i postgres:15 psql "$DB" -v ON_ERROR_STOP=1 "$@"; }

usage() {
  sed -n '2,22p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  exit 2
}

export_groups() {
  psql_run -t -A -c \
    "SELECT COALESCE(json_agg(id), '[]'::json) FROM groups WHERE invite_code LIKE 'LOADTEST-%';" \
    | tr -d '[:space:]' > "$OUT"
  N="$(python3 -c 'import json,sys; print(len(json.load(open(sys.argv[1]))))' "$OUT" 2>/dev/null || echo '?')"
  echo "✓ $N UUIDs exportados a load-tests/loadtest_groups.json"
}

CMD="${1:-}"; shift || true
case "$CMD" in
  up)
    USER_ID="${DYDI_USER:-}"
    COUNT=1000
    FINAL="COMMIT"
    while [ $# -gt 0 ]; do
      case "$1" in
        -u) USER_ID="${2:?falta el uuid después de -u}"; shift 2 ;;
        -n) COUNT="${2:?falta el número después de -n}"; shift 2 ;;
        --dry-run) FINAL="ROLLBACK"; shift ;;
        *) usage ;;
      esac
    done
    # Sin -u ni DYDI_USER: intenta resolver el uuid desde LOADTEST_EMAIL (.env).
    if [ -z "$USER_ID" ]; then
      LT_EMAIL="$(grep -E '^LOADTEST_EMAIL=' "$ROOT/.env" 2>/dev/null | head -1 | cut -d= -f2-)"
      if [ -n "$LT_EMAIL" ]; then
        USER_ID="$(psql_run -t -A -c "SELECT id FROM auth.users WHERE email = '$LT_EMAIL' LIMIT 1;" | tr -d '[:space:]')"
        [ -n "$USER_ID" ] && echo "→ usuario resuelto por LOADTEST_EMAIL ($LT_EMAIL): $USER_ID"
      fi
    fi
    [ -n "$USER_ID" ] || { echo "✗ Falta el usuario: -u <uuid>, env DYDI_USER, o LOADTEST_EMAIL en .env"; exit 1; }
    case "$COUNT" in (*[!0-9]*|'') echo "✗ -n debe ser un entero"; exit 1 ;; esac
    [ "$COUNT" -le 5000 ] || { echo "✗ -n máximo 5000 (¿seguro que necesitas más salas que VUs?)"; exit 1; }

    [ "$FINAL" = "ROLLBACK" ] && echo "── DRY-RUN: se valida todo contra la BD pero termina en ROLLBACK ──"

    psql_run -v user_id="$USER_ID" -v group_count="$COUNT" -f - <<SQL
-- Aborta con mensaje claro si el usuario no existe (evita un error FK críptico).
-- Va en DO + RAISE (no \quit): es lo único que hace fallar el exit code de psql.
SELECT set_config('dydi.seed_user', :'user_id', false);
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM users WHERE id = current_setting('dydi.seed_user')::uuid) THEN
    RAISE EXCEPTION 'el usuario % no existe en public.users (¿ya se registró en la app?)',
      current_setting('dydi.seed_user');
  END IF;
END \$\$;

BEGIN;

-- Grupos: idempotente por invite_code determinista (re-correr no duplica).
INSERT INTO groups (name, invite_code, created_by)
SELECT 'loadtest-' || g, 'LOADTEST-' || g, :'user_id'::uuid
FROM generate_series(1, :group_count) AS g
WHERE NOT EXISTS (SELECT 1 FROM groups WHERE invite_code = 'LOADTEST-' || g);

-- El usuario de prueba entra como owner activo a TODOS los grupos loadtest
-- (el tope de 8 es por miembros del grupo, no de grupos por usuario).
INSERT INTO memberships (group_id, user_id, role, status)
SELECT gr.id, :'user_id'::uuid, 'owner', 'active'
FROM groups gr
WHERE gr.invite_code LIKE 'LOADTEST-%'
ON CONFLICT (group_id, user_id) DO NOTHING;

SELECT count(*) AS grupos_loadtest_totales FROM groups WHERE invite_code LIKE 'LOADTEST-%';

$FINAL;
SQL

    if [ "$FINAL" = "COMMIT" ]; then
      export_groups
      echo "✓ Seed listo. El TOKEN de k6 debe ser de ESTE usuario: $USER_ID"
    else
      echo "✓ Dry-run OK (nada quedó escrito)."
    fi
    ;;

  export)
    export_groups
    ;;

  down)
    psql_run -c \
      "WITH del AS (DELETE FROM groups WHERE invite_code LIKE 'LOADTEST-%' RETURNING 1)
       SELECT count(*) AS grupos_borrados FROM del;"
    echo '[]' > "$OUT"
    echo "✓ Grupos loadtest borrados (memberships caen en cascada)."
    ;;

  *) usage ;;
esac
