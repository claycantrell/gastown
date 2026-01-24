package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/steveyegge/gastown/internal/activity"
)

func TestHub_ServeWs(t *testing.T) {
	// Create mock fetcher
	mock := &MockConvoyFetcher{
		Convoys: []ConvoyRow{
			{
				ID:     "hq-cv-test1",
				Title:  "Test Convoy",
				Status: "open",
				Progress: "1/2",
				LastActivity: activity.Info{
					FormattedAge: "2m ago",
					ColorClass:   activity.ColorGreen,
				},
			},
		},
		MergeQueue: []MergeQueueRow{},
		Polecats:   []PolecatRow{},
	}

	// Create hub
	hub := NewHub(mock)
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	// Convert http://... to ws://...
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Set read deadline
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Read initial message
	var data ConvoyData
	err = ws.ReadJSON(&data)
	if err != nil {
		t.Fatalf("Failed to read JSON from WebSocket: %v", err)
	}

	// Verify data
	if len(data.Convoys) != 1 {
		t.Errorf("Expected 1 convoy, got %d", len(data.Convoys))
	}
	if len(data.Convoys) > 0 && data.Convoys[0].ID != "hq-cv-test1" {
		t.Errorf("Expected convoy ID 'hq-cv-test1', got '%s'", data.Convoys[0].ID)
	}
}

func TestHub_MultipleClients(t *testing.T) {
	// Create mock fetcher
	mock := &MockConvoyFetcher{
		Convoys: []ConvoyRow{
			{
				ID:     "hq-cv-multi",
				Title:  "Multi-client Test",
				Status: "open",
			},
		},
		MergeQueue: []MergeQueueRow{},
		Polecats:   []PolecatRow{},
	}

	// Create hub
	hub := NewHub(mock)
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect first client
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client 1: %v", err)
	}
	defer ws1.Close()

	// Connect second client
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client 2: %v", err)
	}
	defer ws2.Close()

	// Both clients should receive initial data
	ws1.SetReadDeadline(time.Now().Add(5 * time.Second))
	ws2.SetReadDeadline(time.Now().Add(5 * time.Second))

	var data1, data2 ConvoyData
	if err := ws1.ReadJSON(&data1); err != nil {
		t.Errorf("Client 1 failed to read: %v", err)
	}
	if err := ws2.ReadJSON(&data2); err != nil {
		t.Errorf("Client 2 failed to read: %v", err)
	}

	// Verify both received the same data
	if len(data1.Convoys) != len(data2.Convoys) {
		t.Error("Clients received different data")
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	mock := &MockConvoyFetcher{
		Convoys:    []ConvoyRow{},
		MergeQueue: []MergeQueueRow{},
		Polecats:   []PolecatRow{},
	}

	hub := NewHub(mock)
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect and immediately disconnect
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Close connection
	ws.Close()

	// Give hub time to process disconnect
	time.Sleep(100 * time.Millisecond)

	// Hub should handle disconnect gracefully (no panic)
	// If we get here without panic, test passes
}
