# Connect 4 — Real-Time Multiplayer Game

> **Created by [Asit Upadhyay](https://github.com/iamasit07)**

🎮 **[Play Now](https://connect4.iamasit07.me)** | 🚀 **[Backend API](https://connect4.iamasit07.me)**

---

## 🧩 What is this?

Connect 4 is a full-stack, real-time multiplayer web game where two players take turns dropping colored discs into a 7×6 vertical grid, racing to be the first to connect four in a row — horizontally, vertically, or diagonally. Built with a **Go** backend and a **React/TypeScript** frontend, the platform supports live PvP matchmaking over WebSockets, three tiers of AI bot opponents (easy, medium, hard) powered by minimax with alpha-beta pruning, JWT + Google OAuth authentication, Elo-based rating and leaderboard tracking, game history, spectator counts, rematch requests, and a 30-second reconnection window — all deployed as a production-ready monolith on Render with Supabase PostgreSQL and Redis caching.

---

## 📚 Documentation

- [Developer Guide](./DEVELOPER.md) - Setup, architecture, and contribution guidelines.
- [Backend Logic Deep Dive](./BACKEND_LOGIC.md) - Detailed breakdown of the game engine, WebSockets, and bots.

---

## 📸 Screenshots

![Landing Screen](assets/Landing.png)
![Dashboard Screen](assets/dashboard.png)
![Lobby Screen](assets/lobby.png)
![Gameplay](assets/game.png)
![Leaderboard](assets/leaderboard.png)

---

## ✨ Features

- **🎮 Play vs Players** — Real-time PvP matchmaking with automatic opponent pairing via WebSocket
- **🤖 Play vs Bot** — Three difficulty levels: Easy (random with basic blocking), Medium (threat evaluation), Hard (depth-7 minimax with alpha-beta pruning)
- **🔄 Rematch System** — Request and accept rematches after a game ends, with a 10-second countdown overlay
- **🔐 Authentication** — Sign up with email/password or Google OAuth; JWT-based stateless sessions with single-device enforcement
- **📊 Leaderboard** — Elo-based competitive ranking system updated after every match
- **📜 Game History** — Browse your past matches with results, move counts, and timestamps
- **👤 Player Profiles** — View your rating, win/loss/draw stats, and account details
- **📱 Responsive Design** — Fully playable on mobile, tablet, and desktop
- **🌙 Dark/Light/System Theme** — Automatic theme detection with manual toggle

---

## 🛠️ Tech Stack

### Backend

| Component   | Technology                      | Purpose                                       |
| ----------- | ------------------------------- | --------------------------------------------- |
| Language    | **Go 1.24**                     | High-concurrency game server                  |
| WebSocket   | **Gorilla WebSocket**           | Real-time bidirectional communication         |
| Database    | **PostgreSQL** (Supabase)       | Persistent storage for users, games, sessions |
| Cache       | **Redis**                       | Session caching and fast lookups              |
| Auth        | **JWT + bcrypt + Google OAuth** | Stateless authentication with OAuth2 support  |
| Concurrency | **Goroutines + Channels**       | Lightweight concurrent game sessions          |

### Frontend

| Component     | Technology                           | Purpose                                 |
| ------------- | ------------------------------------ | --------------------------------------- |
| Framework     | **React 18** + **TypeScript**        | Type-safe UI components                 |
| Styling       | **Tailwind CSS** + **shadcn/ui**     | Responsive design system                |
| State         | **Zustand**                          | Lightweight global state management     |
| Animation     | **Framer Motion**                    | Smooth transitions and game animations  |
| Data Fetching | **TanStack React Query** + **Axios** | Server state management and HTTP client |
| Bundler       | **Vite**                             | Fast development and optimized builds   |

### Infrastructure

| Component        | Technology                      | Purpose                                            |
| ---------------- | ------------------------------- | -------------------------------------------------- |
| Hosting          | **Render**                      | Monolith deployment (Go serves static React build) |
| Database         | **Supabase PostgreSQL**         | Managed Postgres with transaction-mode pooling     |
| Containerization | **Docker** + **Docker Compose** | Local dev and production builds                    |

---

## 📁 Project Structure

```
connect4/
├── backend/
│   ├── cmd/api/                  # Application entrypoint
│   │   └── main.go
│   ├── internal/
│   │   ├── config/               # App config + Google OAuth setup
│   │   ├── domain/               # Core types: Board, Game, Rules, Messages
│   │   ├── repository/
│   │   │   ├── postgres/         # User, Game, Session DB repositories
│   │   │   └── redis/            # Redis cache client
│   │   ├── service/
│   │   │   ├── bot/              # AI engine: easy, medium, hard (minimax)
│   │   │   ├── cleanup/          # Background session/game cleanup worker
│   │   │   ├── game/             # Game logic, session management
│   │   │   ├── matchmaking/      # PvP queue + bot matching
│   │   │   └── session/          # Auth service, JWT validation
│   │   └── transport/
│   │       ├── http/             # REST handlers: auth, history, OAuth
│   │       └── websocket/        # WebSocket handler + connection manager
│   ├── pkg/                      # Shared packages: JWT, passwords, cookies
│   └── script/migration/         # SQL schema
├── frontend/
│   └── src/
│       ├── components/           # Shared UI (layout, header, shadcn)
│       ├── features/
│       │   ├── auth/             # Login, signup, OAuth, auth store
│       │   └── game/             # Board, game store, WebSocket manager,
│       │                         #   matchmaking, rematch, bot loading
│       ├── hooks/                # Custom React hooks
│       ├── lib/                  # Axios, config, utils
│       ├── pages/                # Dashboard, Leaderboard, History, Profile
│       └── stores/               # UI store (theme)
├── docker-compose.yml            # Dev environment (Vite HMR + Air hot reload)
├── docker-compose.prod.yml       # Production build (Nginx + compiled Go binary)
├── Dockerfile                    # Monolith: frontend build → Go binary → Alpine
└── render.yaml                   # Render deployment config
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.24+ — [Download](https://golang.org/dl/)
- **Node.js** 18+ — [Download](https://nodejs.org/)
- **PostgreSQL** 14+ — [Download](https://www.postgresql.org/download/)
- **Redis** (optional) — [Download](https://redis.io/download/)

### Option 1: Docker (Recommended)

```bash
git clone https://github.com/iamasit07/connect4.git
cd connect4

# Copy and configure environment variables
cp .env.example .env  # Edit with your database URL, JWT secret, etc.

# Start development environment (with hot reload)
docker compose up

# Or production build
docker compose -f docker-compose.prod.yml up
```

### Option 2: Manual Setup

#### Backend

```bash
cd backend
go mod download

# Create .env in backend/ (or project root)
cat > .env << EOF
DATABASE_URI=postgresql://user:password@localhost:5432/four_in_a_row
JWT_SECRET=your-secret-key
PORT=8080
REDIS_URL=redis://localhost:6379
FRONTEND_URL=http://localhost:5173
ALLOWED_ORIGINS=http://localhost:5173
EOF

go run ./cmd/api
# Server starts on http://localhost:8080
# Database tables are auto-migrated on startup
```

#### Frontend

```bash
cd frontend
npm install

# Create .env in frontend/
cat > .env << EOF
VITE_BACKEND_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
EOF

npm run dev
# App available at http://localhost:5173
```

---

## 🔐 Environment Variables

### Backend

| Variable               | Description                   | Required |
| ---------------------- | ----------------------------- | -------- |
| `DATABASE_URI`         | PostgreSQL connection string  | ✅       |
| `JWT_SECRET`           | Secret for signing JWT tokens | ✅       |
| `PORT`                 | Server port (default: `8080`) | ❌       |
| `REDIS_URL`            | Redis connection URL          | ❌       |
| `FRONTEND_URL`         | Frontend origin for cookies   | ❌       |
| `ALLOWED_ORIGINS`      | CORS allowed origins          | ❌       |
| `GOOGLE_CLIENT_ID`     | Google OAuth client ID        | ❌       |
| `GOOGLE_CLIENT_SECRET` | Google OAuth secret           | ❌       |
| `GOOGLE_REDIRECT_URL`  | OAuth callback URL            | ❌       |

### Frontend

| Variable           | Description           | Required |
| ------------------ | --------------------- | -------- |
| `VITE_BACKEND_URL` | Backend HTTP base URL | ❌       |
| `VITE_WS_URL`      | Backend WebSocket URL | ❌       |

---

## 🎮 How It Works

### Game Flow

```
1. Player signs up / logs in (JWT issued, stored in httpOnly cookie)
2. Player clicks "Play Online" or "Play vs Bot"
3. WebSocket connection established → JWT sent for auth
4. Matchmaking:
   - PvP: Queued until opponent found (10s timeout → auto bot match)
   - Bot: Immediate game start with selected difficulty
5. Players take turns dropping discs (validated server-side)
6. Game ends → winner determined → ratings updated → stats saved
7. Rematch available (PvP: request/accept overlay; Bot: instant replay)
```

### Bot AI

| Difficulty | Strategy                                             | Depth |
| ---------- | ---------------------------------------------------- | ----- |
| **Easy**   | Random valid moves with basic win/block detection    | 1     |
| **Medium** | Threat evaluation + positional scoring               | 3     |
| **Hard**   | Minimax with alpha-beta pruning + positional weights | 7     |

### WebSocket Protocol

**Client → Server:**

```json
{"type": "init", "jwt": "..."}
{"type": "find_match", "difficulty": ""}          // PvP
{"type": "find_match", "difficulty": "hard"}      // Bot
{"type": "make_move", "column": 3}
{"type": "abandon_game"}
{"type": "request_rematch"}
{"type": "rematch_response", "rematchResponse": "accept"}
```

**Server → Client:**

```json
{"type": "game_start", "gameId": "...", "opponent": "Player2", "yourPlayer": 1}
{"type": "move_made", "column": 3, "row": 5, "player": 1, "board": [...], "nextTurn": 2}
{"type": "game_over", "winner": "Player1", "reason": "connect4", "allowRematch": true}
{"type": "rematch_request", "rematchRequester": "Player2", "rematchTimeout": 10}
{"type": "error", "message": "Not your turn"}
```

---

## 🗄️ Database Schema

```sql
players         — id, username, email, google_id, password_hash, rating, games_played/won/drawn
game            — game_id, player1/2_id, winner, reason, total_moves, duration, board_state (JSONB)
user_sessions   — session_id, user_id, device_info, ip_address, is_active (single-device enforced)
```

---

## 📞 Contact

- **GitHub**: [@iamasit07](https://github.com/iamasit07)
- **Email**: [asit.upadhyay793@gmail.com](mailto:asit.upadhyay793@gmail.com)
