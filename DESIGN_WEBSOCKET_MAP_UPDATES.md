# Design: WebSocket Integration for Real-Time Map Updates

## Overview

This design adds WebSocket support to the gastown dashboard for real-time map updates, replacing the current 30-second HTMX polling mechanism with push-based updates.

## Current State

### Existing Architecture
- **Web Dashboard**: HTTP handler with HTMX polling every 30 seconds (`internal/web/handler.go`)
- **Event System**: File-based event logs (`.events.jsonl`) with multiple sources
- **Data Fetching**: On-demand convoy/polecat status via `LiveConvoyFetcher`
- **TUI**: Uses Go channels with `MultiSource` fan-in pattern for real-time updates

### No WebSocket Dependencies
- Currently uses standard `net/http`
- No `gorilla/websocket` or similar libraries

## Proposed Solution

### 1. Architecture

```
┌─────────────────┐
│  Event Sources  │ (Existing)
│  - BdActivity   │
│  - GtEvents     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  MultiSource    │ (Reuse existing)
│  (Fan-in)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  WebSocket Hub  │ (New)
│  - Broadcast    │
│  - Client Mgmt  │
└────────┬────────┘
         │
         ▼
    ┌────┴────┐
    │         │
┌───▼──┐  ┌──▼───┐
│Client│  │Client│  (WebSocket Connections)
└──────┘  └──────┘
```

### 2. New Package: `internal/websocket`

#### **hub.go** - Central Message Broker
```go
type Hub struct {
    sources    []EventSource     // Reuse existing event infrastructure
    clients    map[*Client]bool  // Active connections
    broadcast  chan Event        // From MultiSource
    register   chan *Client
    unregister chan *Client
    done       chan struct{}
}

func NewHub(sources []EventSource) *Hub
func (h *Hub) Run()
func (h *Hub) Broadcast(event Event)
```

**Key Responsibilities**:
- Aggregate events from existing `EventSource` implementations
- Maintain registry of connected WebSocket clients
- Broadcast events to all connected clients
- Handle client registration/deregistration

#### **client.go** - Individual Connection Handler
```go
type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan Event
}

func (c *Client) ReadPump()  // Handle incoming messages (heartbeat)
func (c *Client) WritePump() // Push events to client
```

**Key Responsibilities**:
- Manage individual WebSocket connection lifecycle
- Handle write buffering to prevent slow-client blocking
- Implement ping/pong for connection health
- Clean shutdown on disconnect

#### **messages.go** - Message Types
```go
type MessageType string

const (
    TypeMapUpdate      MessageType = "map_update"
    TypeAgentStatus    MessageType = "agent_status"
    TypeConvoyProgress MessageType = "convoy_progress"
    TypeError          MessageType = "error"
)

type Message struct {
    Type      MessageType     `json:"type"`
    Timestamp string          `json:"timestamp"`
    Data      interface{}     `json:"data"`
}

type MapUpdateData struct {
    Rigs    []RigData    `json:"rigs"`
    Convoys []ConvoyData `json:"convoys"`
}
```

### 3. Integration Points

#### **Modify: `internal/web/handler.go`**
```go
type ConvoyHandler struct {
    fetcher  ConvoyFetcher
    template *template.Template
    wsHub    *websocket.Hub  // New field
}

// New WebSocket endpoint
func (h *ConvoyHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Upgrade(w, r)
    if err != nil {
        return
    }
    client := websocket.NewClient(h.wsHub, conn)
    h.wsHub.Register(client)

    go client.WritePump()
    go client.ReadPump()
}
```

#### **Modify: `cmd/dashboard.go`**
```go
// Start WebSocket hub
sources := []EventSource{
    NewBdActivitySource(),
    NewGtEventsSource(),
}
hub := websocket.NewHub(sources)
go hub.Run()

// Mount WebSocket endpoint
http.HandleFunc("/ws", handler.HandleWebSocket)
```

### 4. Message Flow

#### Startup Sequence
1. Dashboard starts HTTP server
2. WebSocket hub initializes with event sources (BdActivity, GtEvents)
3. Hub spawns goroutine to listen to `MultiSource` events
4. Client connects → registers with hub → starts read/write pumps

#### Runtime Event Flow
1. Event occurs (e.g., convoy status change)
2. `BdActivitySource` detects change via `bd activity --follow`
3. Event flows through `MultiSource` → Hub broadcast channel
4. Hub iterates over registered clients and sends to each `client.send` channel
5. Each client's `WritePump` writes to WebSocket connection

#### Client Disconnect
1. `ReadPump` detects connection close
2. Client sends to hub's `unregister` channel
3. Hub removes client from registry and closes `client.send` channel
4. `WritePump` exits on channel close

## Edge Cases & Failure Modes

### 1. Slow Client Protection
**Problem**: Slow client blocks broadcast to all clients
**Solution**: Buffered `client.send` channel (size: 256). If full, drop message or disconnect client
```go
select {
case client.send <- message:
default:
    // Client too slow, disconnect
    close(client.send)
    delete(h.clients, client)
}
```

### 2. Connection Health
**Problem**: Detect dead connections
**Solution**: Ping/pong with timeout
```go
const (
    pongWait   = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
)

conn.SetReadDeadline(time.Now().Add(pongWait))
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(pongWait))
    return nil
})
```

### 3. Event Source Failures
**Problem**: One event source fails (e.g., `bd activity` crashes)
**Solution**:
- Each source runs in separate goroutine with recovery
- Hub continues operating with remaining sources
- Log errors for debugging
```go
func (m *MultiSource) forwardEvents(src EventSource) {
    defer m.wg.Done()
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Event source panicked: %v", r)
        }
    }()
    // ... forward events
}
```

