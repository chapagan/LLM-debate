#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

export GOCACHE=${GOCACHE:-"$ROOT_DIR/.cache/go-build"}
mkdir -p "$GOCACHE"

HAD_ADDR=${ADDR+x}
EXISTING_ADDR=${ADDR-}
HAD_AI_PROVIDER=${AI_PROVIDER+x}
EXISTING_AI_PROVIDER=${AI_PROVIDER-}
HAD_OPENAI_API_KEY=${OPENAI_API_KEY+x}
EXISTING_OPENAI_API_KEY=${OPENAI_API_KEY-}
HAD_OPENAI_MODEL=${OPENAI_MODEL+x}
EXISTING_OPENAI_MODEL=${OPENAI_MODEL-}
HAD_CURSOR_API_KEY=${CURSOR_API_KEY+x}
EXISTING_CURSOR_API_KEY=${CURSOR_API_KEY-}
HAD_CURSOR_BASE_URL=${CURSOR_BASE_URL+x}
EXISTING_CURSOR_BASE_URL=${CURSOR_BASE_URL-}
HAD_CURSOR_MODEL=${CURSOR_MODEL+x}
EXISTING_CURSOR_MODEL=${CURSOR_MODEL-}

if [ -f .env ]; then
  set -a
  . ./.env
  set +a
fi

if [ "$HAD_ADDR" ]; then
  export ADDR="$EXISTING_ADDR"
fi
if [ "$HAD_AI_PROVIDER" ]; then
  export AI_PROVIDER="$EXISTING_AI_PROVIDER"
fi
if [ "$HAD_OPENAI_API_KEY" ]; then
  export OPENAI_API_KEY="$EXISTING_OPENAI_API_KEY"
fi
if [ "$HAD_OPENAI_MODEL" ]; then
  export OPENAI_MODEL="$EXISTING_OPENAI_MODEL"
fi
if [ "$HAD_CURSOR_API_KEY" ]; then
  export CURSOR_API_KEY="$EXISTING_CURSOR_API_KEY"
fi
if [ "$HAD_CURSOR_BASE_URL" ]; then
  export CURSOR_BASE_URL="$EXISTING_CURSOR_BASE_URL"
fi
if [ "$HAD_CURSOR_MODEL" ]; then
  export CURSOR_MODEL="$EXISTING_CURSOR_MODEL"
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

ADDR_VALUE=${ADDR:-127.0.0.1:8080}
case "$ADDR_VALUE" in
  :*) URL="http://localhost$ADDR_VALUE" ;;
  0.0.0.0:*) URL="http://localhost:${ADDR_VALUE#0.0.0.0:}" ;;
  127.0.0.1:*) URL="http://$ADDR_VALUE" ;;
  *) URL="http://$ADDR_VALUE" ;;
esac

PROVIDER=${AI_PROVIDER:-mock}
echo "Starting LLM Debate at $URL (AI_PROVIDER=$PROVIDER)"
case "$PROVIDER" in
  mock|"")
    echo "Using deterministic mock AI responses."
    ;;
  openai)
    echo "Using OpenAI streaming."
    ;;
  cursor)
    echo "Using Cursor OpenAI-compatible streaming."
    ;;
esac

exec go run ./cmd/server
