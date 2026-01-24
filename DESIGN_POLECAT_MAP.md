# Design: Live Polecat Position Tracking on 2D Desert Map

## Overview
Add a real-time 2D visual map to the Gas Town web dashboard showing polecat workers as they move through different work states across rigs, using a desert theme metaphor.

## Context
Currently, polecats (ephemeral AI worker agents) are displayed in table format showing name, rig, and activity status. There is no spatial visualization. This feature introduces a 2D "desert map" view where polecats appear as moving entities on a canvas.

## Goals
1. Visualize polecat distribution across rigs spatially
2. Show work state progression as movement on the map
3. Provide real-time updates as polecats change state
4. Embrace the "desert" theme as a visual metaphor

## Non-Goals
- Terminal/TUI version (web only for initial implementation)
- Interactive controls (click to view details, drag, etc.)
- Historical position replay

## Design

### Coordinate System
Map polecats to 2D coordinates based on existing state data:

**X-axis: Rig Distribution**
- Each rig occupies a horizontal zone
- Polecats within a rig scattered around the zone center
- Example: Rig "gastown" at x=200, "beads" at x=600

**Y-axis: Work State Progression**
```
Y=100:  Ready/Idle zone (top)
Y=300:  Working zone (middle)
Y=500:  Done zone (bottom)
```

**Position Calculation Algorithm:**
```go
type Position struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

func CalculatePosition(polecat Polecat, rigIndex int) Position {
    // X: Rig spacing (200px apart) + random scatter
    baseX := float64(rigIndex * 200 + 100)
    scatterX := rand.Float64()*60 - 30  // ±30px

    // Y: State-based vertical position
    stateY := map[State]float64{
        StateWorking: 300,
        StateDone:    500,
        StateStuck:   150,  // Stuck = between ready and working
    }
    baseY := stateY[polecat.State]
    scatterY := rand.Float64()*40 - 20  // ±20px

    return Position{
        X: baseX + scatterX,
        Y: baseY + scatterY,
    }
}
```

### Visual Design
```
┌──────────────────────────────────────────────────────────┐
│  🌵 Desert Map - Live Polecat Tracking                  │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  🌵 gastown          🌵 beads           🌵 convoy       │
│                                                          │
│     Ready Zone (Idle)                                   │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─           │
│                                                          │
│    🦨 Toast          🦨 Amber                           │
│     ●green           ●green                             │
│                                                          │
│     Working Zone                                        │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─           │
│                                                          │
│    🦨 Shadow         🦨 Oak                             │
│     ●yellow          ●green                             │
│                                                          │
│     Done Zone                                           │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─           │
│                                                          │
│    🦨 Copper                                            │
│     ●red                                                │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

**Visual Elements:**
- **Polecat sprite:** Circle or icon with name label
- **Activity color:** Reuse existing green/yellow/red from activity.go
- **Rig markers:** Cactus emoji or label at top
- **Zone lines:** Dashed horizontal separators
- **Smooth transitions:** Animate position changes over 500ms

### Technology Stack

**Frontend:**
- **Canvas API** for 2D rendering (lighter than SVG for animations)
- **htmx** for data updates (reuse existing 10s polling)
- **Vanilla JS** for rendering (no framework needed)

**Backend:**
- Extend `internal/web/fetcher.go:FetchPolecats()` to include positions
- Add `CalculatePosition()` function in new `internal/polecat/position.go`
- Update `PolecatRow` struct:
```go
type PolecatRow struct {
    Name         string
    Rig          string
    SessionID    string
    LastActivity activity.Info
    StatusHint   string
    Position     Position  // NEW
}
```

### Live Updates

**Mechanism:** Reuse existing htmx 10-second polling
```html
<div id="desert-map"
     hx-get="/api/polecat-positions"
     hx-trigger="every 10s"
     hx-swap="none">
    <canvas id="map-canvas" width="800" height="600"></canvas>
