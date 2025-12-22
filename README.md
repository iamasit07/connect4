# 4-in-a-Row Multiplayer Game

A real-time multiplayer Connect 4 game with WebSocket-based gameplay, intelligent bot opponent, and persistent user accounts. Players can compete against each other or challenge an AI bot with 3-step lookahead strategy.

🎮 **[Live Demo](https://4-in-a-row.iamasit07.me)** | 🚀 **[Backend API](https://four-in-a-row-y7w6.onrender.com)**

---

## 🎯 Features

- **Real-time Multiplayer** - WebSocket-powered live gameplay with instant move updates
- **Intelligent Bot Opponent** - Challenge an AI with 3-step lookahead algorithm and strategic evaluation
- **User Authentication** - Secure JWT-based authentication system
- **Leaderboard System** - Track wins and compete with other players
- **Reconnection Support** - 30-second grace period to reconnect to ongoing games
- **Persistent Game State** - Resume games after disconnect without losing progress
- **Move Validation** - Server-side validation ensures fair play
- **Win Detection** - Automatic detection of winning combinations in all directions
- **Responsive UI** - Clean, mobile-friendly interface built with React and TailwindCSS

---

## 🛠️ Tech Stack

### Backend

- **Language**: Go 1.21+
- **WebSocket**: Gorilla WebSocket
- **Database**: PostgreSQL 14+ with `lib/pq` driver
- **Authentication**: JWT tokens with bcrypt password hashing
- **Environment**: `godotenv` for configuration

### Frontend

- **Framework**: React 19 with TypeScript
- **Styling**: TailwindCSS 4.1
- **Routing**: React Router v7
- **Build Tool**: Vite 7
- **Bundler**: TypeScript 5.9

### Deployment

- **Frontend**: Vercel
- **Backend**: Render
- **Database**: Render PostgreSQL

---

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- **Go** 1.21 or higher ([Download](https://golang.org/dl/))
- **Node.js** 18 or higher ([Download](https://nodejs.org/))
- **PostgreSQL** 14 or higher ([Download](https://www.postgresql.org/download/))
- **Git** for cloning the repository

---

## 🚀 Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/iamasit07/4-in-a-row.git
cd 4-in-a-row
```

### 2. Backend Setup

#### Install Go Dependencies

```bash
cd backend
go mod download
```

#### Database Setup

Create a PostgreSQL database and apply the schema:

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE four_in_a_row;
\c four_in_a_row

# Apply schema (from psql or terminal)
\i db/schema.sql

# Or from terminal:
psql -U postgres -d four_in_a_row -f db/schema.sql
```

#### Configure Environment Variables

Create a `.env` file in the `backend/` directory:

```bash
# Backend .env
DB_URI=postgresql://username:password@localhost:5432/four_in_a_row
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080

# Database Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME_MINUTES=5

# Game Configuration
RECONNECT_TIMEOUT_SECONDS=30
BOT_MATCHMAKING_TIMEOUT_SECONDS=10
BOT_USERNAME=BOT
BOT_TOKEN=tkn_bot_default

# CORS
FRONTEND_URL=http://localhost:5173
```

### 3. Frontend Setup

#### Install Node Dependencies

```bash
cd ../frontend
npm install
```

#### Configure Environment Variables

Create a `.env` file in the `frontend/` directory:

```bash
# Frontend .env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
VITE_BACKEND_URL=http://localhost:8080
```

---

## ▶️ Running the Application

### Start Backend Server

```bash
cd backend
go run main.go
```

The backend server will start on `http://localhost:8080`

### Start Frontend Development Server

In a new terminal:

```bash
cd frontend
npm run dev
```

The frontend will be available at `http://localhost:5173`

### Access the Application

Open your browser and navigate to:

```
http://localhost:5173
```

---

## 🔐 Environment Variables Reference

### Backend Variables

| Variable                          | Description                      | Default                 | Required |
| --------------------------------- | -------------------------------- | ----------------------- | -------- |
| `DB_URI`                          | PostgreSQL connection string     | -                       | ✅ Yes   |
| `JWT_SECRET`                      | Secret key for JWT token signing | -                       | ✅ Yes   |
| `PORT`                            | Server port number               | `8080`                  | ❌ No    |
| `DB_MAX_OPEN_CONNS`               | Max open database connections    | `25`                    | ❌ No    |
| `DB_MAX_IDLE_CONNS`               | Max idle database connections    | `25`                    | ❌ No    |
| `DB_CONN_MAX_LIFETIME_MINUTES`    | Connection max lifetime          | `5`                     | ❌ No    |
| `RECONNECT_TIMEOUT_SECONDS`       | Grace period for reconnection    | `30`                    | ❌ No    |
| `BOT_MATCHMAKING_TIMEOUT_SECONDS` | Bot matchmaking timeout          | `10`                    | ❌ No    |
| `BOT_USERNAME`                    | Display name for bot             | `BOT`                   | ❌ No    |
| `BOT_TOKEN`                       | Special token for bot games      | `tkn_bot_default`       | ❌ No    |
| `FRONTEND_URL`                    | Frontend URL for CORS            | `http://localhost:5173` | ❌ No    |

### Frontend Variables

| Variable           | Description           | Default                 | Required |
| ------------------ | --------------------- | ----------------------- | -------- |
| `VITE_API_URL`     | Backend HTTP API URL  | `http://localhost:8080` | ❌ No    |
| `VITE_WS_URL`      | Backend WebSocket URL | `ws://localhost:8080`   | ❌ No    |
| `VITE_BACKEND_URL` | Backend base URL      | `http://localhost:8080` | ❌ No    |

---

## 📁 Project Structure

```
4-in-a-row/
├── backend/
│   ├── bot/               # AI bot logic with 3-step lookahead
│   ├── config/            # Configuration and environment loading
│   ├── db/                # Database connection and queries
│   ├── game/              # Game logic and win detection
│   ├── handlers/          # HTTP request handlers (auth)
│   ├── middlewares/       # CORS and middleware
│   ├── models/            # Data models and types
│   ├── server/            # Game session management
│   ├── utils/             # Helper utilities (JWT, board utils)
│   ├── websocket/         # WebSocket connection handling
│   └── main.go            # Entry point
├── frontend/
│   ├── src/
│   │   ├── components/    # React components (Board, Notifications)
│   │   ├── hooks/         # Custom hooks (useAuth, useWebSocket)
│   │   ├── pages/         # Page components (Game, Landing, Login)
│   │   ├── types/         # TypeScript type definitions
│   │   └── utils/         # API utilities and helpers
│   └── public/            # Static assets
└── README.md
```

---

## 🎮 Game Rules

**Connect 4** (4-in-a-Row) is a two-player strategy game:

1. Players take turns dropping colored discs into a 7-column, 6-row vertical grid
2. Discs fall to the lowest available position in the chosen column
3. The first player to form a horizontal, vertical, or diagonal line of **four discs** wins
4. If the board fills up with no winner, the game ends in a draw

---

## 🤖 Bot AI Strategy

The bot uses a **3-step lookahead algorithm** with strategic evaluation:

1. **Immediate Win/Block** - Prioritizes winning moves or blocking opponent wins
2. **Threat Creation** - Creates positions with multiple winning opportunities (unblockable threats)
3. **Lookahead Simulation** - Evaluates opponent's best response and counter-strategies
4. **Positional Scoring** - Favors center columns and connected pieces
5. **Gravity Awareness** - Only evaluates playable positions respecting game physics

---

## 🔧 API Endpoints

### Authentication

- `POST /auth/signup` - Create new user account
- `POST /auth/login` - User login (returns JWT token)

### Game

- `GET /leaderboard` - Fetch top players by wins
- `WS /ws` - WebSocket connection for gameplay

### WebSocket Messages

- `matchmaking` - Join matchmaking queue
- `bot_game` - Start game against bot
- `move` - Make a move
- `reconnect` - Rejoin disconnected game

---

## 📝 License

This project is open source and available under the [MIT License](LICENSE).

---

## 👨‍💻 Author

**Asit Kumar**

- GitHub: [@iamasit07](https://github.com/iamasit07)
- Live Demo: [4-in-a-row.iamasit07.me](https://4-in-a-row.iamasit07.me)

---

## 🙏 Acknowledgments

- Built with ❤️ using Go and React
- Inspired by the classic Connect 4 game
- WebSocket implementation powered by Gorilla WebSocket