### 4. Thundering Herd
**Problem**: 100 clients all connect simultaneously
**Solution**:
- Channel-based registration prevents race conditions
- Hub goroutine serializes client management
- Go's scheduler handles goroutine scaling (200 client goroutines = ~400KB memory)

### 5. Memory Leaks
**Problem**: Clients disconnect but goroutines/channels persist
**Solution**:
- Always close `client.send` channel on unregister
- `WritePump` exits on channel close
- `ReadPump` exits on read error
- Deferred cleanup in both pumps:
```go
defer func() {
    h.hub.unregister <- c
    c.conn.Close()
}()
```

### 6. Race Conditions
**Problem**: Concurrent map access in hub
**Solution**: Hub runs in single goroutine, all map operations via channels
```go
for {
    select {
    case client := <-h.register:
        h.clients[client] = true
    case client := <-h.unregister:
        delete(h.clients, client)
    case event := <-h.broadcast:
        for client := range h.clients {
            // ... send event
        }
    }
}
```

## Simpler Alternatives Considered

### Alternative 1: Server-Sent Events (SSE)
**Pros**: Simpler protocol, HTTP-only, auto-reconnect
**Cons**: One-way only (server→client), less efficient for bidirectional future needs
**Decision**: WebSocket chosen for future bidirectional features (e.g., client filters, pause/resume)

### Alternative 2: Long Polling
**Pros**: Works everywhere, no new dependencies
**Cons**: Higher latency, more server load, connection churn
**Decision**: Too similar to current HTMX approach, doesn't solve latency issue

### Alternative 3: GraphQL Subscriptions
**Pros**: Strongly typed, introspectable
**Cons**: Massive complexity, requires GraphQL server, overkill for simple events
**Decision**: Over-engineered for current needs

## Dependencies

### New Go Module
```bash
go get github.com/gorilla/websocket
```

**Why gorilla/websocket?**
- Battle-tested (used by Docker, Kubernetes)
- Clean API for upgrading HTTP connections
- Built-in ping/pong helpers
- 15K+ stars, active maintenance

## Testing Strategy

### Unit Tests
1. **Hub**: Test client registration/unregistration, broadcast fanout
2. **Client**: Test message serialization, buffer overflow handling
3. **Message**: Test JSON serialization of all message types

### Integration Tests
1. Connect multiple clients, verify all receive same events
2. Disconnect client mid-stream, verify no impact on others
3. Send high-frequency events, verify buffering/backpressure

### Manual Testing
1. Open dashboard in multiple browser tabs
2. Trigger events via `bd` commands
3. Verify real-time updates in all tabs
4. Check browser DevTools → Network → WS for message flow

## Rollout Plan

### Phase 1: WebSocket Infrastructure (No Behavioral Change)
- Add `internal/websocket` package
- Add endpoint to dashboard handler
- Hub runs but dashboard still uses HTMX polling
- **Risk**: Low (new code path, not activated)

### Phase 2: Enable for Subset of Users
- Add feature flag or environment variable
- Dashboard checks flag to use WebSocket vs polling
- Monitor logs for connection errors
- **Risk**: Medium (new code path activated)

### Phase 3: Full Rollout
- Remove HTMX polling, WebSocket only
- Remove feature flag
- **Risk**: Low (validated in Phase 2)

## Success Metrics

1. **Latency**: Event-to-UI update < 1 second (vs. 0-30 seconds with polling)
2. **Server Load**: Reduced HTTP request rate (no 30s polling)
3. **Reliability**: Connection uptime > 99% for 10-minute sessions
4. **Resource Usage**: < 10KB memory per connected client

## Security Considerations

### 1. Origin Validation
```go
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "http://localhost:8080" || origin == ""
    },
}
```

### 2. Authentication
- WebSocket inherits HTTP session cookies
- No additional auth needed for local dashboard
- For future remote access: validate JWT in CheckOrigin

### 3. Rate Limiting
```go
type Client struct {
    lastMessageTime time.Time
    messageCount    int
}

// In ReadPump, reject clients sending > 10 msg/sec
```

### 4. Message Size Limits
```go
conn.SetReadLimit(512 * 1024) // 512KB max message
```

## Open Questions

1. **Should we support message filtering?** (e.g., client subscribes to specific convoy)
   - **Decision**: Not in v1. Broadcast all events, client-side filtering if needed

2. **Should we support client→server messages?** (e.g., pause updates)
   - **Decision**: Not in v1. Ping/pong only for connection health

3. **Should we persist event history for reconnecting clients?**
   - **Decision**: Not in v1. Client gets fresh state on connect, then live updates

## Implementation Checklist

- [ ] Add `gorilla/websocket` dependency
- [ ] Create `internal/websocket/hub.go`
- [ ] Create `internal/websocket/client.go`
- [ ] Create `internal/websocket/messages.go`
- [ ] Modify `internal/web/handler.go` to add WebSocket endpoint
- [ ] Modify `cmd/dashboard.go` to start hub
- [ ] Update dashboard HTML template with WebSocket client JS
- [ ] Add unit tests for hub, client, messages
- [ ] Add integration test for multi-client broadcast
- [ ] Manual testing with browser DevTools

## Timeline

**Not applicable** - Task breakdown into concrete implementation steps, no time estimates.

## References

- Existing event infrastructure: `internal/tui/feed/events.go`
- Existing multi-source pattern: `internal/tui/feed/multi_source.go`
- Existing data models: `internal/web/fetcher.go`
- gorilla/websocket examples: https://github.com/gorilla/websocket/tree/master/examples
