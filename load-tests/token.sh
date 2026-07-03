#!/usr/bin/env bash
# token.sh — imprime un access_token FRESCO del usuario loadtest (Supabase).
#
# El JWT de Supabase caduca ~1h; en vez de copiarlo del navegador cada corrida,
# esto lo pide directo a la API de auth (password grant, igual que hace el
# frontend al iniciar sesión). Lo usa run_experiment.sh cuando no le pasas
# DYDI_JWT, y también sirve suelto:
#   DYDI_JWT=$(./load-tests/token.sh) ./scripts/hit.sh GET gateway /groups
#
# Requiere en .env (además de las VITE_* que ya existen):
#   LOADTEST_EMAIL=loadtest@...      (cuenta dedicada registrada en la app)
#   LOADTEST_PASSWORD=...
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

envget() { grep -E "^$1=" "$ROOT/.env" 2>/dev/null | head -1 | cut -d= -f2-; }

URL="$(envget VITE_SUPABASE_URL)"
KEY="$(envget VITE_SUPABASE_ANON_KEY)"
EMAIL="$(envget LOADTEST_EMAIL)"
PASS="$(envget LOADTEST_PASSWORD)"

[ -n "$URL" ] && [ -n "$KEY" ] || { echo "✗ Faltan VITE_SUPABASE_URL/VITE_SUPABASE_ANON_KEY en .env" >&2; exit 1; }
[ -n "$EMAIL" ] && [ -n "$PASS" ] || { echo "✗ Faltan LOADTEST_EMAIL/LOADTEST_PASSWORD en .env (cuenta dedicada de pruebas)" >&2; exit 1; }

# python3 arma y parsea el JSON para no pelearnos con caracteres raros en la
# contraseña (viene preinstalado en la distro; no se instala nada).
PAYLOAD="$(python3 -c 'import json,sys; print(json.dumps({"email": sys.argv[1], "password": sys.argv[2]}))' "$EMAIL" "$PASS")"

RESP="$(curl -sf --max-time 10 -X POST "$URL/auth/v1/token?grant_type=password" \
  -H "apikey: $KEY" -H 'Content-Type: application/json' -d "$PAYLOAD")" || {
  echo "✗ Supabase rechazó el login del usuario loadtest (¿credenciales mal en .env?)" >&2; exit 1; }

printf '%s' "$RESP" | python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])'
