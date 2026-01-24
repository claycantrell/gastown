package feed

import (
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/steveyegge/gastown/internal/beads"
)

// Panel represents which panel has focus
type Panel int

const (
	PanelTree Panel = iota
	PanelConvoy
	PanelFeed
)

// Event represents an activity event
type Event struct {
	Time     time.Time
	Type     string // create, update, complete, fail, delete
	Actor    string // who did it (e.g., "gastown/crew/joe")
	Target   string // what was affected (e.g., "gt-xyz")
	Message  string // human-readable description
	Rig      string // which rig
	Role     string // actor's role
	Raw      string // raw line for fallback display
}

// Agent represents an agent in the tree
type Agent struct {
	ID         string
	Name       string
	Role       string // mayor, witness, refinery, crew, polecat
	Rig        string
	Status     string // running, idle, working, dead
	LastEvent  *Event
	LastUpdate time.Time
	Expanded   bool
}

// Rig represents a rig with its agents
type Rig struct {
	Name     string
	Agents   map[string]*Agent // keyed by role/name
	Expanded bool
}

// AgentDetails contains detailed information about an agent
type AgentDetails struct {
	Name          string
	Role          string
	State         string    // working, done, stuck
	Running       bool      // Is tmux session alive?
	HookBead      string    // Pinned bead ID
	WorkTitle     string    // Title of pinned work
	Branch        string    // Current git branch
	CleanupStatus string    // Git cleanup status
	UnreadMail    int       // Number of unread messages
	FirstSubject  string    // Subject of first unread
	CreatedAt     time.Time // When agent was created
	UpdatedAt     time.Time // Last activity time
}

// ConvoyDetails contains detailed information about a convoy
type ConvoyDetails struct {
	ID            string
	Title         string
	Status        string
	Completed     int
	Total         int
	TrackedIssues []TrackedIssue
	CreatedAt     time.Time
	ClosedAt      time.Time
	LastActivity  time.Time
}

// TrackedIssue represents an issue tracked by a convoy
type TrackedIssue struct {
	ID        string
	Title     string
	Status    string
	Assignee  string
	Worker    string
	WorkerAge string
}

// Model is the main bubbletea model for the feed TUI
type Model struct {
	// Dimensions
	width  int
	height int

	// Panels
	focusedPanel   Panel
	treeViewport   viewport.Model
	convoyViewport viewport.Model
	feedViewport   viewport.Model

	// Data
	rigs        map[string]*Rig
	events      []Event
	convoyState *ConvoyState
	townRoot    string

	// UI state
	keys     KeyMap
	help     help.Model
	showHelp bool
	filter   string

	// Selection and expansion state
	treeCursor         int                // Cursor position in tree panel
	convoyCursor       int                // Cursor position in convoy panel
	expandedAgents     map[string]bool    // Tracks expanded agents by full ID
	expandedConvoys    map[string]bool    // Tracks expanded convoys by ID
	agentDetailsCache  map[string]*AgentDetails  // Cache of detailed agent info
	convoyDetailsCache map[string]*ConvoyDetails // Cache of detailed convoy info

	// Event source
	eventChan <-chan Event
	done      chan struct{}
	closeOnce sync.Once
}

// NewModel creates a new feed TUI model
func NewModel() *Model {
	h := help.New()
	h.ShowAll = false

	return &Model{
		focusedPanel:       PanelTree,
		treeViewport:       viewport.New(0, 0),
		convoyViewport:     viewport.New(0, 0),
		feedViewport:       viewport.New(0, 0),
		rigs:               make(map[string]*Rig),
		events:             make([]Event, 0, 1000),
		keys:               DefaultKeyMap(),
		help:               h,
		done:               make(chan struct{}),
		expandedAgents:     make(map[string]bool),
		expandedConvoys:    make(map[string]bool),
		agentDetailsCache:  make(map[string]*AgentDetails),
		convoyDetailsCache: make(map[string]*ConvoyDetails),
	}
}

// SetTownRoot sets the town root for convoy fetching
func (m *Model) SetTownRoot(townRoot string) {
	m.townRoot = townRoot
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForEvents(),
		m.fetchConvoys(),
		tea.SetWindowTitle("GT Feed"),
	)
}

// eventMsg is sent when a new event arrives
type eventMsg Event

// convoyUpdateMsg is sent when convoy data is refreshed
type convoyUpdateMsg struct {
	state *ConvoyState
}

// tickMsg is sent periodically to refresh the view
type tickMsg time.Time

