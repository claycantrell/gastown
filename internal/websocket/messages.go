package websocket

import (
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	TypeMapUpdate      MessageType = "map_update"
	TypeAgentStatus    MessageType = "agent_status"
	TypeConvoyProgress MessageType = "convoy_progress"
	TypeError          MessageType = "error"
)

// Message is the top-level structure sent over WebSocket
type Message struct {
	Type      MessageType `json:"type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// MapUpdateData contains the full state of the map
type MapUpdateData struct {
	Rigs    []RigData    `json:"rigs"`
	Convoys []ConvoyData `json:"convoys"`
}

// RigData represents a single rig with its agents
type RigData struct {
	Name   string      `json:"name"`
	Agents []AgentData `json:"agents"`
}

// AgentData represents a single agent
type AgentData struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	LastEvent  *EventData `json:"last_event,omitempty"`
	LastUpdate string     `json:"last_update"`
}

// EventData represents an event associated with an agent
type EventData struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// ConvoyData represents a convoy with its progress
type ConvoyData struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Progress     string `json:"progress"`
	LastActivity string `json:"last_activity"`
}

// NewMessage creates a new message with the current timestamp
func NewMessage(msgType MessageType, data interface{}) *Message {
	return &Message{
		Type:      msgType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}
}

// NewErrorMessage creates an error message
func NewErrorMessage(errMsg string) *Message {
	return NewMessage(TypeError, map[string]string{
		"error": errMsg,
	})
}
