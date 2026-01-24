package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// convoyIDPattern validates convoy IDs to prevent SQL injection
var convoyIDPattern = regexp.MustCompile(`^hq-[a-zA-Z0-9-]+$`)

// convoySubprocessTimeout is the timeout for bd and sqlite3 calls in the convoy panel.
// Prevents TUI freezing if these commands hang.
const convoySubprocessTimeout = 5 * time.Second

// Convoy represents a convoy's status for the dashboard
type Convoy struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	ClosedAt  time.Time `json:"closed_at,omitempty"`
}

// ConvoyState holds all convoy data for the panel
type ConvoyState struct {
	InProgress []Convoy
	Landed     []Convoy
	LastUpdate time.Time
}

// FetchConvoys retrieves convoy status from town-level beads
func FetchConvoys(townRoot string) (*ConvoyState, error) {
	townBeads := filepath.Join(townRoot, ".beads")

	state := &ConvoyState{
		InProgress: make([]Convoy, 0),
		Landed:     make([]Convoy, 0),
		LastUpdate: time.Now(),
	}

	// Fetch open convoys
	openConvoys, err := listConvoys(townBeads, "open")
	if err != nil {
		// Not a fatal error - just return empty state
		return state, nil
	}

	for _, c := range openConvoys {
		// Get detailed status for each convoy
		convoy := enrichConvoy(townBeads, c)
		state.InProgress = append(state.InProgress, convoy)
	}

	// Fetch recently closed convoys (landed in last 24h)
	closedConvoys, err := listConvoys(townBeads, "closed")
	if err == nil {
		cutoff := time.Now().Add(-24 * time.Hour)
		for _, c := range closedConvoys {
			convoy := enrichConvoy(townBeads, c)
			if !convoy.ClosedAt.IsZero() && convoy.ClosedAt.After(cutoff) {
				state.Landed = append(state.Landed, convoy)
			}
		}
	}

	// Sort: in-progress by created (oldest first), landed by closed (newest first)
	sort.Slice(state.InProgress, func(i, j int) bool {
		return state.InProgress[i].CreatedAt.Before(state.InProgress[j].CreatedAt)
	})
	sort.Slice(state.Landed, func(i, j int) bool {
		return state.Landed[i].ClosedAt.After(state.Landed[j].ClosedAt)
	})

	return state, nil
}

// listConvoys returns convoys with the given status
func listConvoys(beadsDir, status string) ([]convoyListItem, error) {
	listArgs := []string{"list", "--type=convoy", "--status=" + status, "--json"}

	ctx, cancel := context.WithTimeout(context.Background(), convoySubprocessTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", listArgs...) //nolint:gosec // G204: args are constructed internally
	cmd.Dir = beadsDir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var items []convoyListItem
	if err := json.Unmarshal(stdout.Bytes(), &items); err != nil {
		return nil, err
	}

	return items, nil
}

type convoyListItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	ClosedAt  string `json:"closed_at,omitempty"`
}

// enrichConvoy adds tracked issue counts to a convoy
func enrichConvoy(beadsDir string, item convoyListItem) Convoy {
	convoy := Convoy{
		ID:     item.ID,
		Title:  item.Title,
		Status: item.Status,
	}

	// Parse timestamps
	if t, err := time.Parse(time.RFC3339, item.CreatedAt); err == nil {
		convoy.CreatedAt = t
	} else if t, err := time.Parse("2006-01-02 15:04", item.CreatedAt); err == nil {
		convoy.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, item.ClosedAt); err == nil {
		convoy.ClosedAt = t
	} else if t, err := time.Parse("2006-01-02 15:04", item.ClosedAt); err == nil {
		convoy.ClosedAt = t
	}

	// Get tracked issues and their status
	tracked := getTrackedIssueStatus(beadsDir, item.ID)
	convoy.Total = len(tracked)
	for _, t := range tracked {
		if t.Status == "closed" {
			convoy.Completed++
		}
	}

	return convoy
}

type trackedStatus struct {
	ID     string
	Status string
}

