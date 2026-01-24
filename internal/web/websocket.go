package web

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for dashboard
		return true
	},
}

// Hub maintains active WebSocket connections and broadcasts updates.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan ConvoyData
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	fetcher    ConvoyFetcher
}

// NewHub creates a new WebSocket hub.
func NewHub(fetcher ConvoyFetcher) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan ConvoyData, 10),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		fetcher:    fetcher,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case data := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					// Client send buffer full, skip this message
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// Fetch latest data and broadcast
			h.fetchAndBroadcast()
		}
	}
}

// fetchAndBroadcast fetches current convoy data and broadcasts to all clients.
func (h *Hub) fetchAndBroadcast() {
	convoys, err := h.fetcher.FetchConvoys()
	if err != nil {
		log.Printf("Error fetching convoys: %v", err)
		return
	}

	mergeQueue, _ := h.fetcher.FetchMergeQueue()
	polecats, _ := h.fetcher.FetchPolecats()

	data := ConvoyData{
		Convoys:    convoys,
		MergeQueue: mergeQueue,
		Polecats:   polecats,
	}

	select {
	case h.broadcast <- data:
	default:
		// Broadcast channel full, skip this update
	}
}

// Client represents a WebSocket client connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan ConvoyData
}

// ServeWs handles WebSocket requests from clients.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan ConvoyData, 10),
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	// Send initial data immediately
	h.fetchAndBroadcast()
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send JSON data
			if err := c.conn.WriteJSON(data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
