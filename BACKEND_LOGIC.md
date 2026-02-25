# Backend Logic & Architecture

Deep dive into the core architecture and logic flows powering the Connect 4 backend. Written in **Go 1.24**, it uses goroutines, channels, and WebSockets for a fully real-time experience.

---

## Architecture Overview

The backend follows a layered, DDD-inspired structure:

| Layer         | Package                    | Responsibility                                    |
| ------------- | -------------------------- | ------------------------------------------------- |
| **Domain**    | `internal/domain`          | Core models (Game, Board, Session), events, messages |
| **Repository**| `internal/repository`      | PostgreSQL + Redis data access implementations    |
| **Service**   | `internal/service`         | Business logic: matchmaking, bots, game sessions, auth |
| **Transport** | `internal/transport`       | HTTP REST endpoints + WebSocket connection handler |

---

## WebSocket Lifecycle

All real-time actions happen over a persistent WebSocket connection managed by `internal/transport/websocket/handler.go`.

### Connection Flow

```
Client connects via HTTP upgrade
  → Client sends {"type": "init", "jwt": "..."}
  → Server validates JWT against PostgreSQL/Redis
  → Server registers user in ConnectionManager (userId → *websocket.Conn)
  → If user has active game → server sends game_state for reconnection
  → Client is ready to matchmake, play, or spectate
```

### Decoupled Event Loop

To prevent race conditions during concurrent broadcasts, each `GameSession` has a dedicated event loop powered by Go channels:

1. A game event occurs (e.g., `MakeMove`)
2. The `GameSession` packages it as a `domain.GameEvent` and pushes it to `gs.Events` channel
3. The `ConsumeGameEvents` goroutine continuously reads from this channel
4. Events are serialized and sent to all participants via `ConnManager.SendMessage()`

This design ensures that a slow client or broadcast failure never blocks the game logic.

---

## Game Engine

A `GameSession` (defined in `internal/service/game/service.go`) represents a live match between two players (or player vs bot).

### Turn Processing (`HandleMove`)

When a `make_move` message arrives:

1. **Lock** — Acquire `sync.Mutex` on the game session for thread safety
2. **Validate** — Verify the request came from the player whose turn it is
3. **Execute** — `domain.Game.MakeMove()` finds the lowest empty row in the column (7×6 gravity)
4. **Win Check** — Scan horizontal, vertical, and both diagonal directions for four consecutive discs
5. **Broadcast** — Emit `move_made` or `game_over` event to all participants (including spectators)

### Disconnection & Reconnection

When a WebSocket drops mid-game:

- `HandleDisconnect()` starts a **60-second grace timer** (`DisconnectTimer`)
- If the timer expires → automatic forfeit, opponent wins by abandonment
- If the player reconnects (new WebSocket + valid JWT):
  - The `init` handler detects an ongoing session
  - Cancels the `DisconnectTimer`
  - Sends the latest `game_state` to the reconnecting client

---

## Bot Engine

Three difficulty levels, all non-blocking. When `HandleMove` detects a bot turn, it spawns a goroutine with artificial delay before calling `HandleBotMove()`.

| Difficulty | Algorithm                                                      |
| ---------- | -------------------------------------------------------------- |
| **Easy**   | Random valid columns with basic immediate win/block detection  |
| **Medium** | Positional weight grid + shallow threat evaluation (2/3-in-a-row scoring) |
| **Hard**   | Minimax with alpha-beta pruning, depth 7, positional weight matrix |

The hard bot evaluates thousands of future board states recursively, discarding sub-optimal branches via alpha-beta pruning to keep response times under ~200ms.

---

## Matchmaking

The matchmaking system runs independently in `internal/service/matchmaking/queue.go`:

1. Player enters a synchronized queue via `find_match`
2. A background ticker (`Run()`) continuously scans for two available players
3. When a pair is found:
   - Both are removed from the queue
   - A new `GameSession` is created
   - `game_start` events are sent down both WebSockets
4. If no opponent is found within **10 seconds**, the player is automatically matched against a bot

---

## Data Persistence

When games conclude (naturally or by abandonment), `saveGameAsync` runs in a background goroutine. It persists:

- Match duration and total moves
- Final board state (JSONB)
- Winner, reason, and updated Elo ratings

This powers the Game History page and Leaderboard rankings on the frontend.
