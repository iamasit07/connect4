# Backend Logic & Architecture Deep Dive

This document outlines the core architecture and logic flows governing the backend of the Connect 4 application. The backend is written in **Go 1.24** and heavily utilizes goroutines, channels, and WebSockets to provide a performant, decoupled, and entirely real-time experience.

---

## 🏗️ Core Architectural Patterns

The backend strictly employs layered architecture (Domain-Driven Design inspired):
1. **Domain Layer (`internal/domain`)**: Contains core models like `Game`, `Board`, `Session`, and definitions for `GameEvent` and `ServerMessage`.
2. **Repository Layer (`internal/repository`)**: Interfaces and database implementations for PostgreSQL and Redis caching.
3. **Service Layer (`internal/service`)**: Holds the core business logic (Matchmaking, Bot Engines, Game Sessions, Auth).
4. **Transport Layer (`internal/transport`)**: Handles HTTP and WebSocket connections, parsing incoming requests, and dispatching them into the service layer.

---

## 📡 WebSocket Lifecycle & Event Loop

All real-time actions happen over a persistent WebSocket connection handled in `internal/transport/websocket/handler.go`. 

### Connection Initialization
1. A client connects via the Gin HTTP upgrade endpoint.
2. The initial message **must** be an `init` event containing a valid JWT. The server performs a stateful validation (comparing with PostgreSQL / Redis).
3. We register the user in the `ConnectionManager`, which binds the `int64` userID to the exact `*websocket.Conn`.

### Decoupled Communication
To prevent race conditions during concurrent broadcasts, all game sessions have a dedicated event loop powered by Go channels (`gs.Events`). 
Whenever a game event occurs (e.g., `MakeMove`), the `GameSession` structures a payload into a `domain.GameEvent` and pushes it to its channel. The WebSocket handler continuously listens to this channel in the `ConsumeGameEvents` goroutine and automatically serializes the data to all required clients using `ConnManager.SendMessage()`. 

---

## 🎮 Game Engine & State Management (`game.Service`)

A `GameSession` (`internal/service/game/service.go`) represents a live, active match between two players (or a player and a bot). 

### Turn Logic (`HandleMove`)
When a `make_move` payload reaches the WebSocket handler, it calls `GameSession.HandleMove()`:
1. **Thread safety**: Locked via a `sync.Mutex` on the game session to ensure moves are linearly processed.
2. **Validation**: We verify if the request came from the user who possesses the current turn.
3. **Execution**: The underlying `domain.Game.MakeMove()` evaluates the physics of the 7x6 board, finding the lowest available empty row.
4. **Win Condition**: The game logic runs constant checks (Horizontal, Vertical, Diagonal) to detect a win or a draw. 
5. **Broadcast**: A `move_made` or `game_over` event is broadcasted natively to all participants (including spectators).

### Disconnection & Reconnection
If a WebSocket drops mid-game:
* We trigger `HandleDisconnect()`. If a player hasn't returned within a 60-second grace period (`DisconnectTimer`), the session is automatically forfeited and the opponent wins by abandonment. 
* If the user connects back with their JWT, the `init` process detects they belong to an ongoing session, cancels the `DisconnectTimer`, and executes a fast reconnection broadcasting the latest board state to the client.

---

## 🤖 The Bot Engine (`bot.Service`)

The backend boasts three difficulty levels. Handling bot logic is non-blocking — when `HandleMove` detects a bot turn, it spawns a goroutine with an artificial delay (to simulate thinking) and invokes `HandleBotMove()`.

- **Easy (`bot.Easy`)**: Pulls valid columns randomly with slight heuristic adjustments to block immediate threats.
- **Medium (`bot.Medium`)**: Employs shallow evaluation using positional weighting grids. It assigns points to central columns over edges and calculates simple 2-in-a-row and 3-in-a-row potentials.
- **Hard (`bot.Hard`)**: Implements **Minimax with Alpha-Beta Pruning** to depth 7. By checking countless future move permutations recursively, it maps out the mathematical best outcome while discarding sub-optimal branches to save computing resources.

---

## ⚔️ Matchmaking Queue 

The matchmaking system runs independently in `internal/service/matchmaking/queue.go`.
* A player searches for a session, entering a synchronized queue.
* The matching algorithm loops continually (via a ticker in `Run()`). Once it detects two available human players, it pairs them, extracting them from the queue, creates a new `GameSession`, and injects a `game_start` event down their WebSockets.
* If a human cannot find an opponent after a configurable timeout step (usually 10s), the queue smoothly transitions them into an automated **Bot Match**.

## 💾 Data Persistence
When games conclude naturally or by abandonment, `saveGameAsync` runs in the background. It persists match states (duration, moves, board configurations) back into Supabase PostgreSQL, enabling features like detailed User History and Elo Leaderboard rankings.
