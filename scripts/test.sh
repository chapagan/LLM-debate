#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

export GOCACHE=${GOCACHE:-"$ROOT_DIR/.cache/go-build"}
mkdir -p "$GOCACHE"

. "$ROOT_DIR/scripts/lib/go-version.sh"
require_go_version

echo "Running Go tests..."
go test ./...

echo "Running frontend tests..."
(cd frontend && npm test)