</div>
```

**Update Flow:**
1. htmx fetches `/api/polecat-positions` every 10s
2. Response contains JSON with polecat positions
3. JavaScript updates canvas with smooth transitions

**Future Enhancement:** WebSocket for sub-second updates (out of scope for v1)

### File Changes

**New Files:**
- `internal/polecat/position.go` - Position calculation logic
- `internal/web/templates/map.html` - Map view template
- `internal/web/static/map.js` - Canvas rendering JS
- `internal/web/static/map.css` - Map styling

**Modified Files:**
- `internal/web/fetcher.go` - Add position data to PolecatRow
- `internal/web/handlers.go` - Add `/map` route and `/api/polecat-positions` endpoint
- `internal/web/templates/convoy.html` - Add link to map view

### Edge Cases

1. **No polecats:** Show empty desert with "No active polecats" message
2. **Many polecats (>20):** Increase scatter radius to avoid overlap
3. **New rig appears:** Dynamically calculate X position based on rig count
4. **Polecat state changes:** Smoothly animate from old Y to new Y position
5. **Stalled polecats:** Position between ready and working zones (Y=150)
6. **Name collisions:** Add small vertical offset if two polecats have same X/Y

### Testing Strategy

**Unit Tests:**
- `position_test.go`: Test coordinate calculation for all states
- Test scatter doesn't exceed zone boundaries
- Test rig indexing with 1, 5, 10 rigs

**Integration Tests:**
- Verify `/api/polecat-positions` returns valid JSON
- Test htmx updates trigger canvas re-render
- Verify activity colors match existing logic

**Manual Testing:**
- Spawn multiple polecats across different rigs
- Watch them move as states change (working → done)
- Verify activity colors (green/yellow/red) appear correctly
- Test with 0, 1, 5, 10+ polecats

### Migration/Rollout

**Phase 1:** Add map view as separate page
- No changes to existing convoy table view
- Link from dashboard: "🗺️ View Desert Map"
- Low risk: purely additive

**Phase 2:** (Optional) Replace or augment table view
- User preference toggle between table/map
- Or split-screen showing both

### Alternative Approaches Considered

**1. Force-Directed Graph Layout**
- Pros: Automatic spacing, looks organic
- Cons: Non-deterministic positions, harder to read states
- **Rejected:** Prefer explicit state-based positioning

**2. Terminal TUI Version**
- Pros: Consistent with existing TUI
- Cons: Limited canvas API in terminal, harder animations
- **Deferred:** Web-first, TUI later if valuable

**3. WebSocket for Real-Time Updates**
- Pros: Instant updates (<10s latency)
- Cons: More complex infrastructure, connection management
- **Deferred:** Start with polling, upgrade if needed

**4. 3D Visualization**
- Pros: Cool factor, Z-axis for third dimension
- Cons: Overkill for current needs, harder to read
- **Rejected:** Keep it simple with 2D

## Success Metrics

1. Map renders correctly with 0-20 polecats
2. Positions update within 10 seconds of state changes
3. Activity colors match existing dashboard
4. No performance degradation on dashboard load
5. Visual map provides quicker understanding of polecat distribution than table

## Open Questions

1. **Should the map be on the main dashboard or a separate page?**
   - Proposal: Separate page initially, embed if popular

2. **Should we show historical trails (where polecats have been)?**
   - Proposal: No for v1, consider for v2

3. **Should polecats "walk" between positions or teleport?**
   - Proposal: Smooth CSS transition (500ms ease-in-out)

4. **What happens if a rig has 10+ polecats?**
   - Proposal: Increase scatter radius dynamically, or stack vertically

## Timeline

This is a straightforward feature building on existing infrastructure:
- Position calculation: ~50 lines of Go
- API endpoint: ~30 lines
- Canvas rendering: ~150 lines of JS
- Template: ~50 lines of HTML/CSS

No external dependencies, no database schema changes, purely additive.
