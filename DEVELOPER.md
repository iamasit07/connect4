# Developer Guide

This guide covers local development setup, project architecture, and coding conventions for the Connect 4 project.

---

## Prerequisites

| Tool                         | Version       | Purpose               |
| ---------------------------- | ------------- | --------------------- |
| **Go**                       | 1.24+         | Backend server        |
| **Node.js**                  | 18+           | Frontend toolchain    |
| **PostgreSQL**               | 14+           | Database              |
| **Redis**                    | Latest        | Caching               |
| **Docker** + Docker Compose  | Latest        | Containerized setup   |

---

## Development Setup

### Docker (Recommended)

Docker Compose brings up PostgreSQL, Redis, the Go backend (with [Air](https://github.com/air-verse/air) hot reload), and the React frontend (with Vite HMR).

```bash
git clone https://github.com/iamasit07/connect4.git
cd connect4
cp .env.example .env    # Configure your environment
docker compose up --build
```

### Manual

See [README.md → Getting Started](./README.md#getting-started) for step-by-step backend and frontend commands.

---

## Architecture

### Backend (Go)

The backend follows a layered, DDD-inspired structure inside `backend/internal/`:

```
cmd/api/main.go           → Entry point
internal/config/           → Environment loading, Google OAuth config
internal/domain/           → Core models (Game, Board, Player), events, messages
internal/repository/       → PostgreSQL and Redis data access
internal/service/          → Business logic
  ├── bot/                 → AI engine (easy, medium, hard with minimax)
  ├── cleanup/             → Background session/game garbage collection
  ├── game/                → Game session lifecycle, turn logic
  ├── matchmaking/         → PvP queue + auto bot-match on timeout
  └── session/             → Auth service, JWT validation
internal/transport/        → HTTP and WebSocket handlers
  ├── http/                → REST endpoints (auth, history, OAuth)
  └── websocket/           → WebSocket connection manager + event loop
pkg/                       → Shared packages (JWT, password hashing, cookies)
```

All real-time game communication flows through WebSocket. Game sessions use dedicated Go channels (`gs.Events`) to decouple message handling from broadcasting and prevent race conditions.

### Frontend (React + TypeScript)

The frontend uses a feature-based folder structure:

```
src/
├── components/            → Reusable UI (layout, header, shadcn/ui primitives)
├── features/
│   ├── auth/              → Login, signup, OAuth, auth store
│   └── game/              → Board, game store, WebSocket hook, rematch
├── hooks/                 → Custom hooks (queries, mobile detection)
├── lib/                   → Axios instance, config constants, utilities
├── pages/                 → Route-level page components
└── stores/                → Zustand global state (theme)
```

---

## Coding Conventions

### State Management

- **Zustand** for global state (`AuthStore`, `GameStore`, `UIStore`) — never prop-drill what belongs in a store
- **React Query** for server state (leaderboard, history, user profile)
- Keep WebSocket interactions in hooks (`useGameSocket`) and pure UI state in Zustand

### Real-time Communication

All game actions flow through WebSocket. Key files:
- **Backend:** `internal/transport/websocket/handler.go` (message routing)
- **Frontend:** `features/game/hooks/useGameSocket.ts` (event handling)

### Styling

- **Tailwind CSS 4** utility classes for all styling
- **shadcn/ui** for complex component primitives (`src/components/ui/`)
- Custom CSS utilities for game-specific visuals are in `index.css` (`disk-shadow-red`, `win-glow`, `board-3d`)

### Bot AI

The minimax engine lives in `backend/internal/service/bot/`. If tuning AI behavior:
- **Easy:** Random moves with basic win/block detection
- **Medium:** Shallow positional evaluation
- **Hard:** Depth-7 minimax with alpha-beta pruning and positional weight matrices

---

## Debugging

| Scenario              | How to debug                                                  |
| --------------------- | ------------------------------------------------------------- |
| Backend hot-reload    | Docker Compose uses Air — check `backend/.air.toml`           |
| WebSocket events      | Go server logs are prefixed with `[WS]`                       |
| Frontend HMR issues   | Restart Vite dev server to clear Tailwind config cache        |
| Auth flow problems    | Check browser DevTools → Application → Cookies for JWT tokens |
| Game state sync       | Add `console.log(useGameStore.getState())` in browser console |

---

## Useful Commands

```bash
# Backend
cd backend && go run ./cmd/api       # Run server
cd backend && go test ./...          # Run tests

# Frontend
cd frontend && npm run dev           # Dev server with HMR
cd frontend && npm run build         # Production build
cd frontend && npm run lint          # ESLint

# Docker
docker compose up                    # Full dev stack
docker compose -f docker-compose.prod.yml up   # Production
```
