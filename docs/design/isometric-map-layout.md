# Isometric Gas Town Map Layout Design

**Issue:** hq-ftc3.3
**Author:** Polecat Rictus
**Date:** 2026-01-23
**Status:** Design Specification

## Overview

This document specifies the design for an isometric Gas Town map visualization that renders the multi-agent orchestration system as an interactive 3D-style map. The visualization will replace or augment the existing table-based dashboard with a spatial representation that shows:

- Geographic relationships between components
- Real-time agent activity and status
- Work flow through the system (convoys)
- Communication patterns (messages between agents)
- System health at a glance

## Design Goals

### Primary Goals
1. **Spatial Cognition**: Make system architecture immediately understandable through spatial metaphor
2. **Real-Time Visibility**: Show live agent status, convoy progress, and activity
3. **Interactivity**: Enable exploration through clicks, hovers, and drill-downs
4. **Performance**: Support 60fps rendering with dozens of animated sprites
5. **Scalability**: Handle variable numbers of rigs, polecats, and convoys

### Secondary Goals
- Aesthetic consistency with "Mad Max" theming
- Mobile responsiveness (future consideration)
- Accessibility via keyboard navigation
- Export/screenshot capability

## Isometric Projection System

### Coordinate System

**Isometric Projection Formula:**
```
screen_x = (grid_x - grid_y) * TILE_WIDTH / 2
screen_y = (grid_x + grid_y) * TILE_HEIGHT / 2
```

**Constants:**
- `TILE_WIDTH = 128` pixels (width of isometric tile base)
- `TILE_HEIGHT = 64` pixels (height of isometric tile base)
- Projection angle: 30° (2:1 ratio for isometric dimetric)
- Z-axis factor: 0.5 (for building height rendering)

### Grid Layout

**World Grid:** 21x21 tiles (center at 10,10)
- Total canvas size: 1600x1200 pixels (with margins)
- Viewport: Scrollable/pannable for larger screens
- Center tile (10,10): Mayor's Office

### Rendering Order (Z-Index)

Sprites render back-to-front based on:
```
sort_key = grid_y * 1000 + grid_x * 100 + z_height
```

**Layer order (bottom to top):**
1. Ground tiles (roads, plazas)
2. Building bases
3. Building walls
4. Building roofs
5. Polecats (ground level)
6. Convoys (on roads)
7. Message particles (flying)
8. UI overlays (tooltips, modals)

## Spatial Layout

### Central Plaza (Mayor District)

