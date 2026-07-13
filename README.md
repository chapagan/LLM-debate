# LLM Debate

LLM Debate is a real-time multi-agent debate simulator. A Go backend orchestrates two alternating AI personas over WebSocket, and a React frontend renders live streamed tokens in a split arena.

## Requirements

- Go 1.26.4 or newer
- Node.js 20 or newer and npm for setup, tests, frontend rebuilds, and packaging
- Optional: OpenAI or Cursor credentials when `AI_PROVIDER` is `openai` or `cursor`

For a packaged zip that already contains `frontend/dist`, the mock demo path only needs Go.

Download Go from the official release page: https://go.dev/dl/.

## Quick Start From A Zip
Run:

```bash
cd llm-debate-demo
./scripts/run.sh
```

Open:

```text
http://localhost:8080
```

`scripts/run.sh` starts the Go server and serves the built React app from `frontend/dist`. If `frontend/dist` is missing, it installs frontend dependencies and builds it first.

## Quick Start From Source

```bash
make setup
make run
```

Open:

```text
http://localhost:8080
```

Mock mode is the default when `AI_PROVIDER` is unset or set to `mock`.

## Choose An AI Provider

Copy the example environment file:

```bash
cp .env.example .env
```

Set `AI_PROVIDER` to one of:

- `mock` (default): deterministic local responses, no API key
- `openai`: live OpenAI Responses streaming
- `cursor`: live streaming via an OpenAI-compatible Cursor base URL

### OpenAI

```bash
AI_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-5.5
```

### Cursor

Point at an OpenAI-compatible `/v1` endpoint (for example a local Cursor proxy):

```bash
AI_PROVIDER=cursor
CURSOR_API_KEY=...
CURSOR_BASE_URL=http://127.0.0.1:9000/v1
CURSOR_MODEL=composer-2.5
```

Then start the server:

```bash
./scripts/run.sh
```

Values passed directly in the shell, such as `ADDR=127.0.0.1:8090 AI_PROVIDER=openai ./scripts/run.sh`, override `.env`.

Available environment variables:

- `ADDR`: backend listen address, default `127.0.0.1:8080` (loopback only; use `ADDR=:8080` to bind all interfaces)
- `AI_PROVIDER`: `mock` (default), `openai`, or `cursor`
- `OPENAI_API_KEY`: required when `AI_PROVIDER=openai`
- `OPENAI_MODEL`: OpenAI model, default `gpt-5.5`
- `CURSOR_API_KEY`: required when `AI_PROVIDER=cursor`
- `CURSOR_BASE_URL`: required OpenAI-compatible `/v1` root when `AI_PROVIDER=cursor`
- `CURSOR_MODEL`: Cursor model, default `composer-2.5`

## Useful Commands

```bash
make setup      # install Go and frontend dependencies
make test       # run Go tests and frontend tests
make build      # build frontend assets and backend binary
make run        # start the single-process local demo
make package    # create release/llm-debate-demo-<timestamp>.zip
make clean      # remove generated build and package artifacts
```

The same commands are available directly under `scripts/`:

```bash
./scripts/setup.sh
./scripts/test.sh
./scripts/run.sh
./scripts/package.sh
```

## Create A Zip For Testers

From the project root:

```bash
./scripts/package.sh
```

The script runs tests, builds `frontend/dist`, checks backend compilation, excludes local-only files such as `.git`, `node_modules`, `.env`, and writes:

```text
release/llm-debate-demo-<timestamp>.zip
```

To skip tests while iterating on the package script:

```bash
SKIP_TESTS=1 ./scripts/package.sh
```

## Development Mode

Use this mode if you are actively editing the frontend and want Vite hot reload.

Terminal 1:

```bash
go run ./cmd/server
```

Terminal 2:

```bash
cd frontend
npm run dev
```

Open the Vite URL, usually:

```text
http://127.0.0.1:5173
```

## Troubleshooting

If port `8080` is already in use, choose another port:

```bash
ADDR=127.0.0.1:8090 ./scripts/run.sh
```

To expose the demo beyond this machine (not recommended with live AI providers):

```bash
ADDR=:8080 ./scripts/run.sh
```

If frontend dependencies are missing or stale:

```bash
make setup
```

If you want to force a fresh frontend build:

```bash
rm -rf frontend/dist
./scripts/run.sh
```

If live provider calls fail, switch back to mock mode:

```bash
AI_PROVIDER=mock ./scripts/run.sh
```

Or remove the local `.env` file:

```bash
rm .env
./scripts/run.sh
```
