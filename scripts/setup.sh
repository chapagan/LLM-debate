#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

export GOCACHE=${GOCACHE:-"$ROOT_DIR/.cache/go-build"}
mkdir -p "$GOCACHE"

. "$ROOT_DIR/scripts/lib/go-version.sh"
require_go_version

command -v npm >/dev/null 2>&1 || {
  echo "npm is required. Install Node.js 20 or newer, then rerun this script." >&2
  exit 1
}

echo "Downloading Go modules..."
go mod download

echo "Installing frontend dependencies..."
(cd frontend && npm ci)

echo "Setup complete."