// listenForEvents returns a command that listens for events
func (m *Model) listenForEvents() tea.Cmd {
	if m.eventChan == nil {
		return nil
	}
	// Capture channels to avoid race with Model mutations
	eventChan := m.eventChan
	done := m.done
	return func() tea.Msg {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return nil
			}
			return eventMsg(event)
		case <-done:
			return nil
		}
	}
}

// tick returns a command for periodic refresh
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchConvoys returns a command that fetches convoy data
func (m *Model) fetchConvoys() tea.Cmd {
	if m.townRoot == "" {
		return nil
	}
	townRoot := m.townRoot
	return func() tea.Msg {
		state, _ := FetchConvoys(townRoot)
		return convoyUpdateMsg{state: state}
	}
}

// convoyRefreshTick returns a command that schedules the next convoy refresh
func (m *Model) convoyRefreshTick() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return convoyUpdateMsg{} // Empty state triggers a refresh
	})
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewportSizes()

	case eventMsg:
		m.addEvent(Event(msg))
		cmds = append(cmds, m.listenForEvents())

	case convoyUpdateMsg:
		if msg.state != nil {
			// Fresh data arrived - update state and schedule next tick
			m.convoyState = msg.state
			m.updateViewContent()
			cmds = append(cmds, m.convoyRefreshTick())
		} else {
			// Tick fired - fetch new data
			cmds = append(cmds, m.fetchConvoys())
		}

	case tickMsg:
		cmds = append(cmds, tick())
	}

	// Update viewports
	var cmd tea.Cmd
	switch m.focusedPanel {
	case PanelTree:
		m.treeViewport, cmd = m.treeViewport.Update(msg)
	case PanelConvoy:
		m.convoyViewport, cmd = m.convoyViewport.Update(msg)
	case PanelFeed:
		m.feedViewport, cmd = m.feedViewport.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleKey processes key presses
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.closeOnce.Do(func() { close(m.done) })
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		m.help.ShowAll = m.showHelp
		return m, nil

	case key.Matches(msg, m.keys.Tab):
		// Cycle: Tree -> Convoy -> Feed -> Tree
		switch m.focusedPanel {
		case PanelTree:
			m.focusedPanel = PanelConvoy
		case PanelConvoy:
			m.focusedPanel = PanelFeed
		case PanelFeed:
			m.focusedPanel = PanelTree
		}
		return m, nil

	case key.Matches(msg, m.keys.FocusTree):
		m.focusedPanel = PanelTree
		return m, nil

	case key.Matches(msg, m.keys.FocusFeed):
		m.focusedPanel = PanelFeed
		return m, nil

	case key.Matches(msg, m.keys.FocusConvoy):
		m.focusedPanel = PanelConvoy
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		m.updateViewContent()
		return m, nil

	case key.Matches(msg, m.keys.Enter), key.Matches(msg, m.keys.Expand):
		// Toggle expansion on focused panel
		switch m.focusedPanel {
		case PanelTree:
			m.toggleTreeExpansion()
			m.updateViewContent()
		case PanelConvoy:
			m.toggleConvoyExpansion()
			m.updateViewContent()
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		// Move cursor up in focused panel
		switch m.focusedPanel {
		case PanelTree:
			if m.treeCursor > 0 {
				m.treeCursor--
				m.updateViewContent()
			}
		case PanelConvoy:
			if m.convoyCursor > 0 {
				m.convoyCursor--
				m.updateViewContent()
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		// Move cursor down in focused panel
		switch m.focusedPanel {
		case PanelTree:
			max := m.maxTreeCursor()
			if m.treeCursor < max {
				m.treeCursor++
				m.updateViewContent()
			}
		case PanelConvoy:
			max := m.maxConvoyCursor()
			if m.convoyCursor < max {
				m.convoyCursor++
				m.updateViewContent()
			}
		}
		return m, nil
	}

	// Pass to focused viewport for scrolling
	var cmd tea.Cmd
	switch m.focusedPanel {
	case PanelTree:
		m.treeViewport, cmd = m.treeViewport.Update(msg)
	case PanelConvoy:
		m.convoyViewport, cmd = m.convoyViewport.Update(msg)
	case PanelFeed:
		m.feedViewport, cmd = m.feedViewport.Update(msg)
	}
	return m, cmd
}

// updateViewportSizes recalculates viewport dimensions
func (m *Model) updateViewportSizes() {
	// Reserve space: header (1) + borders (6 for 3 panels) + status bar (1) + help (1-2)
	headerHeight := 1
	statusHeight := 1
	helpHeight := 1
	if m.showHelp {
		helpHeight = 3
	}
	borderHeight := 6 // top and bottom borders for 3 panels

	availableHeight := m.height - headerHeight - statusHeight - helpHeight - borderHeight
	if availableHeight < 6 {
		availableHeight = 6
	}

	// Split: 30% tree, 25% convoy, 45% feed
	treeHeight := availableHeight * 30 / 100
	convoyHeight := availableHeight * 25 / 100
	feedHeight := availableHeight - treeHeight - convoyHeight

	// Ensure minimum heights
	if treeHeight < 3 {
		treeHeight = 3
	}
	if convoyHeight < 3 {
		convoyHeight = 3
	}
	if feedHeight < 3 {
		feedHeight = 3
	}

	contentWidth := m.width - 4 // borders and padding
	if contentWidth < 20 {
		contentWidth = 20
	}

	m.treeViewport.Width = contentWidth
	m.treeViewport.Height = treeHeight
	m.convoyViewport.Width = contentWidth
	m.convoyViewport.Height = convoyHeight
	m.feedViewport.Width = contentWidth
	m.feedViewport.Height = feedHeight

	m.updateViewContent()
}

// updateViewContent refreshes the content of all viewports
func (m *Model) updateViewContent() {
	m.treeViewport.SetContent(m.renderTree())
	m.convoyViewport.SetContent(m.renderConvoys())
	m.feedViewport.SetContent(m.renderFeed())
}

// addEvent adds an event and updates the agent tree
func (m *Model) addEvent(e Event) {
	// Update agent tree first (always do this for status tracking)
	if e.Rig != "" {
		rig, ok := m.rigs[e.Rig]
		if !ok {
			rig = &Rig{
				Name:     e.Rig,
				Agents:   make(map[string]*Agent),
				Expanded: true,
			}
			m.rigs[e.Rig] = rig
		}

		if e.Actor != "" {
			agent, ok := rig.Agents[e.Actor]
			if !ok {
				agent = &Agent{
					ID:   e.Actor,
					Name: e.Actor,
					Role: e.Role,
					Rig:  e.Rig,
				}
				rig.Agents[e.Actor] = agent
			}
			agent.LastEvent = &e
			agent.LastUpdate = e.Time
		}
	}

	// Filter out events with empty bead IDs (malformed mutations)
	if e.Type == "update" && e.Target == "" {
		return
	}

	// Filter out noisy agent session updates from the event feed.
	// Agent session molecules (like gt-gastown-crew-joe) update frequently
	// for status tracking. These updates are visible in the agent tree,
	// so we don't need to clutter the event feed with them.
	// We still show create/complete/fail/delete events for agent sessions.
	if e.Type == "update" && beads.IsAgentSessionBead(e.Target) {
		// Skip adding to event feed, but still refresh the view
		// (agent tree was updated above)
		m.updateViewContent()
		return
	}

	// Deduplicate rapid updates to the same bead within 2 seconds.
	// This prevents spam when multiple deps/labels are added to one issue.
	if e.Type == "update" && e.Target != "" && len(m.events) > 0 {
		lastEvent := m.events[len(m.events)-1]
		if lastEvent.Type == "update" && lastEvent.Target == e.Target {
			// Same bead updated within 2 seconds - skip duplicate
			if e.Time.Sub(lastEvent.Time) < 2*time.Second {
				return
			}
		}
	}

	// Add to event feed
	m.events = append(m.events, e)

	// Keep max 1000 events
	if len(m.events) > 1000 {
		m.events = m.events[len(m.events)-1000:]
	}

	m.updateViewContent()
}

// SetEventChannel sets the channel to receive events from
func (m *Model) SetEventChannel(ch <-chan Event) {
	m.eventChan = ch
}

// View renders the TUI
func (m *Model) View() string {
	return m.render()
}

// maxTreeCursor returns the maximum valid cursor position in the tree panel
func (m *Model) maxTreeCursor() int {
	count := 0
	for _, rig := range m.rigs {
		byRole := m.groupAgentsByRole(rig.Agents)
		for _, role := range []string{"mayor", "witness", "refinery", "deacon", "crew", "polecat"} {
			agents, ok := byRole[role]
			if !ok || len(agents) == 0 {
				continue
			}
			// Each agent is a selectable item
			for _, agent := range agents {
				count++ // agent row
				if m.expandedAgents[agent.ID] {
					count++ // expanded details take one more line conceptually
				}
			}
		}
	}
	if count == 0 {
		return 0
	}
	return count - 1
}

// maxConvoyCursor returns the maximum valid cursor position in the convoy panel
func (m *Model) maxConvoyCursor() int {
	if m.convoyState == nil {
		return 0
	}
	count := 0
	// Count in-progress convoys
	for range m.convoyState.InProgress {
		count++
	}
	// Count landed convoys
	for range m.convoyState.Landed {
		count++
	}
	if count == 0 {
		return 0
	}
	return count - 1
}

// toggleTreeExpansion toggles expansion of the agent at the current cursor
func (m *Model) toggleTreeExpansion() {
	agent := m.getAgentAtCursor()
	if agent != nil {
		if m.expandedAgents[agent.ID] {
			delete(m.expandedAgents, agent.ID)
		} else {
			m.expandedAgents[agent.ID] = true
			// Fetch details if not cached
			if _, ok := m.agentDetailsCache[agent.ID]; !ok {
				details := m.fetchAgentDetails(agent)
				m.agentDetailsCache[agent.ID] = details
			}
		}
	}
}

// toggleConvoyExpansion toggles expansion of the convoy at the current cursor
func (m *Model) toggleConvoyExpansion() {
	convoy := m.getConvoyAtCursor()
	if convoy != nil {
		convoyID := convoy.ID
		if m.expandedConvoys[convoyID] {
			delete(m.expandedConvoys, convoyID)
		} else {
			m.expandedConvoys[convoyID] = true
			// Fetch details if not cached
			if _, ok := m.convoyDetailsCache[convoyID]; !ok {
				details := m.fetchConvoyDetails(convoy)
				m.convoyDetailsCache[convoyID] = details
			}
		}
	}
}

// getAgentAtCursor returns the agent at the current tree cursor position
func (m *Model) getAgentAtCursor() *Agent {
	pos := 0
	for _, rig := range m.rigs {
		byRole := m.groupAgentsByRole(rig.Agents)
		for _, role := range []string{"mayor", "witness", "refinery", "deacon", "crew", "polecat"} {
			agents, ok := byRole[role]
			if !ok || len(agents) == 0 {
				continue
			}
			for _, agent := range agents {
				if pos == m.treeCursor {
					return agent
				}
				pos++
				if m.expandedAgents[agent.ID] {
					pos++ // skip expanded details
				}
			}
		}
	}
	return nil
}

// getConvoyAtCursor returns the convoy at the current convoy cursor position
func (m *Model) getConvoyAtCursor() *Convoy {
	if m.convoyState == nil {
		return nil
	}
	pos := 0
	// Check in-progress convoys
	for i := range m.convoyState.InProgress {
		if pos == m.convoyCursor {
			return &m.convoyState.InProgress[i]
		}
		pos++
	}
	// Check landed convoys
	for i := range m.convoyState.Landed {
		if pos == m.convoyCursor {
			return &m.convoyState.Landed[i]
		}
		pos++
	}
	return nil
}

// fetchAgentDetails fetches detailed information about an agent
func (m *Model) fetchAgentDetails(agent *Agent) *AgentDetails {
	// For now, return basic info from the Agent struct
	// In a full implementation, this would query beads for full details
	return &AgentDetails{
		Name:       agent.Name,
		Role:       agent.Role,
		State:      agent.Status,
		Running:    agent.Status == "running" || agent.Status == "working",
		UpdatedAt:  agent.LastUpdate,
		WorkTitle:  "", // Would fetch from bead
		Branch:     "", // Would fetch from bead
		HookBead:   "", // Would fetch from bead
		UnreadMail: 0,  // Would fetch from bead
	}
}

// fetchConvoyDetails fetches detailed information about a convoy
func (m *Model) fetchConvoyDetails(convoy *Convoy) *ConvoyDetails {
	// For now, return basic info from the Convoy struct
	// In a full implementation, this would query beads for tracked issues
	return &ConvoyDetails{
		ID:            convoy.ID,
		Title:         convoy.Title,
		Status:        convoy.Status,
		Completed:     convoy.Completed,
		Total:         convoy.Total,
		TrackedIssues: []TrackedIssue{}, // Would fetch from beads
		CreatedAt:     convoy.CreatedAt,
		ClosedAt:      convoy.ClosedAt,
		LastActivity:  time.Now(), // Would calculate from events
	}
}