// getTrackedIssueStatus queries tracked issues and their status
func getTrackedIssueStatus(beadsDir, convoyID string) []trackedStatus {
	// Validate convoyID to prevent SQL injection
	if !convoyIDPattern.MatchString(convoyID) {
		return nil
	}

	dbPath := filepath.Join(beadsDir, "beads.db")

	ctx, cancel := context.WithTimeout(context.Background(), convoySubprocessTimeout)
	defer cancel()

	// Query tracked dependencies from SQLite
	// convoyID is validated above to match ^hq-[a-zA-Z0-9-]+$
	cmd := exec.CommandContext(ctx, "sqlite3", "-json", dbPath, //nolint:gosec // G204: convoyID is validated against strict pattern
		fmt.Sprintf(`SELECT depends_on_id FROM dependencies WHERE issue_id = '%s' AND type = 'tracks'`, convoyID))

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil
	}

	var deps []struct {
		DependsOnID string `json:"depends_on_id"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &deps); err != nil {
		return nil
	}

	var tracked []trackedStatus
	for _, dep := range deps {
		issueID := dep.DependsOnID

		// Handle external reference format: external:rig:issue-id
		if strings.HasPrefix(issueID, "external:") {
			parts := strings.SplitN(issueID, ":", 3)
			if len(parts) == 3 {
				issueID = parts[2]
			}
		}

		// Get issue status
		status := getIssueStatus(issueID)
		tracked = append(tracked, trackedStatus{ID: issueID, Status: status})
	}

	return tracked
}

// getIssueStatus fetches just the status of an issue
func getIssueStatus(issueID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), convoySubprocessTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", "show", issueID, "--json")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "unknown"
	}

	var issues []struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &issues); err != nil || len(issues) == 0 {
		return "unknown"
	}

	return issues[0].Status
}

// Convoy panel styles
var (
	ConvoyPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorDim).
				Padding(0, 1)

	ConvoyTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary)

	ConvoySectionStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Bold(true)

	ConvoyIDStyle = lipgloss.NewStyle().
			Foreground(colorHighlight)

	ConvoyNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	ConvoyProgressStyle = lipgloss.NewStyle().
				Foreground(colorSuccess)

	ConvoyLandedStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	ConvoyAgeStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

// renderConvoyPanel renders the convoy status panel
func (m *Model) renderConvoyPanel() string {
	style := ConvoyPanelStyle
	if m.focusedPanel == PanelConvoy {
		style = FocusedBorderStyle
	}
	// Add title before content
	title := ConvoyTitleStyle.Render("🚚 Convoys")
	content := title + "\n" + m.convoyViewport.View()
	return style.Width(m.width - 2).Render(content)
}

// renderConvoys renders the convoy panel content
func (m *Model) renderConvoys() string {
	if m.convoyState == nil {
		return AgentIdleStyle.Render("Loading convoys...")
	}

	var lines []string
	pos := 0

	// In Progress section
	lines = append(lines, ConvoySectionStyle.Render("IN PROGRESS"))
	if len(m.convoyState.InProgress) == 0 {
		lines = append(lines, "  "+AgentIdleStyle.Render("No active convoys"))
	} else {
		for _, c := range m.convoyState.InProgress {
			lines = append(lines, m.renderConvoyLine(c, false, pos))
			pos++

			// Render expanded details if convoy is expanded
			if m.expandedConvoys[c.ID] {
				details := m.renderConvoyDetails(c.ID)
				if details != "" {
					lines = append(lines, details)
				}
			}
		}
	}

	lines = append(lines, "")

	// Recently Landed section
	lines = append(lines, ConvoySectionStyle.Render("RECENTLY LANDED (24h)"))
	if len(m.convoyState.Landed) == 0 {
		lines = append(lines, "  "+AgentIdleStyle.Render("No recent landings"))
	} else {
		for _, c := range m.convoyState.Landed {
			lines = append(lines, m.renderConvoyLine(c, true, pos))
			pos++

			// Render expanded details if convoy is expanded
			if m.expandedConvoys[c.ID] {
				details := m.renderConvoyDetails(c.ID)
				if details != "" {
					lines = append(lines, details)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// renderConvoyLine renders a single convoy status line
func (m *Model) renderConvoyLine(c Convoy, landed bool, pos int) string {
	// Format: "  hq-xyz  Title       2/4 ●●○○" or "  hq-xyz  Title       ✓ 2h ago"

	// Expand/collapse indicator
	expandIcon := ""
	if m.focusedPanel == PanelConvoy {
		if m.expandedConvoys[c.ID] {
			expandIcon = "▼ "
		} else {
			expandIcon = "▶ "
		}
	}

	id := ConvoyIDStyle.Render(c.ID)

	// Truncate title if too long
	title := c.Title
	titleMaxLen := 20
	if m.focusedPanel == PanelConvoy {
		titleMaxLen = 18 // Make room for expand icon
	}
	if len(title) > titleMaxLen {
		title = title[:titleMaxLen-3] + "..."
	}

	var line string
	if landed {
		// Show checkmark and time since landing
		age := formatAge(time.Since(c.ClosedAt))
		status := ConvoyLandedStyle.Render("✓") + " " + ConvoyAgeStyle.Render(age+" ago")
		line = fmt.Sprintf("  %s%s  %-20s  %s", expandIcon, id, title, status)
	} else {
		// Show progress bar
		progress := renderProgressBar(c.Completed, c.Total)
		count := ConvoyProgressStyle.Render(fmt.Sprintf("%d/%d", c.Completed, c.Total))
		line = fmt.Sprintf("  %s%s  %-20s  %s %s", expandIcon, id, title, count, progress)
	}

	// Apply selection style if this convoy is selected
	if m.focusedPanel == PanelConvoy && pos == m.convoyCursor {
		return SelectedStyle.Render(line)
	}

	return line
}

// renderConvoyDetails renders expanded details for a convoy
func (m *Model) renderConvoyDetails(convoyID string) string {
	details, ok := m.convoyDetailsCache[convoyID]
	if !ok {
		return ""
	}

	var lines []string
	baseIndent := "    "

	// Status
	lines = append(lines, baseIndent+DetailKeyStyle.Render("Status: ")+DetailValueStyle.Render(details.Status))

	// Created/Closed times
	if !details.CreatedAt.IsZero() {
		age := formatAge(time.Since(details.CreatedAt))
		lines = append(lines, baseIndent+DetailKeyStyle.Render("Created: ")+DetailValueDimStyle.Render(age+" ago"))
	}
	if !details.ClosedAt.IsZero() {
		age := formatAge(time.Since(details.ClosedAt))
		lines = append(lines, baseIndent+DetailKeyStyle.Render("Closed: ")+DetailValueDimStyle.Render(age+" ago"))
	}

	// Progress
	percentage := 0
	if details.Total > 0 {
		percentage = (details.Completed * 100) / details.Total
	}
	lines = append(lines, baseIndent+DetailKeyStyle.Render("Progress: ")+DetailValueStyle.Render(fmt.Sprintf("%d of %d issues completed (%d%%)", details.Completed, details.Total, percentage)))

	// Tracked issues (show first few)
	if len(details.TrackedIssues) > 0 {
		lines = append(lines, baseIndent+DetailKeyStyle.Render("Tracked Issues:"))
		maxIssues := len(details.TrackedIssues)
		if maxIssues > 5 {
			maxIssues = 5 // Limit to first 5
		}
		for i := 0; i < maxIssues; i++ {
			issue := details.TrackedIssues[i]
			connector := "├─"
			if i == maxIssues-1 && len(details.TrackedIssues) <= 5 {
				connector = "└─"
			}

			statusIcon := "○"
			statusStyle := issueOpenStyle
			if issue.Status == "closed" {
				statusIcon = "✓"
				statusStyle = issueClosedStyle
			} else if issue.Status == "in_progress" {
				statusIcon = "▶"
			}

			issueTitle := issue.Title
			if len(issueTitle) > 40 {
				issueTitle = issueTitle[:37] + "..."
			}

			workerInfo := ""
			if issue.Worker != "" {
				workerInfo = fmt.Sprintf(" @%s", issue.Worker)
				if issue.WorkerAge != "" {
					workerInfo += fmt.Sprintf(" (%s)", issue.WorkerAge)
				}
			}

			issueLine := fmt.Sprintf("%s  %s %s %s: %s%s", baseIndent, connector, statusIcon, issue.ID, issueTitle, workerInfo)
			lines = append(lines, statusStyle.Render(issueLine))
		}

		if len(details.TrackedIssues) > 5 {
			lines = append(lines, baseIndent+DetailValueDimStyle.Render(fmt.Sprintf("  ... and %d more", len(details.TrackedIssues)-5)))
		}
	}

	return strings.Join(lines, "\n")
}

var (
	issueOpenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")) // yellow

	issueClosedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")) // green
)

// renderProgressBar creates a simple progress bar: ●●○○
func renderProgressBar(completed, total int) string {
	if total == 0 {
		return ""
	}

	// Cap at 5 dots for display
	displayTotal := total
	if displayTotal > 5 {
		displayTotal = 5
	}

	filled := (completed * displayTotal) / total
	if filled > displayTotal {
		filled = displayTotal
	}

	bar := strings.Repeat("●", filled) + strings.Repeat("○", displayTotal-filled)
	return ConvoyProgressStyle.Render(bar)
}

