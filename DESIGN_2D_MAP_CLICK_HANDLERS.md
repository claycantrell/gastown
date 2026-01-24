# Design: Click Handlers for Buildings and Polecats on 2D Map

**Date:** 2026-01-23
**Status:** Design Phase
**Issue:** gt-wisp-1rqi

## Overview

This design adds interactive 2D map visualization to Gas Town, allowing users to click on buildings (convoys) and polecats (workers) to view details and perform actions.

## Current State

The codebase has two visualization approaches:
1. **Web Dashboard** (`/internal/web/`) - Table-based HTML with HTMX auto-refresh
2. **TUI** (`/internal/tui/feed/`) - Terminal UI using Bubble Tea framework

Neither provides spatial/2D visualization. All displays are text-based (tables or lists).

## Proposed Architecture

### 1. Technology Choice: Web-Based SVG Map

**Rationale:**
- Existing web infrastructure (`/internal/web/`)
- Native mouse event support
- SVG provides scalable, interactive graphics
- Consistent with current HTML/HTMX approach
- Easier than terminal-based graphics

**Alternative Considered:** Terminal UI with mouse support - rejected due to complexity and limited spatial representation.

### 2. Component Structure

```
/internal/web/
├── handler.go          [MODIFY] Add /map endpoint
├── templates.go        [MODIFY] Add MapData struct
├── fetcher.go          [REUSE] Existing data fetching
└── templates/
    ├── convoy.html     [EXISTING]
    └── map.html        [NEW] 2D map view
```

### 3. Data Model

```go
// Add to internal/web/templates.go

type MapData struct {
    Buildings []BuildingPosition
    Polecats  []PolecatPosition
    Timestamp time.Time
}

type BuildingPosition struct {
    ID       string  // Convoy ID
    Title    string
    Status   string  // "open", "closed"
    X        float64 // Map coordinates (0-100%)
    Y        float64
    Progress int     // % completion
}

type PolecatPosition struct {
    Name      string  // e.g., "nux", "dag"
    Rig       string
    SessionID string
    X         float64
    Y         float64
    Status    string  // "active", "idle", "stuck"
}
```

### 4. Positioning Strategy

**Problem:** Convoy/polecat data has no inherent spatial properties.

**Solution:** Auto-layout algorithm based on status/activity

**Buildings (Convoys):**
- Group by status: open (left), closed (right)
- Vertical position by last activity (recent = top)
- Spacing proportional to number of items

**Polecats:**
- Position near assigned convoy (if hooked)
- Otherwise, position in "idle zone" at bottom
- Avoid overlap with proximity-based jitter

**Future Enhancement:** Allow drag-to-reposition with position persistence.

### 5. Click Handler Implementation

**Frontend (SVG + HTMX):**

```html
<!-- templates/map.html -->
<svg id="map-canvas" width="100%" height="600" viewBox="0 0 100 100">
    <!-- Buildings -->
    {{range .Buildings}}
    <g class="building"
       data-id="{{.ID}}"
       hx-get="/map/building/{{.ID}}"
       hx-target="#detail-panel"
       hx-trigger="click">
        <circle cx="{{.X}}" cy="{{.Y}}" r="2"
                class="building-{{.Status}}" />
        <text x="{{.X}}" y="{{.Y}}" dy="-2.5">{{.Title}}</text>
    </g>
    {{end}}

    <!-- Polecats -->
    {{range .Polecats}}
    <g class="polecat"
       data-id="{{.SessionID}}"
       hx-get="/map/polecat/{{.SessionID}}"
       hx-target="#detail-panel"
       hx-trigger="click">
        <rect x="{{.X}}" y="{{.Y}}" width="1.5" height="1.5"
              class="polecat-{{.Status}}" />
        <text x="{{.X}}" y="{{.Y}}" dy="-2">{{.Name}}</text>
    </g>
    {{end}}
</svg>

<div id="detail-panel">
    <!-- Click handler populates this via HTMX -->
</div>
```

**Backend (Go Handlers):**

