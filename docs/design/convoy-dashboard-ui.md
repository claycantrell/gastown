# Gas Town Convoy Dashboard UI Design

**Mad Max Retro City Builder Aesthetic**

## Color Palette

### Primary Colors
```
Rust:    #8B4513  (rust brown - primary UI elements)
Fire:    #FF4500  (orange-red - alerts, active states)
Oil:     #1C1C1C  (deep black - backgrounds)
Metal:   #A9A9A9  (dark gray - borders, inactive elements)
```

### Accent Colors
```
Warning:  #FFD700  (gold - stranded convoys)
Success:  #228B22  (forest green - completed)
Dust:     #D2B48C  (tan - text on dark backgrounds)
Smoke:    #696969  (dim gray - secondary text)
```

## Typography

### Retro Fonts
- **Headers**: Monospace bold (Courier New, Monaco) - retro terminal feel
- **Body**: Sans-serif (Arial, Helvetica) - readability
- **Status**: Monospace (Console font) - technical data display

### Text Style
```css
.convoy-title {
  font-family: 'Courier New', monospace;
  font-weight: bold;
  color: #D2B48C; /* Dust */
  text-transform: uppercase;
  letter-spacing: 2px;
}

.status-text {
  font-family: 'Monaco', monospace;
  color: #A9A9A9; /* Metal */
}
```

## Layout: Isometric City Builder

### Dashboard Grid
```
┌──────────────────────────────────────────────┐
│  GAS TOWN CONVOY OPERATIONS                  │
├────────────┬─────────────────────────────────┤
│            │    Isometric Rig View           │
│  Convoy    │         ⬢                       │
│  List      │      ⬢  ⬢  ⬢                    │
│            │    ⬢  ⬢  ⬢  ⬢                   │
│  ● Active  │      🚚 🚚 🚚                    │
│  ○ Landed  │                                 │
│  ⚠ Strand  │   [Mayor] [Refinery] [Witness]  │
└────────────┴─────────────────────────────────┘
```

### Isometric Tile System
- 32x32 pixel base tiles
- 45° diamond projection
- Staggered row layout for depth
- War rigs move along convoy routes

## Animated Elements

### War Rigs (Convoys)
```
States:
  - IDLE:     Stationary, engine rumble (subtle bounce)
  - ACTIVE:   Moving forward, smoke trails
  - LANDED:   Parked, dust settling animation
  - STRANDED: Broken down, warning flash
```

Animation specs:
- Frame rate: 8-12 fps (retro choppiness)
- Movement: Linear with slight shake
- Smoke particles: 3-5 frames loop

### Sprite Designs

#### Mayor (Coordinator)
```
█████████
█ ◉   ◉ █  (Top hat silhouette)
█   ▼   █
█████████
   ███
```
- Color: Gold (#FFD700) highlights
- Animation: Slight head tilt when active

#### Witness (Monitor)
```
  █████
 █ ● ● █  (Watchful eyes)
 █  ▬  █
  █████
    █
```
- Color: Metal (#A9A9A9) with Fire (#FF4500) eye glow
- Animation: Eyes scan left-right

#### Refinery (Merge Processor)
```
███████████
█  ⚙  ⚙  █  (Industrial machinery)
█ ═══════ █
███████████
```
- Color: Rust (#8B4513) with rotating gears
- Animation: Gears spin when processing

#### Polecat (Worker)
```
  ███
 █ ◉ █  (Simple worker avatar)
  ███
  █ █
```
- Color: Dust (#D2B48C)
- Animation: Walk cycle when assigned, idle bounce when waiting
- Status indicator: Green dot (working), Red dot (stuck)

#### Deacon (Patrol)
```
  █████
 █ ▲ ▲ █  (Alert watcher)
 █  ═  █
  █████
```
- Color: Metal with Warning accents
- Animation: Patrol path movement

### War Rig Types
```
Standard Convoy:  🚚═══  (Simple cargo)
Priority:         🚚══★  (Star marker)
Stranded:         🚚⚠⚠   (Warning symbols)
Landed:           🚚✓    (Checkmark)
```

## Dashboard Components

### Convoy Card (Left Panel)
```
┌────────────────────────┐
│ 🚚 hq-cv-abc           │
│ Feature X       [●]    │
│ ██████░░░░ 60%         │
│ 3/5 issues             │
│                        │
│ Workers: nux, furiosa  │
└────────────────────────┘
```

### Status Indicators
```
● ACTIVE   (Fire orange)
✓ LANDED   (Success green)
⚠ STRANDED (Warning gold)
○ IDLE     (Metal gray)
```

### Isometric View (Right Panel)
- Polecats move to/from Refinery
- Witness patrols perimeter
- Mayor central position
- Convoy progress bars hover above rigs

## Asset Requirements

### Sprites Needed (16x16 or 32x32)
1. Mayor avatar (gold hat)
2. Witness avatar (scanning eyes)
3. Refinery building (industrial)
4. Polecat worker (x4 variations: idle, walk, work, stuck)
5. Deacon patrol (walking animation)
6. War rig convoy (4 states: idle, active, landed, stranded)

### Backgrounds
1. Oil-stained concrete texture (tile)
2. Rust metal panels (UI frames)
3. Fire glow gradient (alerts)
4. Smoke particle sprites (3-5 frames)

### UI Elements
1. Retro terminal border frames
2. Progress bars (rust/fire gradient)
3. Status icons (●○✓⚠)
4. Isometric ground tiles

## Implementation Stack

### Recommended
- **PixiJS** for sprite rendering and animation
- **CSS Grid** for dashboard layout
- **Web Components** for convoy cards
- **Canvas API** for isometric projection

### Alternative (Lightweight)
- **Pure CSS** for retro terminal UI
- **SVG sprites** for scalable icons
- **CSS animations** for simple effects

## Animation Timing

```
Sprite idle:      2s loop
War rig move:     500ms per tile
Smoke particles:  1.5s fade out
Status flash:     750ms pulse
Progress update:  300ms smooth
```

## Accessibility

- High contrast mode: Increase Fire/Dust contrast
- Screen reader labels for all status indicators
- Keyboard navigation through convoy list
- Reduced motion option (disable animations, keep layout)

## File Structure

```
gastown/
├── web/
│   ├── convoy-dashboard.html
│   ├── css/
│   │   ├── mad-max-theme.css
│   │   └── isometric-grid.css
│   ├── js/
│   │   ├── convoy-dashboard.js
│   │   └── sprite-animator.js
│   └── assets/
│       ├── sprites/
│       │   ├── mayor.png
│       │   ├── witness.png
│       │   ├── refinery.png
│       │   ├── polecat-*.png
│       │   └── war-rig-*.png
│       ├── textures/
│       │   ├── rust-panel.png
│       │   └── oil-concrete.png
│       └── fonts/
│           └── retro-terminal.ttf
```

## Quick Win: Terminal Dashboard (No Graphics)

### ASCII Art Version
```
═══════════════════════════════════════════════════
  GAS TOWN CONVOY OPERATIONS
═══════════════════════════════════════════════════

ACTIVE CONVOYS:

  🚚 hq-cv-abc  Feature X                    [████░] 80%
     Workers: nux, furiosa
     Issues:  4/5 closed

  🚚 hq-cv-def  Bug fixes                   [██░░░] 40%
     Workers: toast
     Issues:  2/5 closed

STRANDED:
  ⚠  hq-cv-xyz  Deploy work                [█░░░░] 20%
     No workers assigned!

═══════════════════════════════════════════════════
```

Could start with terminal version, upgrade to graphical later.
