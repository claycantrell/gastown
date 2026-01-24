package websocket

import (
	"testing"
	"time"

	"github.com/steveyegge/gastown/internal/tui/feed"
)

// MockEventSource is a mock implementation of feed.EventSource for testing
type MockEventSource struct {
	events chan feed.Event
	closed bool
}

func NewMockEventSource() *MockEventSource {
	return &MockEventSource{
		events: make(chan feed.Event, 10),
	}
}

func (m *MockEventSource) Events() <-chan feed.Event {
	return m.events
}

func (m *MockEventSource) Close() error {
	if !m.closed {
		m.closed = true
		close(m.events)
	}
	return nil
}

func (m *MockEventSource) SendEvent(event feed.Event) {
	if !m.closed {
		m.events <- event
	}
}

func TestHub_CreateAndClose(t *testing.T) {
	mockSource := NewMockEventSource()
	hub := NewHub(mockSource)

	if hub == nil {
		t.Fatal("NewHub returned nil")
	}

	// Close immediately
	hub.Close()

	if !mockSource.closed {
		t.Error("Hub.Close() should close the event source")
	}
}

func TestHub_Run(t *testing.T) {
	mockSource := NewMockEventSource()
	hub := NewHub(mockSource)

	// Start hub in goroutine
	go hub.Run()

	// Give it time to start
	time.Sleep(10 * time.Millisecond)

	// Close hub
	hub.Close()

	// Give it time to shut down
	time.Sleep(10 * time.Millisecond)
}

func TestHub_Broadcast(t *testing.T) {
	mockSource := NewMockEventSource()
	hub := NewHub(mockSource)

	go hub.Run()
	defer hub.Close()

	// Create a test message
	msg := NewMessage(TypeMapUpdate, map[string]string{"test": "data"})

	// Broadcast message
	hub.Broadcast(msg)

	// Give it time to process
	time.Sleep(10 * time.Millisecond)
}

func TestHub_EventProcessing(t *testing.T) {
	mockSource := NewMockEventSource()
	hub := NewHub(mockSource)

	go hub.Run()
	defer hub.Close()

	// Send a test event
	testEvent := feed.Event{
		Time:    time.Now(),
		Type:    "create",
		Actor:   "test-actor",
		Target:  "test-target",
		Message: "Test message",
		Rig:     "test-rig",
		Role:    "test-role",
	}

	mockSource.SendEvent(testEvent)

	// Give it time to process
	time.Sleep(50 * time.Millisecond)
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage(TypeMapUpdate, map[string]string{"key": "value"})

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}

	if msg.Type != TypeMapUpdate {
		t.Errorf("Message type = %v, want %v", msg.Type, TypeMapUpdate)
	}

	if msg.Timestamp == "" {
		t.Error("Message timestamp should not be empty")
	}

	if msg.Data == nil {
		t.Error("Message data should not be nil")
	}
}

func TestNewErrorMessage(t *testing.T) {
	errMsg := NewErrorMessage("test error")

	if errMsg == nil {
		t.Fatal("NewErrorMessage returned nil")
	}

	if errMsg.Type != TypeError {
		t.Errorf("Error message type = %v, want %v", errMsg.Type, TypeError)
	}

	data, ok := errMsg.Data.(map[string]string)
	if !ok {
		t.Fatal("Error message data is not map[string]string")
	}

	if data["error"] != "test error" {
		t.Errorf("Error message = %q, want %q", data["error"], "test error")
	}
}

func TestGetStatusFromEventType(t *testing.T) {
	tests := []struct {
		eventType string
		want      string
	}{
		{"create", "working"},
		{"update", "working"},
		{"complete", "idle"},
		{"fail", "error"},
		{"delete", "dead"},
		{"unknown", "idle"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			got := getStatusFromEventType(tt.eventType)
			if got != tt.want {
				t.Errorf("getStatusFromEventType(%q) = %q, want %q", tt.eventType, got, tt.want)
			}
		})
	}
}