**Grid Position:** (10, 10)
**Components:**
- **Mayor's Office**: Large hexagonal building (3x3 tiles)
  - Grid bounds: (9,9) to (11,11)
  - Building height: 180px
  - Visual style: Gold/yellow (#facc15) with hexagonal roof
  - Windows showing activity status
  - Flag on rooftop

**Plaza Features:**
- Circular plaza with roads extending in 4 cardinal directions
- Decorative elements: Lampposts, benches (optional)
- Central fountain or statue (optional)

### Rig Districts (Compass Layout)

Each rig occupies a district with 3 buildings arranged in an L-shape:

#### North District (Example: "gastown" rig)
**Grid Position:** (10, 4) center
- **Witness Tower**: (9, 4) - Triangle roof, green (#4ade80)
- **Deacon Post Office**: (10, 4) - Square building, blue (#60a5fa)
- **Refinery Plant**: (11, 4) - Pentagon roof, red (#f87171)

#### East District (Example: "claycantrell" rig)
**Grid Position:** (16, 10) center
- Buildings at: (16, 9), (16, 10), (16, 11)

#### South District (Example: "longeye" rig)
**Grid Position:** (10, 16) center
- Buildings at: (9, 16), (10, 16), (11, 16)

#### West District (Example: "beads" rig)
**Grid Position:** (4, 10) center
- Buildings at: (4, 9), (4, 10), (4, 11)

**Dynamic Districts:**
- Additional rigs appear in NE, SE, SW, NW diagonal positions
- Grid positions: (15, 5), (15, 15), (5, 15), (5, 5)
- Up to 8 rigs supported before requiring map expansion

### Road Network

**Primary Roads:** 5 tiles wide, connecting plaza to districts
- North road: (10, 5) → (10, 9)
- East road: (11, 10) → (15, 10)
- South road: (10, 11) → (10, 15)
- West road: (5, 10) → (9, 10)

**Visual Style:**
- Desert dirt roads with tire tracks
- Dashed center line
- Dust particles (optional animation)

**Purpose:**
- Convoy movement paths
- Visual connection between components
- Navigation guides for user

## Building Specifications

### Mayor's Office

**Dimensions:**
- Base: 128x128 pixels (hexagonal footprint)
- Height: 180 pixels
- Roof: Hexagonal, flat top with antenna

**Visual Elements:**
- Gold/yellow color (#facc15)
- Large front entrance facing south
- Multiple windows with light indicators
- Rooftop flag or banner
- Shadow cast to southeast

**Status Indicators:**
- Window brightness indicates system health
- Pulsing glow when processing requests
- Color shift to red on critical errors

### Witness Tower

**Dimensions:**
- Base: 64x64 pixels (square footprint)
- Height: 140 pixels
- Roof: Triangular/pyramid, green (#4ade80)

**Visual Elements:**
- Observation deck at mid-height
- Rotating radar dish on roof
- Green status light
- Telescope or binoculars visible

**Status Indicators:**
- Light color: green (healthy), yellow (degraded), red (critical)
- Radar rotation speed indicates monitoring activity
- Polecat count displayed as floating number

### Deacon Post Office

**Dimensions:**
- Base: 64x64 pixels (square footprint)
- Height: 100 pixels
- Roof: Flat with postal flag, blue (#60a5fa)

**Visual Elements:**
- Blue color scheme
- Mail sorting area visible through windows
- Postal flag on roof
- Loading dock on side

**Status Indicators:**
- Envelope icons floating up (outgoing messages)
- Blinking light when messages queued
- Busy animation when delivering

### Refinery Plant

**Dimensions:**
- Base: 80x80 pixels (pentagon footprint)
- Height: 120 pixels
- Roof: Pentagon with smokestack, red (#f87171)

**Visual Elements:**
- Industrial style with pipes
- Smokestack emitting steam/smoke
- Merge conveyor visible
- Red warning lights

**Status Indicators:**
- Smoke color: white (idle), gray (processing), black (error)
- Conveyor animation speed = queue length
- PR icons on conveyor belt

## Sprite Specifications

### Polecat Agents

**Base Sprite:**
- Size: 24x32 pixels (small humanoid figure)
- Idle animation: 4 frames, 500ms cycle
- Walking animation: 8 frames, 200ms per frame, 8 directions

**Visual Style:**
- Worker outfit with hardhat
- Color-coded by status:
  - Active: Purple (#a78bfa) - animated, moving
  - Idle: Gray (#9ca3af) - standing still
  - Busy: Yellow (#facc15) - working at building
  - Stuck: Red (#f87171) - error indicator above head
  - Stale: Orange (#fb923c) - reduced opacity

**Positioning:**
- Clustered near their rig's Witness building
- Moving along roads when assigned to convoy
- Enter/exit buildings through doors

**Interactivity:**
- Hover: Show tooltip with name, rig, current task
- Click: Open modal with full polecat details
- Nameplate appears below sprite on hover

### Convoy War Rigs

**Base Sprite:**
- Size: 64x48 pixels (truck/vehicle)
- Facing: 8 directions (45° increments)
- Animation: Wheels rotating, exhaust smoke

**Visual Style:**
- Mad Max aesthetic: rugged, industrial
- Color: Desert tan with rust (#d4a574)
- Cargo trailer showing bead icons
- Dust trail behind vehicle

**Movement:**
- Path: District → Roads → Plaza → Mayor's Office
- Speed: 100 pixels/second
- Easing: Smooth acceleration/deceleration
- Turn animation: Sprite rotates to face direction

**Status Display:**
- Cargo fill level (progress bar on trailer)
- Bead count badge
- Status color glow:
  - Yellow: In progress
  - Green: Completed
  - Red: Blocked/error

**Interactivity:**
- Hover: Show convoy ID, progress, bead count
- Click: Open convoy detail modal

### Message Particles

**Sprite:**
- Size: 16x16 pixels (envelope icon)
- Animation: Bobbing/floating motion
- Trail: Dotted line path

**Movement:**
- Source: Deacon building
- Path: Arc trajectory (bezier curve)
- Destination: Target building (Witness, Refinery, Mayor)
- Speed: 200 pixels/second
- Lifetime: 2-3 seconds

**Visual Style:**
- White envelope with blue outline
- Sparkle effect on delivery
- Fades in/out at endpoints

## Color Palette

### Building Colors (Mad Max Theme)
- **Mayor**: Gold `#facc15` (authority, central)
- **Witness**: Green `#4ade80` (health, monitoring)
- **Deacon**: Blue `#60a5fa` (communication, trust)
- **Refinery**: Red `#f87171` (industry, processing)

### Status Colors
- **Active/Healthy**: `#4ade80` (green)
- **Warning/Degraded**: `#facc15` (yellow)
- **Error/Critical**: `#f87171` (red)
- **Idle/Unknown**: `#9ca3af` (gray)
- **Busy/Working**: `#a78bfa` (purple)

### Environment Colors
- **Background**: Dark desert `#0a0a14` (night sky)
- **Ground**: Sand `#c19a6b`
- **Roads**: Dirt `#8b7355`
- **Shadows**: `rgba(0, 0, 0, 0.4)`

### Accent Colors
- **Text**: `#eee` (primary), `#aaa` (secondary)
- **Borders**: `#0f3460` (dark blue)
- **Highlights**: `#fff` (white glow)

## Interactive Elements

### Click Handlers

**Buildings:**
- **Target**: Any building sprite
- **Action**: Open modal overlay with:
  - Building type and role description
  - Current status and metrics
  - Associated agents/polecats
  - Recent activity log
  - Quick actions (if applicable)

**Polecats:**
- **Target**: Polecat sprite
- **Action**: Open modal with:
  - Polecat name and ID
  - Current rig and convoy assignment
  - Work status and current task
  - Last activity timestamp
  - Recent work history
  - Debug logs link

**Convoys:**
- **Target**: War rig sprite
- **Action**: Open modal with:
  - Convoy ID and title
  - Progress bar and bead list
  - Assigned polecats
  - Timeline of activities
  - Link to convoy detail page

**Roads/Ground:**
- **Target**: Empty map area
- **Action**: Deselect current selection (close modals)

### Hover Effects

**All Interactive Sprites:**
- Brightness increase (+20%)
- Glow outline (2px, white)
- Cursor change to pointer
- Tooltip appears after 500ms delay

**Tooltip Content:**
- **Buildings**: Name, type, status line
- **Polecats**: Name, current task (truncated)
- **Convoys**: ID, progress percentage

### Camera Controls

**Pan:**
- Click-and-drag on background
- Arrow keys (20px per press)
- Edge scrolling (when mouse near edge)

**Zoom:**
- Mouse wheel (10% per scroll tick)
- Zoom range: 50% to 200%
- Zoom preserves center focus point

**Reset:**
- Double-click background to reset view
- "R" key to reset camera
- "F" key to fit all content

### Keyboard Navigation

**Selection:**
- Tab: Cycle through interactive elements
- Enter: Activate selected element (click)
- Esc: Close modals/deselect

**Camera:**
- Arrow keys: Pan view
- +/-: Zoom in/out
- Home: Reset to default view

## Animation Specifications

### Building Animations

**Idle Animations:**
- **Mayor**: Flag waving (subtle breeze)
- **Witness**: Radar dish rotating (2 RPM)
- **Deacon**: Flag fluttering
- **Refinery**: Smoke rising, conveyor moving

**Activity Animations:**
- **Polecat enters building**: Door opens, sprite shrinks and fades
- **Polecat exits building**: Door opens, sprite grows and fades in
- **Convoy arrives**: Mayor door opens, cargo unloads
- **Message delivery**: Particle sparkle, building flashes

### Movement Animations

**Polecat Walking:**
- 8-frame walking cycle
- Smooth position interpolation (linear easing)
- Direction changes: Sprite rotates smoothly (200ms)

**Convoy Driving:**
- Wheel rotation (proportional to speed)
- Suspension bounce on road bumps
- Exhaust smoke particles trailing
- Turn anticipation (slight lean before direction change)

**Message Flight:**
- Parabolic arc trajectory
- Rotation animation (spinning envelope)
- Sparkle particles along path
- Impact effect on arrival

### Transition Animations

**Building Construction (on rig registration):**
- Fade in from ground (500ms)
- Scale up from 0% to 100%
- Brief shine effect on completion

**Building Demolition (on rig removal):**
- Fade out (300ms)
- Scale down to 50%
- Dust cloud particle effect

**Polecat Spawn:**
- Fade in near Witness building
- Brief glow effect (status color)

**Polecat Despawn:**
- Fade out (200ms)
- Small puff particle

## Real-Time Updates

### WebSocket Integration

**Endpoint:** `ws://[host]/ws/gastown`

**Message Types:**

1. **polecat_status**
   ```json
   {
     "type": "polecat_status",
     "polecat_id": "rictus",
     "rig": "gastown",
     "status": "active",
     "position": {"x": 10, "y": 12},
     "task": "Working on hq-ftc3.3"
   }
   ```
   **Action**: Update polecat sprite position, status color, tooltip

2. **convoy_progress**
   ```json
   {
     "type": "convoy_progress",
     "convoy_id": "convoy-123",
     "progress": 0.75,
     "position": {"x": 12, "y": 10},
     "beads": ["hq-ftc3.3", "hq-ftc3.4"]
   }
   ```
   **Action**: Update convoy sprite position, cargo fill level

3. **message_sent**
   ```json
   {
     "type": "message_sent",
     "from": "deacon_gastown",
     "to": "witness_gastown",
     "message_type": "health_check"
   }
   ```
   **Action**: Spawn message particle, animate from source to destination

4. **refinery_event**
   ```json
   {
     "type": "refinery_event",
     "rig": "gastown",
     "event": "pr_merged",
     "pr_number": 123
   }
   ```
   **Action**: Refinery smoke puff, conveyor animation spike

5. **map_update**
   ```json
   {
     "type": "map_update",
     "rigs": ["gastown", "beads", "claycantrell"],
     "mayor_status": "healthy"
   }
   ```
   **Action**: Full map rebuild (on rig add/remove)

**Fallback:** HTMX polling every 10s if WebSocket unavailable

### Update Strategy

**Optimization:**
- Sprite pooling (reuse particle objects)
- Dirty rectangle rendering (only redraw changed areas)
- Throttle updates to 60fps max
- Batch WebSocket messages (debounce 100ms)

**State Management:**
- Maintain client-side map state object
- Apply WebSocket updates as state patches
- Render from state on every frame
- Persist view settings (zoom, pan) to localStorage

## Technical Implementation

### Technology Stack

**Rendering Engine:** PixiJS v7+
- GPU-accelerated WebGL rendering
- Sprite batching and texture atlases
- Built-in animation support
- Robust event system

**Alternative Consideration:** Phaser 3 (if game features needed)

**Data Fetching:**
- WebSocket for real-time updates
- Initial state from REST endpoint `/api/gastown/state`
- HTMX integration for graceful fallback

### Asset Requirements

**Sprite Sheets:**
- `buildings.png` - All building sprites (1024x1024 atlas)
- `polecats.png` - Polecat animation frames (512x512 atlas)
- `convoys.png` - Convoy sprites, 8 directions (512x512 atlas)
- `particles.png` - Message icons, effects (256x256 atlas)
- `tiles.png` - Ground tiles, roads (512x512 atlas)

**Sprite Sheet Format:**
- PNG with transparency
- Accompanied by JSON atlas file (TexturePacker format)
- Naming convention: `sprite_name_frame_direction.png`

**Fonts:**
- Primary: 'SF Mono', 'Menlo', 'Monaco' (monospace)
- Fallback: System monospace
- Sizes: 12px (tooltips), 14px (labels), 16px (headers)

### Performance Targets

**Requirements:**
- 60fps sustained with 50+ sprites
- <100ms initial load time
- <16ms frame time (60fps budget)
- <50MB memory footprint

**Optimization Techniques:**
- Sprite culling (don't render off-screen)
- Object pooling for particles
- Texture atlasing (reduce draw calls)
- Level-of-detail (simplify distant sprites)
- Debounced WebSocket message processing

### File Structure

```
internal/web/
├── templates/
│   └── convoy.html (embed map canvas)
├── static/
│   ├── js/
│   │   ├── gastown-map.js (main map controller)
│   │   ├── isometric.js (projection utilities)
│   │   ├── sprites.js (sprite classes)
│   │   └── animations.js (animation system)
│   ├── assets/
│   │   ├── sprites/
│   │   │   ├── buildings.png
│   │   │   ├── buildings.json
│   │   │   ├── polecats.png
│   │   │   ├── polecats.json
│   │   │   ├── convoys.png
│   │   │   ├── convoys.json
│   │   │   └── particles.png
│   │   └── sounds/ (optional)
│   │       └── ambient.mp3
│   └── css/
│       └── map.css (modal overlays, UI)
└── websocket.go (existing WebSocket handler)
```

## Data Model

### Map State Object

```typescript
interface GasTownMapState {
  mayor: MayorState;
  rigs: Map<string, RigState>;
  polecats: Map<string, PolecatState>;
  convoys: Map<string, ConvoyState>;
  messages: MessageState[];
}

interface MayorState {
  status: 'healthy' | 'degraded' | 'critical';
  position: GridPosition; // Always (10, 10)
}

interface RigState {
  name: string;
  position: GridPosition;
  buildings: {
    witness: BuildingState;
    deacon: BuildingState;
    refinery: BuildingState;
  };
}

interface BuildingState {
  type: 'mayor' | 'witness' | 'deacon' | 'refinery';
  position: GridPosition;
  status: 'idle' | 'active' | 'busy' | 'error';
  metrics: Record<string, number>; // For tooltips
}

interface PolecatState {
  id: string;
  name: string;
  rig: string;
  position: GridPosition;
  status: 'active' | 'idle' | 'busy' | 'stuck' | 'stale';
  task: string;
  lastActivity: Date;
}

interface ConvoyState {
  id: string;
  title: string;
  progress: number; // 0.0 to 1.0
  position: GridPosition;
  path: GridPosition[]; // Movement waypoints
  beads: string[];
  status: 'in_progress' | 'completed' | 'blocked';
}

interface MessageState {
  id: string;
  from: GridPosition;
  to: GridPosition;
  type: string;
  spawnTime: Date;
}

interface GridPosition {
  x: number; // Grid coordinate (0-20)
  y: number; // Grid coordinate (0-20)
  z: number; // Height offset (for flying sprites)
}
```

### REST API Endpoints

**GET `/api/gastown/state`**
- Returns: Full `GasTownMapState` object
- Used: Initial map load
- Cache: 5 seconds

**GET `/api/rigs`**
- Returns: List of registered rigs
- Used: Build rig district layout

**GET `/api/polecats`**
- Returns: List of active polecats with status
- Used: Spawn polecat sprites

**GET `/api/convoys`**
- Returns: List of active convoys with progress
- Used: Spawn convoy sprites

## Accessibility

### Keyboard Support
- Full keyboard navigation (Tab, Arrow keys, Enter)
- Screen reader labels on all interactive elements
- Focus indicators (outline on selected sprite)

### Alternative Text
- ARIA labels on buildings: "Mayor's Office - Status: Healthy"
- Polecat descriptions: "Polecat 'rictus' - Working on hq-ftc3.3"
- Convoy descriptions: "Convoy convoy-123 - 75% complete"

### Color Blindness
- Status indicators use both color AND shape/icon
- High contrast mode option (future)
- Pattern overlays for red/green distinctions

### Screen Reader Mode
- Fallback to table view if canvas not supported
- Toggle button to switch between map and table view
- Semantic HTML for modal content

## Future Enhancements

### Phase 2 (Post-MVP)
- Sound effects (ambient town noise, convoy engines)
- Day/night cycle (visual theme change)
- Weather effects (sandstorms, heat haze)
- Minimap in corner (overview of full map)
- Click-to-command (assign polecat to convoy from map)

### Phase 3 (Advanced)
- Historical playback (scrub timeline to see past state)
- Heatmap overlays (activity density, error hotspots)
- 3D buildings (full 3D WebGL rendering)
- VR mode (explore Gas Town in virtual reality)
- Multiplayer cursors (see other users' viewports)

## Implementation Phases

### Phase 1: Static Map (hq-ftc3.4)
**Scope:**
- Render isometric grid and buildings
- Static sprite placement based on API data
- Basic interactivity (clicks, hovers)
- Modal overlays for details

**Deliverables:**
- PixiJS integration in convoy.html
- Building sprites rendered
- Click handlers working
- Basic WebSocket connection

**Success Criteria:**
- Map loads in <2 seconds
- All buildings clickable
- Modals show correct data
- No visual glitches

### Phase 2: Animation
**Scope:**
- Polecat movement animations
- Convoy driving along roads
- Message particle effects
- Building idle animations

**Deliverables:**
- Animation system in place
- Movement interpolation smooth
- Particle effects performant

**Success Criteria:**
- Maintains 60fps with 20 polecats
- Animations feel natural
- No jank or stuttering

### Phase 3: Real-Time Updates
**Scope:**
- WebSocket event processing
- Live polecat status updates
- Convoy progress tracking
- Dynamic map changes (rig add/remove)

**Deliverables:**
- Full WebSocket integration
- State management system
- Automatic map updates

**Success Criteria:**
- Updates reflect within 500ms
- No memory leaks over 24hrs
- Graceful WebSocket reconnection

## Testing Strategy

### Visual Regression Tests
- Screenshot comparison of map renders
- Compare against reference images
- Detect unintended visual changes

### Performance Tests
- FPS monitoring under load (100 polecats)
- Memory leak detection (24hr soak test)
- Network bandwidth usage (WebSocket)

### Interaction Tests
- Click handler coverage (all sprites)
- Keyboard navigation flow
- Modal open/close cycles

### Cross-Browser Tests
- Chrome, Firefox, Safari, Edge
- Mobile Safari, Chrome Mobile
- Fallback to table view on unsupported browsers

## Open Questions

1. **Sprite Art Style**: Should we use pixel art or vector illustrations?
   - **Recommendation**: Pixel art for retro Mad Max aesthetic

2. **Mobile Support**: Should MVP include mobile-optimized view?
   - **Recommendation**: Desktop-first, mobile in Phase 2

3. **Sound**: Include ambient sound effects?
   - **Recommendation**: Optional, disabled by default

4. **Performance**: Target lowest supported hardware?
   - **Recommendation**: Modern laptops (2018+), degrade gracefully

5. **Map Size**: Support more than 8 rigs?
   - **Recommendation**: Plan for 16 rigs, expand grid if needed

## References

- [PixiJS Documentation](https://pixijs.com/)
- [Isometric Game Programming](https://en.wikipedia.org/wiki/Isometric_video_game_graphics)
- [WebSocket API MDN](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
- [Gas Town Architecture](./architecture.md)
- [Convoy Lifecycle](./convoy-lifecycle.md)

## Appendix: Coordinate Examples

### Example Grid Positions

**Mayor's Office:**
- Grid: (10, 10)
- Screen: (800, 600) [center of 1600x1200 canvas]

**North District (Gastown):**
- Witness: Grid (9, 4) → Screen (544, 416)
- Deacon: Grid (10, 4) → Screen (608, 448)
- Refinery: Grid (11, 4) → Screen (672, 480)

**Convoy Path (North District → Mayor):**
```
Waypoints:
  Start: (10, 4) → (608, 448)
  Road: (10, 6) → (608, 512)
  Road: (10, 8) → (608, 576)
  Plaza: (10, 9) → (608, 608)
  Mayor: (10, 10) → (608, 640)
```

### Isometric Math Examples

**Convert grid to screen:**
```javascript
function gridToScreen(gridX, gridY) {
  const TILE_WIDTH = 128;
  const TILE_HEIGHT = 64;
  const ORIGIN_X = 800; // Canvas center X
  const ORIGIN_Y = 400; // Canvas center Y

  const screenX = ORIGIN_X + (gridX - gridY) * TILE_WIDTH / 2;
  const screenY = ORIGIN_Y + (gridX + gridY) * TILE_HEIGHT / 2;

  return { x: screenX, y: screenY };
}
```

**Convert screen to grid:**
```javascript
function screenToGrid(screenX, screenY) {
  const TILE_WIDTH = 128;
  const TILE_HEIGHT = 64;
  const ORIGIN_X = 800;
  const ORIGIN_Y = 400;

  const relX = screenX - ORIGIN_X;
  const relY = screenY - ORIGIN_Y;

  const gridX = (relX / (TILE_WIDTH / 2) + relY / (TILE_HEIGHT / 2)) / 2;
  const gridY = (relY / (TILE_HEIGHT / 2) - relX / (TILE_WIDTH / 2)) / 2;

  return { x: Math.round(gridX), y: Math.round(gridY) };
}
```

---

**End of Design Specification**
