# AICenter

AI-powered operations control center for Linux servers, Docker, AI models, and AI agents.

## Architecture

- **Frontend**: React + TypeScript + Vite + Arco Design
- **Backend**: Go + Gin + SQLite/PostgreSQL
- **Real-time**: WebSocket
- **AI**: OpenAI Compatible API + Anthropic + Gemini + DeepSeek + Ollama

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
