# Developer Guide

Welcome to the Connect 4 project! This guide is designed to help you set up the development environment, understand the architecture, and start contributing to the project.

## 🛠️ Prerequisites

Before you begin, ensure you have the following installed on your machine:
- **Go** 1.24+ (For the backend)
- **Node.js** 18+ (For the frontend)
- **PostgreSQL** 14+ (Local database or Supabase)
- **Redis** (Used for rate limiting and cache)
- **Docker** & **Docker Compose** (For easier environment setup)

---

## 🚀 Setting Up the Development Environment

### Using Docker (Recommended)

The easiest way to get everything running is via Docker Compose, which brings up PostgreSQL, Redis, the Go backend (with hot reload via Air), and the React frontend (with Vite HMR).

```bash
# 1. Clone the repository
git clone https://github.com/iamasit07/connect4.git
cd connect4

# 2. Setup your environment variables
cp .env.example .env

# 3. Start the entire stack
docker compose up --build
```

### Manual Setup

If you prefer to run services manually on your host machine:

#### 1. Backend

The backend is a monolithic Go service built with standard library tools and Gorilla WebSocket.

```bash
cd backend
go mod download

# Create a local .env file
cat > .env << EOF
DATABASE_URI=postgresql://user:password@localhost:5432/four_in_a_row
JWT_SECRET=your-secret-key
PORT=8080
REDIS_URL=redis://localhost:6379
FRONTEND_URL=http://localhost:5173
ALLOWED_ORIGINS=http://localhost:5173
EOF

# Run the server
go run ./cmd/api
```

#### 2. Frontend

The frontend is a React 18 application built with TypeScript and Vite. It heavily utilizes Tailwind CSS and shadcn/ui.

```bash
cd frontend
npm install

# Create a local .env file
cat > .env << EOF
VITE_BACKEND_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
EOF

# Start the Vite development server
npm run dev
```

---

## 🗂️ Project Structure

### Backend Architecture

The backend follows a Domain-Driven Design (DDD) inspired structure within the `internal` package:
- `cmd/api/main.go`: Entry point for the Go server.
- `internal/config`: Configuration and environment loading.
- `internal/domain`: Core business models (Game, Board, Player), events, and message definitions.
- `internal/repository`: Implementations for PostgreSQL and Redis interactions.
- `internal/service`: Business logic for game sessions, matchmaking queues, bot opponents, and authentication.
- `internal/transport`: HTTP REST endpoints and the WebSocket connection handler.

### Frontend Architecture

The React application uses a feature-based structure for better scalability:
- `src/components/`: Reusable, generic UI components (mostly shadcn/ui).
- `src/features/`: Domain-specific logic grouped by feature (e.g., `auth/`, `game/`).
- `src/hooks/`: Reusable React Hooks like network state or websocket handlers.
- `src/pages/`: Main page components mapping to routes.
- `src/stores/`: Zustand global state slices.
- `src/lib/`: Utilities, Axios configurations.

---

## 💡 Best Practices & Guidelines

1. **State Management**:
   The frontend relies heavily on `zustand` for global state (e.g., separating `AuthStore` and `GameStore`) rather than complex prop drilling or big context blobs. When modifying game logic, keep the network interactions in hooks (like `useGameSocket`) and pure UI state in Zustand.

2. **Real-time Communication**:
   The vast majority of the real action happens via WebSocket. Pay close attention to `backend/internal/transport/websocket/handler.go` for message routing, and `frontend/src/features/game/hooks/useGameSocket.ts` for how the frontend reacts to server events.

3. **Styling**:
   Use standard **Tailwind CSS** utility classes. For complex components, refer to our `shadcn/ui` custom implementations inside `src/components/ui/`.

4. **Bot AI Adjustments**:
   If you wish to optimize or tweak the AI, check `backend/internal/service/bot/` where the minimax algorithm and alpha-beta pruning logic live. It is deeply connected to board evaluations.

## 🐛 Debugging

- **Backend Hot-Reload**: When running Docker Compose, Go code is continuously watched using Air. Check `backend/.air.toml` for the exact configuration.
- **WebSocket Logs**: The Go server prints generous logs prefixed with `[WS]` for connection lifecycles and event decoupling.
- **Vite HMR**: Changes inside `frontend/src/` will instantly update locally. If you run into Tailwind config cache issues, restarting Vite usually clears it.

Happy coding!
