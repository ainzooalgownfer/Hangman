# Architecture du serveur

Diagrammes générés à partir du code (commit `5ce8d34`). GitHub affiche les blocs Mermaid directement.


## 2. Classes principales

```mermaid
%%{init: {"class": {"curve": "stepAfter"}}}%%
classDiagram
    direction LR

    class TCPServer {
        -listener net.Listener
        -dispatcher *CommandDispatcher
        -nextConnID atomic.Int64
        +Start() error
        +Server() error
        +Stop() error
        -handleConnection(conn)
    }
    class ClientConnection {
        -conn net.Conn
        -scanner *bufio.Scanner
        -idleTimeout
        +ReadCommand() CommandRequest
        +SendResponse(resp) error
    }
    class CommandDispatcher {
        -handlers map[string]CommandHandler
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

    class SessionStore {
        <<interface>>
        +Add(s)
        +Get(id) GameSession, bool
        +Remove(id)
    }
    class SessionManager {
        -mu sync.Mutex
        -sessions map[string]*GameSession
        +RemoveExpired(now, ttl) int
        +StartJanitor(ttl, interval) stop
    }
    class GameSession {
        -mu sync.Mutex
        -currentState GameState
        -resumeToken
        -owner int64
        -lastRequestID / lastResponse
        +GuessOnce(requestID, connID, letter)
        +QuitOnce(requestID, connID)
        +Claim(connID)
        +Snapshot()
    }
    class StateSession {
        <<interface>>
        +SetState(GameState)
        +AddTriedLetter(letter)
        +DecrementAttempts()
        ...
    }
    class GameState {
        <<interface>>
        +Guess(session, letter)
        +Quit(session)
        +GetStatus() string
    }
    class InProgressState
    class WonState
    class LostState
    class AbandonedState

    TCPServer --> ClientConnection : une par client
    TCPServer --> CommandDispatcher
    CommandDispatcher o-- CommandHandler
    CommandHandler <|.. NewGameHandler
    CommandHandler <|.. GuessLetterHandler
    CommandHandler <|.. ResumeHandler
    CommandHandler <|.. QuitHandler
    NewGameHandler --> SessionStore
    GuessLetterHandler --> SessionStore
    ResumeHandler --> SessionStore
    QuitHandler --> SessionStore
    SessionStore <|.. SessionManager
    SessionManager o-- GameSession
    StateSession <|.. GameSession
    GameSession --> GameState : état courant
    GameState <|.. InProgressState
    GameState <|.. WonState
    GameState <|.. LostState
    GameState <|.. AbandonedState
```

Patrons utilisés : **Command** (dispatcher + un handler par action), **State** (un objet par état de partie, pas de `if partieFinie`), **Repository** (`SessionStore` / `SessionManager`).

## 3. Machine à états d'une partie

```mermaid
stateDiagram-v2
    [*] --> EN_COURS : NEW_GAME

    EN_COURS --> EN_COURS : LETTER (lettre trouvée ou ratée, essais > 0)
    EN_COURS --> GAGNÉ : LETTER (mot complet)
    EN_COURS --> PERDU : LETTER (essais = 0)
    EN_COURS --> ABANDONNÉ : QUIT

    GAGNÉ --> [*] : expiration (TTL)
    PERDU --> [*] : expiration (TTL)
    ABANDONNÉ --> [*] : expiration (TTL)
    EN_COURS --> [*] : expiration (TTL)
```

Messages autorisés dans chaque état :

| État | `LETTER` | `QUIT` | `RESUME` |
|---|---|---|---|
| `IN_PROGRESS` |  `STATE` (ou `INVALID_LETTER`, `ALREADY_GUESSED`) |  `BYE` |  `STATE` |
| `WON` |  `INVALID_STATE` |  `INVALID_STATE` |  `STATE` + `secret_word` |
| `LOST` |  `INVALID_STATE` |  `INVALID_STATE` |  `STATE` + `secret_word` |
| `ABANDONED` |  `INVALID_STATE` |  `INVALID_STATE` |  `STATE` + `secret_word` |
| expirée / inconnue | `UNKNOWN_SESSION` | `UNKNOWN_SESSION` | `UNKNOWN_SESSION` |

Dans tous les états, `LETTER` et `QUIT` envoyés depuis une autre connexion que le propriétaire donnent `PLAYER_NOT_ALLOWED`, et `RESUME` sans le bon `resume_token` aussi.

## 4. Parcours d'un `LETTER`

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant G as goroutine handleConnection<br/>(une par client)
    participant CC as ClientConnection
    participant D as CommandDispatcher
    participant H as GuessLetterHandler
    participant M as SessionManager
    participant S as GameSession
    participant St as InProgressState

    C->>CC: {"action":"LETTER",...}\n (octets TCP)
    G->>CC: ReadCommand()
    Note over CC: SetReadDeadline (idle timeout)<br/>Scan() jusqu'au \n<br/>json.Unmarshal
    CC-->>G: CommandRequest
    Note over G: req.ConnID = connID
    G->>D: Dispatch(req)
    D->>H: Handle(req)
    Note over H: session_id présent ?<br/>une seule lettre valide ?
    H->>M: Get(session_id)
    Note over M: verrou M.mu (lecture de la map)
    M-->>H: *GameSession
    H->>S: GuessOnce(request_id, connID, lettre)
    Note over S: verrou S.mu<br/>1. connID == owner ? sinon PLAYER_NOT_ALLOWED<br/>2. request_id == dernier ? → réponse rejouée (replayed)<br/>3. lastActive = now
    S->>St: Guess(session, lettre)
    Note over St: ALREADY_GUESSED ?<br/>ajoute la lettre, décrémente si ratée<br/>mot complet → WonState<br/>essais = 0 → LostState
    St-->>S: STATE
    Note over S: remember(request_id, réponse)<br/>libère S.mu
    S-->>H: CommandResponse
    H-->>D: CommandResponse
    D-->>G: CommandResponse
    Note over G: resp.RequestID = req.RequestID
    G->>CC: SendResponse(resp)
    Note over CC: SetWriteDeadline<br/>json.Marshal + \n
    CC->>C: {"type":"STATE",...}\n
```

Ordre des verrous : toujours `SessionManager.mu` puis `GameSession.mu`, jamais l'inverse. Ici `M.mu` est déjà relâché quand `S.mu` est pris.