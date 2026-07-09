#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

export GOCACHE=${GOCACHE:-"$ROOT_DIR/.cache/go-build"}
mkdir -p "$GOCACHE"

. "$ROOT_DIR/scripts/lib/go-version.sh"
require_go_version

for tool in go npm tar zip; do
  command -v "$tool" >/dev/null 2>&1 || {
    echo "$tool is required to create the package." >&2
    exit 1
  }
done

PACKAGE_NAME=${PACKAGE_NAME:-llm-debate-demo}
PACKAGE_VERSION=${PACKAGE_VERSION:-$(date +%Y%m%d-%H%M%S)}
PACKAGE_DIR=${PACKAGE_DIR:-release}
PACKAGE_PATH="$ROOT_DIR/$PACKAGE_DIR/$PACKAGE_NAME-$PACKAGE_VERSION.zip"

if [ "${SKIP_TESTS:-0}" != "1" ]; then
  if [ ! -d frontend/node_modules ]; then
    ./scripts/setup.sh
  fi
  ./scripts/test.sh
fi

echo "Building frontend assets for single-process serving..."
(cd frontend && npm run build)

echo "Checking backend compilation..."
go build -o /tmp/llm-debate-package-check ./cmd/server

mkdir -p "$PACKAGE_DIR"
rm -f "$PACKAGE_PATH"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT
STAGING_DIR="$TMP_DIR/$PACKAGE_NAME"
mkdir -p "$STAGING_DIR"

echo "Staging package..."
COPYFILE_DISABLE=1 tar \
  --exclude './.git' \
  --exclude './.worktrees' \
  --exclude './.superpowers' \
  --exclude './frontend/node_modules' \
  --exclude './node_modules' \
  --exclude './release' \
  --exclude './bin' \
  --exclude './.cache' \
  --exclude './.env' \
  --exclude './.DS_Store' \
  -cf - . | tar -xf - -C "$STAGING_DIR"

echo "Creating $PACKAGE_PATH..."
(cd "$TMP_DIR" && zip -qr "$PACKAGE_PATH" "$PACKAGE_NAME")

echo "Package ready: $PACKAGE_PATH"
