# Gas Town Map Design Guide

This guide explains the Gas Town isometric map and how to use it for understanding the system architecture.

## Available Maps

### 1. ASCII Isometric Map
**File**: `gas-town-map.md`
**Use case**: Terminal viewing, quick reference, embedded in CLI tools

The ASCII map provides a text-based isometric view that works in any terminal or text editor. Perfect for:
- Quick architectural reference
- Documentation in markdown files
- Terminal-based presentations
- Embedded in CLI help text

### 2. SVG Isometric Map
**File**: `gas-town-map.svg`
**Use case**: Documentation, presentations, web viewing

The SVG map is a scalable vector graphic with full color and detail. Best for:
- Online documentation
- README files (GitHub, GitLab)
- Presentations and slides
- Print materials
- Interactive web documentation

## Map Components

### Spatial Layout

The isometric map organizes Gas Town into distinct zones:

```
                    Town Center (Coordination)
                           |
        +-----------------+------------------+
        |                                    |
   Rig District A                      Rig District B
   (Project 1)                         (Project 2)
```

### Building Types and Hierarchy

#### Town Level (Gold/Red - Central Authority)
Located in the town center with highest visibility:

- **Mayor Building** (🎩 Gold): Chief coordinator
  - Initiates convoys
  - Distributes work across rigs
  - Monitors global progress

- **Deacon Building** (🔔 Red): System watchdog
  - Runs patrol cycles
  - Monitors agent health
  - Triggers recovery actions

#### Rig Level (Green/Pink - District Management)
Each rig has its own management layer:

- **Refinery** (🏭 Green): Merge control center
  - Manages merge queue
  - Integrates polecat work
  - Ensures code quality

- **Witness** (👁️ Pink): Oversight agent
  - Monitors polecats
  - Detects stuck agents
  - Reports to Mayor

#### Worker Level (Orange/Brown - Execution)
Where actual work happens:

- **Polecats** (🦨 Orange): Ephemeral workers
  - Spawned for specific tasks
  - Work in isolated branches
  - Auto-cleanup on completion

- **Crew** (👤 Brown): Persistent workers
  - Long-lived collaboration
  - Direct main branch access
  - Manual lifecycle management

## Work Flow Patterns

### Convoy Route (Blue Solid Lines)
Shows how work travels from Mayor to rigs:
```
Mayor → Convoy Creation → Distribution to Rig → Assignment to Worker
```

### Mail Flow (Green Dashed Lines)
Asynchronous messaging between agents:
```
Agent A → Mail System → Agent B (reads when ready)
```

### Work Paths (Gray Dashed Lines)
Internal rig communication:
```
Witness → Polecat Assignment → Execution → Refinery Integration
```

## Reading the Map

### Vertical Axis (Authority)
- **Top**: Town-level coordination (Mayor, Deacon)
- **Middle**: Rig-level management (Refinery, Witness)
- **Bottom**: Worker execution (Polecats, Crew)

### Horizontal Axis (Separation)
- **Center**: Shared town resources
- **Left/Right**: Independent rig districts
- Each rig operates autonomously

### Depth (Worker Layers)
- **Front**: Ephemeral workers (Polecats)
- **Back**: Persistent workers (Crew)

## Scaling the Map

The map can be mentally scaled to match your deployment:

### Small Deployment (1-2 rigs)
```
Mayor + Deacon
    |
   Rig A (2-3 polecats, 1 crew)
```

### Medium Deployment (3-5 rigs)
```
Mayor + Deacon
    |
    +-- Rig A (5 polecats, 2 crew)
    +-- Rig B (3 polecats, 1 crew)
    +-- Rig C (8 polecats, 3 crew)
```

### Large Deployment (10+ rigs)
```
Mayor + Deacon
    |
    +-- Multiple rig districts
    +-- 20+ active polecats per rig
    +-- 5-10 crew members per rig
```

## Common Scenarios Visualized

### Scenario 1: Feature Development
```
1. User tells Mayor: "Build auth system"
2. Mayor (center) creates convoy
3. Convoy routes to appropriate rig (left/right)
4. Witness assigns to polecat
5. Polecat executes in worktree
6. Refinery merges completed work
7. Mayor notified → User updated
```

**Map path**: Mayor → Convoy Route → Rig District → Polecat → Refinery → Main Branch

### Scenario 2: Bug Fix
```
1. Issue created in rig
2. Witness detects ready work
3. Spawns polecat
4. Polecat fixes bug
5. Refinery validates and merges
```

**Map path**: Rig District → Witness → Polecat → Refinery

### Scenario 3: Cross-Rig Work
```
1. Crew member in Rig A needs to work on Rig B
2. Creates worktree in Rig B
3. Work attributed to Rig A crew member
4. Commits appear in Rig B with Rig A attribution
```

**Map path**: Rig A Crew → (worktree) → Rig B District

## Map Symbols Reference

| Symbol | Meaning | Role |
|--------|---------|------|
| 🎩 | Mayor | Town coordinator |
| 🔔 | Deacon | System watchdog |
| 🏭 | Refinery | Merge queue manager |
| 👁️ | Witness | Oversight agent |
| 🦨 | Polecat | Ephemeral worker |
| 👤 | Crew | Persistent worker |
| 🚚 | Convoy | Work distribution |
| ✉️ | Mail | Async messaging |
| 📡 | Nudge | Real-time message |

## Using the Map in Documentation

### Embedding ASCII Map
```markdown
See the architecture map in [docs/design/gas-town-map.md](docs/design/gas-town-map.md)
```

### Embedding SVG Map
```markdown
![Gas Town Architecture](docs/design/gas-town-map.svg)
```

### In HTML
```html
<object type="image/svg+xml" data="docs/design/gas-town-map.svg">
  Gas Town Architecture Map
</object>
```

## Map Updates

When updating the map to reflect architectural changes:

1. **Modify ASCII map** - Edit `gas-town-map.md`
2. **Update SVG map** - Edit `gas-town-map.svg`
3. **Update this guide** - Document new components
4. **Update architecture docs** - Sync with `docs/design/architecture.md`

## Accessibility

### For Terminal Users
The ASCII map works in all terminals and screen readers. Each building is represented with box-drawing characters and emoji labels.

### For Visual Users
The SVG map uses distinct colors and shapes:
- Gold: Town authority (Mayor)
- Red: System monitoring (Deacon)
- Green: Integration (Refinery)
- Pink: Oversight (Witness)
- Orange: Ephemeral work (Polecats)
- Brown: Persistent work (Crew)

### For Colorblind Users
Buildings are differentiated by:
- Shape variations
- Text labels
- Emoji icons
- Position in hierarchy

## Further Reading

- [Architecture Overview](architecture.md) - Detailed system architecture
- [Polecat Lifecycle](../concepts/polecat-lifecycle.md) - How workers are managed
- [Convoy System](../concepts/convoy.md) - Work tracking details
- [Glossary](../glossary.md) - Complete terminology reference

## Contributing

To propose map improvements:
1. Create an issue describing the architectural change
2. Sketch the proposed map update
3. Submit PR with updated maps and documentation
4. Ensure both ASCII and SVG maps stay in sync
