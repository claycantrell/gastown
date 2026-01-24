package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/steveyegge/gastown/internal/tui/feed"
)

// Hub manages WebSocket clients and broadcasts events
type Hub struct {
	// Event source
	eventSource feed.EventSource

	// Client management
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client

	// Message broadcasting
	broadcast chan *Message

	// Lifecycle
	done      chan struct{}
	closeOnce sync.Once
}

// NewHub creates a new WebSocket hub
func NewHub(eventSource feed.EventSource) *Hub {
	return &Hub{
		eventSource: eventSource,
		clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *Message, 256),
		done:        make(chan struct{}),
	}
}

// Run starts the hub's main event loop
// This should be called in a goroutine
func (h *Hub) Run() {
	defer h.Close()

	// Start event listener
	go h.listenToEvents()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected (total: %d)", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected (total: %d)", len(h.clients))
			}

		case message := <-h.broadcast:
			// Broadcast to all clients
			for client := range h.clients {
				select {
				case client.send <- message:
					// Message sent successfully
				default:
					// Client's send buffer is full, disconnect it
					log.Printf("Client send buffer full, disconnecting")
					close(client.send)
					delete(h.clients, client)
				}
			}

		case <-h.done:
			return
		}
	}
}

// listenToEvents listens to the event source and broadcasts events
func (h *Hub) listenToEvents() {
	events := h.eventSource.Events()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				log.Printf("Event source closed")
				return
			}

			// Convert feed.Event to WebSocket message
			message := h.convertEventToMessage(event)
			if message != nil {
				select {
				case h.broadcast <- message:
				default:
					// Broadcast buffer full, drop message
					log.Printf("Broadcast buffer full, dropping message")
				}
			}

		case <-h.done:
			return
		}
	}
}

// convertEventToMessage converts a feed.Event to a WebSocket Message
func (h *Hub) convertEventToMessage(event feed.Event) *Message {
	// Determine message type based on event type
	var msgType MessageType
	switch event.Type {
	case "create", "update", "complete", "fail", "delete":
		msgType = TypeMapUpdate
	default:
		msgType = TypeMapUpdate
	}

	// Create event data
	eventData := &EventData{
		Type:      event.Type,
		Message:   event.Message,
		Timestamp: event.Time.Format(time.RFC3339),
	}

	// Create agent data
	agentData := &AgentData{
		ID:         event.Actor,
		Name:       event.Actor,
		Role:       event.Role,
		Status:     getStatusFromEventType(event.Type),
		LastEvent:  eventData,
		LastUpdate: event.Time.Format(time.RFC3339),
	}

	// Create rig data (simplified for now)
	rigData := &RigData{
		Name:   event.Rig,
		Agents: []AgentData{*agentData},
	}

	// Create map update
	mapUpdate := &MapUpdateData{
		Rigs:    []RigData{*rigData},
		Convoys: []ConvoyData{}, // TODO: Add convoy data from fetcher
	}

	return NewMessage(msgType, mapUpdate)
}

// getStatusFromEventType maps event types to agent status
func getStatusFromEventType(eventType string) string {
	switch eventType {
	case "create":
		return "working"
	case "update":
		return "working"
	case "complete":
		return "idle"
	case "fail":
		return "error"
	case "delete":
		return "dead"
	default:
		return "idle"
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message *Message) {
	select {
	case h.broadcast <- message:
	default:
		log.Printf("Broadcast buffer full, dropping message")
	}
}

// Close gracefully shuts down the hub
func (h *Hub) Close() {
	h.closeOnce.Do(func() {
		close(h.done)

		// Close all client connections
		for client := range h.clients {
			close(client.send)
		}

		// Close event source
		if h.eventSource != nil {
			if err := h.eventSource.Close(); err != nil {
				log.Printf("Error closing event source: %v", err)
			}
		}
	})
}
