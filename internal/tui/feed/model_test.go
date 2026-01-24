package feed

import (
	"testing"
	"time"
)

func TestMaxTreeCursor_EmptyModel(t *testing.T) {
	m := NewModel()
	max := m.maxTreeCursor()
	if max != 0 {
		t.Errorf("Expected max cursor 0 for empty model, got %d", max)
	}
}

func TestMaxTreeCursor_WithAgents(t *testing.T) {
	m := NewModel()

	// Add a rig with 2 agents
	m.rigs["testrig"] = &Rig{
		Name:     "testrig",
		Expanded: true,
		Agents: map[string]*Agent{
			"agent1": {
				ID:   "agent1",
				Name: "agent1",
				Role: "polecat",
			},
			"agent2": {
				ID:   "agent2",
				Name: "agent2",
				Role: "polecat",
			},
		},
	}

	max := m.maxTreeCursor()
	// Should be 2 agents - 1 = 1
	if max != 1 {
		t.Errorf("Expected max cursor 1 for 2 agents, got %d", max)
	}
}

func TestMaxTreeCursor_WithExpandedAgent(t *testing.T) {
	m := NewModel()

	// Add a rig with 2 agents, one expanded
	m.rigs["testrig"] = &Rig{
		Name:     "testrig",
		Expanded: true,
		Agents: map[string]*Agent{
			"agent1": {
				ID:   "agent1",
				Name: "agent1",
				Role: "polecat",
			},
			"agent2": {
				ID:   "agent2",
				Name: "agent2",
				Role: "polecat",
			},
		},
	}

	// Expand agent1
	m.expandedAgents["agent1"] = true

	max := m.maxTreeCursor()
	// Should be 2 agents + 1 expanded details row - 1 = 2
	if max != 2 {
		t.Errorf("Expected max cursor 2 for 2 agents (1 expanded), got %d", max)
	}
}

func TestMaxConvoyCursor_EmptyState(t *testing.T) {
	m := NewModel()
	max := m.maxConvoyCursor()
	if max != 0 {
		t.Errorf("Expected max cursor 0 for nil convoy state, got %d", max)
	}
}

func TestMaxConvoyCursor_WithConvoys(t *testing.T) {
	m := NewModel()
	m.convoyState = &ConvoyState{
		InProgress: []Convoy{
			{ID: "hq-1", Title: "Conv 1"},
			{ID: "hq-2", Title: "Conv 2"},
		},
		Landed: []Convoy{
			{ID: "hq-3", Title: "Conv 3"},
		},
	}

	max := m.maxConvoyCursor()
	// Should be 3 convoys - 1 = 2
	if max != 2 {
		t.Errorf("Expected max cursor 2 for 3 convoys, got %d", max)
	}
}

func TestGetAgentAtCursor(t *testing.T) {
	m := NewModel()

	// Add a rig with agents
	m.rigs["testrig"] = &Rig{
		Name:     "testrig",
		Expanded: true,
		Agents: map[string]*Agent{
			"polecat1": {
				ID:   "polecat1",
				Name: "polecat1",
				Role: "polecat",
			},
			"polecat2": {
				ID:   "polecat2",
				Name: "polecat2",
				Role: "polecat",
			},
		},
	}

	// Cursor at position 0 should return first agent
	m.treeCursor = 0
	agent := m.getAgentAtCursor()
	if agent == nil {
		t.Fatal("Expected agent at cursor 0, got nil")
	}

	// Cursor at position 1 should return second agent
	m.treeCursor = 1
	agent = m.getAgentAtCursor()
	if agent == nil {
		t.Fatal("Expected agent at cursor 1, got nil")
	}

	// Cursor beyond range should return nil
	m.treeCursor = 999
	agent = m.getAgentAtCursor()
	if agent != nil {
		t.Errorf("Expected nil agent at cursor 999, got %v", agent)
	}
}

