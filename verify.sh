#!/usr/bin/env bash
# verify.sh — verificación de UN comando para Dydi. Espeja el CI, todo en Docker.
# Go y Node NO están instalados localmente: este script maneja los contenedores.
#
# En Windows, córrelo desde la distro WSL:
#   wsl -d ubuntu bash -lc './verify.sh'
#
# Uso:
#   ./verify.sh            # todo (Go x4 + frontend + móvil)
#   ./verify.sh go         # solo los 4 servicios Go
#   ./verify.sh frontend   # solo el frontend Vue
#   ./verify.sh mobile     # solo el typecheck del móvil
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"
mkdir -p .gocache .gomodcache

GO_SERVICES="api-gateway groups-service habits-service realtime-service"
GO_IMAGE="golang:1.24"
LINT_IMAGE="golangci/golangci-lint:v2.12-alpine"
NODE_IMAGE="node:20-alpine"

green() { printf '\n\033[1;32m== %s ==\033[0m\n' "$1"; }
fail()  { printf '\n\033[1;31m✗ FALLÓ: %s\033[0m\n' "$1"; exit 1; }

verify_go() {
  for svc in $GO_SERVICES; do
    green "Go: $svc — gofmt · vet · build · test -race"
    docker run --rm \
      -v "$ROOT/$svc":/app -v "$ROOT/.gocache":/gocache -v "$ROOT/.gomodcache":/gomodcache \
      -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -w /app "$GO_IMAGE" \
      sh -c 'test -z "$(gofmt -l .)" && go vet ./... && go build ./... && go test -race ./...' \
      || fail "$svc (gofmt/vet/build/test)"

    green "Go: $svc — golangci-lint"
    # La config se monta FUERA de /app (en /golangci.yml) a propósito: montarla
    # dentro del dir del servicio deja un archivo .golangci.yml vacío (artefacto
    # del bind-mount de Docker).
    docker run --rm \
      -v "$ROOT/$svc":/app -v "$ROOT/.golangci.yml":/golangci.yml \
      -v "$ROOT/.gocache":/gocache -v "$ROOT/.gomodcache":/gomodcache \
      -e GOCACHE=/gocache -e GOMODCACHE=/gomodcache -w /app "$LINT_IMAGE" \
      golangci-lint run --config /golangci.yml \
      || fail "$svc (golangci-lint)"
  done
}

verify_frontend() {
  green "Frontend — lint · format · build · test"
  # Se copia a un dir interno (/w) porque los symlinks de node_modules/.bin se
  # rompen en el mount de Windows.
  docker run --rm -v "$ROOT/frontend":/src -v dydi_npm_cache:/root/.npm "$NODE_IMAGE" sh -c '
    cp -r /src /w && cd /w && rm -rf node_modules &&
    npm ci --prefer-offline && npm run lint && npm run format:check && npm run build && npm run test' \
    || fail "frontend"
}

verify_mobile() {
  green "Móvil — tsc --noEmit"
  docker run --rm -v "$ROOT/mobile":/src -v dydi_npm_cache:/root/.npm "$NODE_IMAGE" sh -c '
    cp -r /src /w && cd /w && rm -rf node_modules &&
    npm ci --legacy-peer-deps --prefer-offline && npx tsc --noEmit' \
    || fail "mobile"
}

case "${1:-all}" in
  go)       verify_go ;;
  frontend) verify_frontend ;;
  mobile)   verify_mobile ;;
  all)      verify_go; verify_frontend; verify_mobile ;;
  *)        echo "uso: ./verify.sh [go|frontend|mobile|all]"; exit 2 ;;
esac

printf '\n\033[1;32m✓ TODO VERDE\033[0m\n'
