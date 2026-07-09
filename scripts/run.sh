#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

export GOCACHE=${GOCACHE:-"$ROOT_DIR/.cache/go-build"}
mkdir -p "$GOCACHE"

HAD_ADDR=${ADDR+x}
EXISTING_ADDR=${ADDR-}
HAD_OPENAI_API_KEY=${OPENAI_API_KEY+x}
EXISTING_OPENAI_API_KEY=${OPENAI_API_KEY-}
HAD_OPENAI_MODEL=${OPENAI_MODEL+x}
EXISTING_OPENAI_MODEL=${OPENAI_MODEL-}

if [ -f .env ]; then
  set -a
  . ./.env
  set +a
fi

if [ "$HAD_ADDR" ]; then
  export ADDR="$EXISTING_ADDR"
fi
if [ "$HAD_OPENAI_API_KEY" ]; then
  export OPENAI_API_KEY="$EXISTING_OPENAI_API_KEY"
fi
if [ "$HAD_OPENAI_MODEL" ]; then
  export OPENAI_MODEL="$EXISTING_OPENAI_MODEL"
fi

. "$ROOT_DIR/scripts/lib/go-version.sh"
require_go_version

if [ ! -f frontend/dist/index.html ]; then
  command -v npm >/dev/null 2>&1 || {
    echo "frontend/dist is missing and npm is not available to build it." >&2
    echo "Install Node.js 20 or newer, then rerun this script." >&2
    exit 1
  }

  echo "frontend/dist not found; building the React app..."
  if [ ! -d frontend/node_modules ]; then
    (cd frontend && npm ci)
  fi
  (cd frontend && npm run build)
fi

ADDR_VALUE=${ADDR:-:8080}
case "$ADDR_VALUE" in
  :*) URL="http://localhost$ADDR_VALUE" ;;
  0.0.0.0:*) URL="http://localhost:${ADDR_VALUE#0.0.0.0:}" ;;
  *) URL="http://$ADDR_VALUE" ;;
esac

echo "Starting LLM Debate at $URL"
if [ -z "${OPENAI_API_KEY:-}" ]; then
  echo "OPENAI_API_KEY is not set; using deterministic mock AI responses."
fi

exec go run ./cmd/server
