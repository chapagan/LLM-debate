# LLM Debate

LLM Debate is a small demo of a live multi-agent argument. You give it a topic; two opposing personas (a bold visionary and a skeptical CFO) take turns streaming their case over WebSocket into a split UI. The point is to show how token streaming, turn orchestration, and a simple frontend feel when agents disagree in real time.

By default it runs offline with mock replies so you can try the product without API keys. Switch to OpenAI, Claude, or an OpenAI-compatible Cursor endpoint when you want real models.

## Requirements

- Go 1.26.4+
- Node.js 20+ and npm (for first-time frontend build)

## Run locally

```bash
make setup
make run
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080).

That starts in mock mode. Enter a topic and watch the debate stream.

## Use a real model

```bash
cp .env.example .env
```

Edit `.env` and set one provider:

**OpenAI**

```bash
AI_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-5.5
```

**Claude** (native Anthropic Messages API)

```bash
AI_PROVIDER=claude
ANTHROPIC_API_KEY=...
ANTHROPIC_MODEL=claude-sonnet-4-5
```

`ANTHROPIC_MODEL` is required — there is no default.

**Cursor** (OpenAI-compatible `/v1` base URL, e.g. a local proxy)

```bash
AI_PROVIDER=cursor
CURSOR_API_KEY=...
CURSOR_BASE_URL=http://127.0.0.1:9000/v1
CURSOR_MODEL=composer-2.5
```

`CURSOR_MODEL` is required when using Cursor — there is no default.

Then:

```bash
./scripts/run.sh
```

## Useful commands

```bash
make test      # Go + frontend tests
make package   # zip demo for sharing
```

Shell overrides win over `.env`, e.g. `AI_PROVIDER=mock ./scripts/run.sh`. The server listens on `127.0.0.1:8080` by default (`ADDR` to change).
