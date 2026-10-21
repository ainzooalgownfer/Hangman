# Hangman

Multiplayer Hangman server in Go. TCP, one JSON object per line, one goroutine per client.

## Run

```bash
go run ./cmd/server        # :8080
nc localhost 8080
```

```json
{"action":"NEW_GAME","player":"Alice"}
{"action":"LETTER","session_id":"<id>","letter":"e"}
{"action":"RESUME","session_id":"<id>"}
{"action":"QUIT","session_id":"<id>"}
```

With Docker (image built by CI on every push to `main`):

```bash
docker run --rm -p 8080:8080 ghcr.io/ainzooalgownfer/hangman:latest
```

Update: `docker pull ghcr.io/ainzooalgownfer/hangman:latest`

Build locally instead: `docker compose up --build`

## Protocol

| Action | Fields | Response |
|---|---|---|
| `NEW_GAME` | `player` | `STATE` |
| `LETTER` | `session_id`, `letter` | `STATE` (`IN_PROGRESS`, `WON`, `LOST`) |
| `RESUME` | `session_id` | `STATE` |
| `QUIT` | `session_id` | `BYE` (`ABANDONED`, `secret_word`), then the connection closes |

Errors: `{"type":"ERROR","code":"...","message":"..."}`

| Code | When |
|---|---|
| `INVALID_FORMAT` | invalid JSON or missing `player` |
| `UNKNOWN_ACTION` | unknown action |
| `SESSION_NOT_FOUND` | missing or unknown `session_id` |
| `INVALID_LETTER` | not a single letter |
| `ALREADY_GUESSED` | letter already guessed |
| `GAME_ALREADY_OVER` | game already won or lost |

## Structure

```
cmd/server         entry point
internal/network   TCP server, connections, logging
internal/command   dispatcher
  handlers         one handler per action, SessionStore interface
internal/game      GameSession, states, SessionManager, words
internal/protocol  request/response types
scripts            load simulation
```

## Class diagrams

### Server and commands

```mermaid
classDiagram
    direction TB
    class TCPServer {
        -address string
        -listener net.Listener
        +Start() error
        +Stop() error
        -handleConnection(conn)
    }
    class ClientConnection {
        -conn net.Conn
        -scanner bufio.Scanner
        +ReadCommand() CommandRequest
        +SendResponse(res) error
        +Close() error
    }
    class CommandDispatcher {
        -handlers map
        +Register(action, handler)
        +Dispatch(req) CommandResponse
    }
    class CommandHandler {
        <<interface>>
        +Handle(req) CommandResponse
    }
    class NewGameHandler
    class GuessLetterHandler
    class ResumeHandler
    class QuitHandler

    TCPServer *-- CommandDispatcher
    TCPServer ..> ClientConnection
    CommandDispatcher o-- CommandHandler
    CommandHandler <|.. NewGameHandler
    CommandHandler <|.. GuessLetterHandler
    CommandHandler <|.. ResumeHandler
    CommandHandler <|.. QuitHandler

    style TCPServer fill:#fdf0fd,stroke:#d36bd8,color:#1a1a1a
    style ClientConnection fill:#eefcf8,stroke:#2bbfa0,color:#1a1a1a
    style CommandDispatcher fill:#fff4e8,stroke:#f08a3c,color:#1a1a1a
    style CommandHandler fill:#e9fbfd,stroke:#22b8d4,color:#1a1a1a
    style NewGameHandler fill:#effcf2,stroke:#4cc970,color:#1a1a1a
    style GuessLetterHandler fill:#f4f1fe,stroke:#9b80f0,color:#1a1a1a
    style ResumeHandler fill:#fef1f1,stroke:#ef6b6b,color:#1a1a1a
    style QuitHandler fill:#fffbe8,stroke:#f2c230,color:#1a1a1a
```

### Game session and State pattern

```mermaid
classDiagram
    direction TB
    class GameSession {
        -id string
        -player string
        -secretWord string
        -maskedWord string
        -attemptsLeft int
        -triedLetters []string
        -currentState GameState
        +Guess(letter) CommandResponse
        +Snapshot() CommandResponse
        +SetState(state)
    }
    class GameState {
        <<interface>>
        +Guess(session, letter) CommandResponse
        +GetStatus() string
    }
    class InProgressState
    class WonState
    class LostState

    GameSession --> GameState : currentState
    GameState <|.. InProgressState
    GameState <|.. WonState
    GameState <|.. LostState

    style GameSession fill:#fdf0fd,stroke:#d36bd8,color:#1a1a1a
    style GameState fill:#e9fbfd,stroke:#22b8d4,color:#1a1a1a
    style InProgressState fill:#fff4e8,stroke:#f08a3c,color:#1a1a1a
    style WonState fill:#e9fbfd,stroke:#22b8d4,color:#1a1a1a
    style LostState fill:#effcf2,stroke:#4cc970,color:#1a1a1a
```

### Session store

```mermaid
classDiagram
    direction TB
    class Handlers {
        NewGame, GuessLetter
        Resume, Quit
    }
    class SessionStore {
        <<interface>>
        +Add(session)
        +Get(id) GameSession
        +Remove(id)
    }
    class SessionManager {
        -mu sync.Mutex
        -sessions map
        +Add(session)
        +Get(id) GameSession
        +Remove(id)
    }
    class GameSession

    Handlers --> SessionStore : uses
    SessionStore <|.. SessionManager
    SessionManager --> GameSession : stores

    style Handlers fill:#f4f1fe,stroke:#9b80f0,color:#1a1a1a
    style SessionStore fill:#fdf0fd,stroke:#d36bd8,color:#1a1a1a
    style SessionManager fill:#eefcf8,stroke:#2bbfa0,color:#1a1a1a
    style GameSession fill:#fff4e8,stroke:#f08a3c,color:#1a1a1a
```

### Patterns

| Pattern | Classes |
|---|---|
| Command / Handler registry | `CommandDispatcher`, `CommandHandler`, handlers |
| State | `GameSession`, `GameState`, `InProgressState`, `WonState`, `LostState` |
| Repository | `SessionStore`, `SessionManager` |

## Tests

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...    # needs gcc
```

Load test against a running server:

```bash
scripts/simulate.sh 200
```

## Status

Done:

- NEW_GAME, LETTER, RESUME, QUIT
- State pattern for game rules
- Thread-safe in-memory session store
- Resume after disconnect
- Input limits: 4 KB per line, 5 min idle timeout
- Structured logs (`log/slog`) with session IDs
- Unit, concurrency and load tests

TODO:

- SQLite persistence
- Check that a session belongs to the connection using it
- Expire abandoned sessions
- Graceful shutdown
- Config for port and attempts
- ResponseFactory
- Leaderboard