```go
// Add to internal/web/handler.go

func (h *ConvoyHandler) HandleMapView(w http.ResponseWriter, r *http.Request) {
    data := h.fetcher.FetchMapData()
    tmpl.ExecuteTemplate(w, "map.html", data)
}

func (h *ConvoyHandler) HandleBuildingClick(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    // Fetch full convoy details
    convoy := h.fetcher.FetchConvoyDetails(id)

    // Render detail partial
    tmpl.ExecuteTemplate(w, "building-detail.html", convoy)
}

func (h *ConvoyHandler) HandlePolecatClick(w http.ResponseWriter, r *http.Request) {
    sessionID := chi.URLParam(r, "sessionID")

    // Fetch polecat status
    polecat := h.fetcher.FetchPolecatDetails(sessionID)

    // Render detail partial
    tmpl.ExecuteTemplate(w, "polecat-detail.html", polecat)
}
```

### 6. Integration Points

**Existing Code Reuse:**
- `LiveConvoyFetcher` in `/internal/web/fetcher.go` - Already fetches convoy/polecat data
- HTTP router setup in `/internal/cmd/dashboard.go` - Add new routes
- Template rendering in `templates.go` - Add new template structs

**New Code Required:**
1. Map positioning algorithm (`/internal/web/layout.go` - NEW)
2. Map template (`/templates/map.html` - NEW)
3. Detail panel templates (`building-detail.html`, `polecat-detail.html` - NEW)
4. Route handlers for clicks (modify `handler.go`)

### 7. Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| No convoys exist | Display empty map with message |
| No polecats active | Show only buildings |
| Many items overlap | Apply jitter/spacing algorithm |
| Click on empty space | Clear detail panel |
| Data fetch fails | Show last known state + error banner |
| Convoy deleted while viewing | Detail panel shows "Not found" |
| Stale data | HTMX auto-refresh every 10s (like existing dashboard) |

### 8. Simpler Alternative Approach

**Minimal Implementation:**
- Skip auto-layout; use fixed grid positions
- Buildings in top half, polecats in bottom half
- Simple click → full page navigation (no HTMX panels)

**Rejected because:**
- Fixed grid doesn't convey status relationships
- Full page reload is jarring
- Deviates from existing HTMX patterns

### 9. Testing Strategy

**Unit Tests:**
- Layout algorithm: verify no overlaps, boundaries
- Position calculation: edge cases (0 items, 100 items)

**Integration Tests:**
- HTTP handlers: verify correct data passed to templates
- Click endpoints: verify detail fetching

**Manual Testing:**
- Visual verification of layout
- Click responsiveness
- Multiple simultaneous clicks
- Browser compatibility (Chrome, Firefox, Safari)

### 10. Rollout Plan

**Phase 1 (MVP):**
- Basic SVG map with fixed layout
- Click handlers for buildings only
- Detail panel with convoy info

**Phase 2:**
- Add polecat visualization
- Polecat click handlers
- Auto-layout algorithm

**Phase 3:**
- Real-time updates (WebSocket)
- Drag-to-reposition
- Zoom/pan controls

### 11. Open Questions

1. **Should the map be a separate page or integrate into convoy.html?**
   - Recommendation: Separate page initially (`/map`), link from dashboard header

2. **How to handle mobile/small screens?**
   - Recommendation: Responsive SVG viewBox, pinch-to-zoom

3. **What actions should be available on click?**
   - Recommendation: View details only (Phase 1), add actions later (e.g., "Assign polecat", "Restart convoy")

## Summary

This design adds a web-based SVG map with click handlers following existing HTMX patterns. The implementation:
- Reuses existing data fetchers
- Follows established web handler conventions
- Handles edge cases gracefully
- Provides clear upgrade path for future enhancements

**Estimated Complexity:** Medium
- New code: ~300 lines (layout algorithm, templates, handlers)
- Modified code: ~50 lines (routing, template struct)
- Test code: ~200 lines

**Risk Areas:**
- Layout algorithm complexity (overlap handling)
- Performance with many items (100+ convoys)
- Browser SVG compatibility

**Mitigation:**
- Start with simple grid layout, iterate
- Limit initial render to 50 items, add pagination
- Test on multiple browsers early
