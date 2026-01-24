# Design: Tooltips for Polecat and Convoy Status

## Overview
Add detailed status information for polecats and convoys that users can access on-demand within the Feed TUI and web dashboard.

## Context
- **Codebase**: Go CLI/TUI tool using Bubble Tea, tmux, Lipgloss
- **Current State**: Basic status shown inline (name, state, work title)
- **User Need**: See detailed status without leaving the current view
- **No traditional tooltips**: TUI environment requires keyboard-driven interactions

## Design Decision: Inline Expansion Pattern

### Rationale
Use the existing expand/collapse pattern from the Convoy TUI (`internal/tui/convoy/`) for consistency and discoverability.

**Why this approach:**
1. ✅ Already familiar to users (used in convoy tree view)
2. ✅ Works in TUI environment (no mouse hover needed)
3. ✅ Shows multi-line details without new windows
4. ✅ Minimal code changes - reuse existing patterns
5. ✅ Works for both polecats and convoys

**Alternatives considered:**
- ❌ Status bar context: Too limited (single line)
- ❌ tmux popup: Requires extra key, feels heavyweight
- ❌ Verbose flag: Requires rerunning command

## Implementation Design

### 1. Feed TUI Changes

#### A. Polecat Panel - Expandable Items

**File**: `internal/tui/feed/model.go`

**New state tracking:**
```go
type Model struct {
    // ... existing fields ...
    expandedPolecats map[string]bool  // Track which polecats are expanded
    expandedConvoys  map[string]bool  // Track which convoys are expanded
}
```

**Keyboard bindings:**
- `Enter` or `Space`: Toggle expansion on selected item
- `i`: Show info (alias for expand)
- `Escape` or `q`: Collapse if expanded
- `j/k` or arrow keys: Navigate (existing)

#### B. Expanded Polecat Display

**Format** (when expanded with Enter/Space/i):
```
🦨 Polecats
  ▼ toast  ● → Fix auth bug           📬2
    ├─ Status: Working (tmux running)
    ├─ Hook: gt-wisp-abc12 → Fix authentication bug in OAuth flow
    ├─ Branch: feat/oauth-provider
    ├─ Git State: Clean worktree (safe to remove)
    ├─ Unread Mail: 2 messages
    │  └─ First: "RE: Need help with auth tests"
    ├─ Created: 2h ago
    └─ Last Activity: 5m ago
  ▶ goose  ○
  ▶ nux    ● stuck → Refinery build
```

**Data sources:**
- From `AgentRuntime` struct (status.go):
  - `Name`, `Running`, `State`, `WorkTitle`, `HookBead`, `UnreadMail`, `FirstSubject`
- From `Polecat` struct (polecat/types.go):
  - `Branch`, `CleanupStatus`, `CreatedAt`, `UpdatedAt`
- Combine both via bead lookup in feed model

**Visual indicators:**
- `▼` = expanded, `▶` = collapsed (clickable/toggleable)
- Indented tree structure using `├─` and `└─`
- Truncate long text with ellipsis (e.g., work title max 50 chars in expanded view)

#### C. Expanded Convoy Display

**Format**:
```
IN PROGRESS
  ▼ hq-cv-abc: Deploy v2.0 (2/5) ●●○○○
    ├─ Status: in_progress
    ├─ Created: 3h ago
    ├─ Progress: 2 of 5 issues completed (40%)
    ├─ Tracked Issues:
    │  ├─ ✓ gt-xyz: Add OAuth provider [crew/joe] @joe
    │  ├─ ▶ gt-abc: Update API docs [feature] @toast (12m)
    │  ├─ ○ gt-def: Fix bug [bugfix]
    │  ├─ ○ gt-ghi: Write tests [test]
    │  └─ ○ gt-jkl: Update changelog [docs]
    ├─ Workers: 2 active (toast @gt-abc, joe @gt-xyz)
    └─ Last Activity: 5m ago
  ▶ hq-cv-xyz: Feature rollout (1/4)
```

**Data sources:**
- From convoy tracking (`convoy.go:trackedIssueInfo`):
  - Issue ID, title, status, worker, worker age
- From convoy metadata:
  - Created/closed timestamps, progress counts

**Visual indicators:**
- ✓ = completed issue, ▶ = in progress, ○ = open
- Progress bar: `●●○○○` (filled/unfilled circles)
- Worker age in parentheses when active (e.g., `@toast (12m)`)

### 2. Implementation Steps

**Phase 1: Core Data Structures**
1. Add `expandedPolecats` and `expandedConvoys` maps to Feed model
2. Add expansion state to `tree.Node` or create wrapper struct
3. Create helper functions:
   - `fetchPolecatDetails(name string) (*DetailedPolecatInfo, error)`
   - `fetchConvoyDetails(id string) (*DetailedConvoyInfo, error)`

