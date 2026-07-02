#!/usr/bin/env bash
# verify.sh - verificacion de UN comando para Dydi. Espeja el CI, todo en Docker.
# Go y Node NO estan instalados localmente: este script maneja los contenedores.
#
# En Windows, correlo desde la distro WSL:
#   wsl -d ubuntu bash -lc './verify.sh'
#
# Uso:
#   ./verify.sh                  # todo (Go x4 + frontend + movil)
#   ./verify.sh go               # solo los 4 servicios Go
#   ./verify.sh go:habits-service
#   ./verify.sh frontend         # solo el frontend Vue
#   ./verify.sh mobile           # solo el typecheck del movil
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"
mkdir -p .gocache .gomodcache .npmcache

GO_SERVICES="api-gateway groups-service habits-service realtime-service"
GO_IMAGE="golang:1.24"
LINT_IMAGE="golangci/golangci-lint:v2.12-alpine"
NODE_IMAGE="node:22-alpine"

green() { printf '\n\033[1;32m== %s ==\033[0m\n' "$1"; }
fail()  { printf '\n\033[1;31mX FALLO: %s\033[0m\n' "$1"; exit 1; }

is_go_service() {
  case " $GO_SERVICES " in
    *" $1 "*) return 0 ;;
    *) return 1 ;;
  esac
}

verify_go_service() {
  local svc="$1"
  if ! is_go_service "$svc"; then
    echo "servicio Go desconocido: $svc"
    echo "servicios validos: $GO_SERVICES"
    exit 2
  fi

  green "Go: $svc - gofmt, vet, build, test -race"
  docker run --rm \
    -v "$ROOT":/workspace -v "$ROOT/.gocache":/gocache -v "$ROOT/.gomodcache":/gomodcache \
    -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -w "/workspace/$svc" "$GO_IMAGE" \
    sh -c 'test -z "$(gofmt -l .)" && go vet ./... && go build -buildvcs=false ./... && go test -buildvcs=false -race ./...' \
    || fail "$svc (gofmt/vet/build/test)"

  green "Go: $svc - golangci-lint"
  # La config se monta fuera de /workspace/$svc (en /golangci.yml) a proposito.
  docker run --rm \
    -v "$ROOT":/workspace -v "$ROOT/.golangci.yml":/golangci.yml \
    -v "$ROOT/.gocache":/gocache -v "$ROOT/.gomodcache":/gomodcache \
    -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -w "/workspace/$svc" "$LINT_IMAGE" \
    golangci-lint run --config /golangci.yml \
    || fail "$svc (golangci-lint)"
}

verify_go() {
  for svc in $GO_SERVICES; do
    verify_go_service "$svc"
  done
}

verify_frontend() {
  green "Frontend - lint, format, build, test"
  # Se copia a un dir interno (/w) porque los symlinks de node_modules/.bin se
  # rompen en el mount de Windows.
  docker run --rm -v "$ROOT/frontend":/src -v "$ROOT/.npmcache":/root/.npm "$NODE_IMAGE" sh -c '
    cp -r /src /w && cd /w && rm -rf node_modules &&
    npm ci --prefer-offline && npm run lint && npm run format:check && npm run build && npm run test' \
    || fail "frontend"
}

verify_mobile() {
  green "Movil - tsc --noEmit y tests"
  docker run --rm -v "$ROOT/mobile":/src -v "$ROOT/.npmcache":/root/.npm "$NODE_IMAGE" sh -c '
    cp -r /src /w && cd /w && rm -rf node_modules &&
    npm ci --legacy-peer-deps --prefer-offline && npm install --no-save jest jest-expo @types/jest @react-native/jest-preset && npx tsc --noEmit && npx jest' \
    || fail "mobile"
}

case "${1:-all}" in
  go)       verify_go ;;
  go:*)     verify_go_service "${1#go:}" ;;
  frontend) verify_frontend ;;
  mobile)   verify_mobile ;;
  all)      verify_go; verify_frontend; verify_mobile ;;
  *)        echo "uso: ./verify.sh [go|go:<servicio>|frontend|mobile|all]"; exit 2 ;;
esac

printf '\n\033[1;32mOK TODO VERDE\033[0m\n'