func TestGetConvoyAtCursor(t *testing.T) {
	m := NewModel()
	m.convoyState = &ConvoyState{
		InProgress: []Convoy{
			{ID: "hq-1", Title: "Conv 1"},
			{ID: "hq-2", Title: "Conv 2"},
		},
		Landed: []Convoy{
			{ID: "hq-3", Title: "Conv 3"},
		},
	}

	// Cursor at position 0 should return first in-progress convoy
	m.convoyCursor = 0
	convoy := m.getConvoyAtCursor()
	if convoy == nil || convoy.ID != "hq-1" {
		t.Errorf("Expected convoy hq-1 at cursor 0, got %v", convoy)
	}

	// Cursor at position 1 should return second in-progress convoy
	m.convoyCursor = 1
	convoy = m.getConvoyAtCursor()
	if convoy == nil || convoy.ID != "hq-2" {
		t.Errorf("Expected convoy hq-2 at cursor 1, got %v", convoy)
	}

	// Cursor at position 2 should return first landed convoy
	m.convoyCursor = 2
	convoy = m.getConvoyAtCursor()
	if convoy == nil || convoy.ID != "hq-3" {
		t.Errorf("Expected convoy hq-3 at cursor 2, got %v", convoy)
	}

	// Cursor beyond range should return nil
	m.convoyCursor = 999
	convoy = m.getConvoyAtCursor()
	if convoy != nil {
		t.Errorf("Expected nil convoy at cursor 999, got %v", convoy)
	}
}

func TestToggleTreeExpansion(t *testing.T) {
	m := NewModel()

	// Add a rig with an agent
	m.rigs["testrig"] = &Rig{
		Name:     "testrig",
		Expanded: true,
		Agents: map[string]*Agent{
			"agent1": {
				ID:         "agent1",
				Name:       "agent1",
				Role:       "polecat",
				Status:     "working",
				LastUpdate: time.Now(),
			},
		},
	}

	m.treeCursor = 0

	// Initially not expanded
	if m.expandedAgents["agent1"] {
		t.Error("Agent should not be expanded initially")
	}

	// Toggle expansion
	m.toggleTreeExpansion()

	// Should be expanded now
	if !m.expandedAgents["agent1"] {
		t.Error("Agent should be expanded after toggle")
	}

	// Should have cached details
	if m.agentDetailsCache["agent1"] == nil {
		t.Error("Agent details should be cached")
	}

	// Toggle again
	m.toggleTreeExpansion()

	// Should be collapsed now
	if m.expandedAgents["agent1"] {
		t.Error("Agent should be collapsed after second toggle")
	}
}

func TestToggleConvoyExpansion(t *testing.T) {
	m := NewModel()
	m.convoyState = &ConvoyState{
		InProgress: []Convoy{
			{
				ID:        "hq-1",
				Title:     "Conv 1",
				Status:    "open",
				Completed: 2,
				Total:     5,
				CreatedAt: time.Now(),
			},
		},
	}

	m.convoyCursor = 0

	// Initially not expanded
	if m.expandedConvoys["hq-1"] {
		t.Error("Convoy should not be expanded initially")
	}

	// Toggle expansion
	m.toggleConvoyExpansion()

	// Should be expanded now
	if !m.expandedConvoys["hq-1"] {
		t.Error("Convoy should be expanded after toggle")
	}

	// Should have cached details
	if m.convoyDetailsCache["hq-1"] == nil {
		t.Error("Convoy details should be cached")
	}

	// Toggle again
	m.toggleConvoyExpansion()

	// Should be collapsed now
	if m.expandedConvoys["hq-1"] {
		t.Error("Convoy should be collapsed after second toggle")
	}
}

func TestFetchAgentDetails(t *testing.T) {
	m := NewModel()

	agent := &Agent{
		ID:         "test-agent",
		Name:       "test-agent",
		Role:       "polecat",
		Status:     "working",
		LastUpdate: time.Now(),
	}

	details := m.fetchAgentDetails(agent)

	if details == nil {
		t.Fatal("Expected agent details, got nil")
	}

	if details.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", details.Name)
	}

	if details.Role != "polecat" {
		t.Errorf("Expected role 'polecat', got '%s'", details.Role)
	}

	if details.State != "working" {
		t.Errorf("Expected state 'working', got '%s'", details.State)
	}

	if !details.Running {
		t.Error("Expected Running to be true for 'working' status")
	}
}

func TestFetchConvoyDetails(t *testing.T) {
	m := NewModel()

	now := time.Now()
	convoy := &Convoy{
		ID:        "hq-test",
		Title:     "Test Convoy",
		Status:    "open",
		Completed: 3,
		Total:     10,
		CreatedAt: now,
	}

	details := m.fetchConvoyDetails(convoy)

	if details == nil {
		t.Fatal("Expected convoy details, got nil")
	}

	if details.ID != "hq-test" {
		t.Errorf("Expected ID 'hq-test', got '%s'", details.ID)
	}

	if details.Title != "Test Convoy" {
		t.Errorf("Expected title 'Test Convoy', got '%s'", details.Title)
	}

	if details.Completed != 3 {
		t.Errorf("Expected completed 3, got %d", details.Completed)
	}

	if details.Total != 10 {
		t.Errorf("Expected total 10, got %d", details.Total)
	}
}