**Phase 2: Rendering**
1. Update `renderTreePanel()` to check expansion state
2. Create `renderExpandedPolecat(p *DetailedPolecatInfo) string`
3. Create `renderExpandedConvoy(c *DetailedConvoyInfo) string`
4. Add indentation and tree characters using Lipgloss

**Phase 3: Keyboard Handling**
1. Update `Update()` to handle Enter/Space/i keys
2. Toggle expansion state in map
3. Trigger data fetch if not cached
4. Update view to show expanded content

**Phase 4: Styling**
1. Define styles in `internal/tui/feed/styles.go`:
   - `ExpandedContentStyle` (dimmed, indented)
   - `DetailKeyStyle` (for labels like "Status:")
   - `DetailValueStyle` (for values)
2. Use existing color palette from `internal/ui/styles.go`

### 3. Web Dashboard Enhancement

**File**: `internal/web/templates/convoy.html`

Add native HTML tooltips using `title` attribute or custom CSS tooltips:

**Option A: Native tooltips** (simplest):
```html
<tr title="Status: {{ .Status }}&#10;Created: {{ .CreatedAt }}&#10;Progress: {{ .Completed }}/{{ .Total }}">
  <td>{{ .ID }}</td>
  <td>{{ .Title }}</td>
  ...
</tr>
```

**Option B: Custom CSS tooltips** (better styling):
```html
<style>
.tooltip {
  position: relative;
  cursor: help;
}
.tooltip::after {
  content: attr(data-tooltip);
  position: absolute;
  background: #1a1a1a;
  color: #fff;
  padding: 8px;
  border-radius: 4px;
  white-space: pre-line;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s;
}
.tooltip:hover::after {
  opacity: 1;
}
</style>

<tr class="tooltip" data-tooltip="Status: {{ .Status }}
Created: {{ .CreatedAt }}
Progress: {{ .Completed }}/{{ .Total }}">
  ...
</tr>
```

### 4. Edge Cases & Considerations

**Performance:**
- Fetch detailed data lazily (only when expanded)
- Cache expanded data for 30s to avoid repeated bead lookups
- Limit to 1-2 expanded items at a time (auto-collapse others)

**Long Content:**
- Truncate issue titles to 60 chars in expanded view
- Limit tracked issues display to 10 (show "... N more")
- Wrap long git branches with ellipsis

**Error Handling:**
- If bead lookup fails: Show "⚠ Could not fetch details"
- If data is stale (>5min): Show age indicator "(stale)"

**Accessibility:**
- Ensure keyboard navigation works smoothly
- Use clear visual indicators (▼/▶) for expansion state
- Provide help text in TUI footer: "Enter: expand | Esc: collapse"

### 5. Files to Modify

**Core Implementation:**
1. `internal/tui/feed/model.go` - Add expansion state, update handlers
2. `internal/tui/feed/view.go` - Rendering logic for expanded items
3. `internal/tui/feed/styles.go` - New styles for detailed content
4. `internal/tui/feed/update.go` - Keyboard handling for expansion

**Data Fetching:**
5. `internal/cmd/status.go` - Export or create helper to fetch detailed agent info
6. `internal/cmd/convoy.go` - Export helper to fetch convoy details

**Web Dashboard:**
7. `internal/web/templates/convoy.html` - Add tooltips to HTML

**Testing:**
8. `internal/tui/feed/model_test.go` - Test expansion state logic
9. Manual testing in live TUI

## Success Criteria

1. ✅ Users can press Enter/Space on any polecat or convoy in Feed TUI to see details
2. ✅ Expanded view shows all relevant status information inline
3. ✅ Expansion state toggles (press again to collapse)
4. ✅ Visual consistency with existing convoy tree UI
5. ✅ Web dashboard shows tooltips on hover
6. ✅ No performance degradation (lazy loading)
7. ✅ Help text in footer explains new feature

## Out of Scope

- ❌ Hover tooltips in TUI (not possible in terminal)
- ❌ Animated transitions (limited in Bubble Tea)
- ❌ Mouse click to expand (keyboard-focused design)
- ❌ Persistent expansion across TUI refreshes (state resets intentionally)

## Testing Plan

1. **Unit Tests**: Expansion state toggling logic
2. **Integration Tests**: Data fetching from beads
3. **Manual Testing**:
   - Expand/collapse polecats with various states (working, stuck, done)
   - Expand convoys with different progress levels
   - Navigate with keyboard while items are expanded
   - Verify truncation of long content
   - Test web dashboard tooltips in browser

## Future Enhancements

- Context menu on expanded items (e.g., "Jump to session", "Send mail")
- Filtering: Show only expanded items
- Export expanded view to markdown/JSON
- Clickable links to jump to tmux session or open convoy details
