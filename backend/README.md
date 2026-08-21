# AICenter Backend

Go + Gin + SQLite/PostgreSQL + WebSocket

## Quick Start

```bash
# Install dependencies
make deps

# Run in development mode
make dev

# Build binary
make build

# Run tests
make test
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| WS | /ws | WebSocket endpoint |
| POST | /api/v1/auth/login | Login |
| GET | /api/v1/servers | List servers |
| GET | /api/v1/docker/containers | List containers |
| GET | /api/v1/agents | List agents |

## Configuration

Copy `.env.example` to `.env` and adjust values.
