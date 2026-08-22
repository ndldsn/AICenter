# AICenter

AI-powered operations control center for Linux servers, Docker, AI models, and AI agents.

## Architecture

- **Frontend**: React + TypeScript + Vite + Arco Design
- **Backend**: Go + Gin + SQLite/PostgreSQL
- **Real-time**: WebSocket
- **AI**: OpenAI Compatible API + Anthropic + Gemini + DeepSeek + Ollama

## Features

- **Web Terminal** — browser-based PTY shell per server via WebSocket (`GET /ws/terminal?session=<id>`, `POST/GET /api/v1/terminal/sessions`, `DELETE /terminal/sessions/:id`).
- **Batch Operations** — run one command across many servers at once (`POST /api/v1/servers/batch/command`), with per-host timeout, process-tree cleanup, and aggregated results.
- **In-memory caching** — TTL+LRU cache for hot server reads (server list / get-by-id), invalidating on write.
- **Provider rate limiting** — per-provider concurrency limiter (`ai.DefaultProviderConcurrency=4`) guards API credentials and prevents 429 storms.

## API

See `ARCHITECTURE.md` §7 for the full REST + WebSocket reference.

## Quick Start

### Development

```bash
# Terminal 1: Start backend
cd backend
make dev

# Terminal 2: Start frontend
cd frontend
npm install
npm run dev
```

### Docker

```bash
docker compose -f deployments/docker-compose.yml up -d
```

## Project Structure

```
AI_Server_Center/
├── frontend/          # React SPA
├── backend/           # Go API server
├── agent/             # Edge agent for managed servers
├── deployments/       # Docker Compose files
├── migrations/        # Database migrations
└── docs/              # Documentation
```

## License

MIT
