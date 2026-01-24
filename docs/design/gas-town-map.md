# Gas Town Isometric Map

This document contains the visual layout of Gas Town's architecture in isometric view.

## ASCII Isometric Map

```
                                 ╔══════════════════╗
                                 ║   TOWN SQUARE    ║
                                 ║                  ║
                          ┌──────╢    🎩 MAYOR     ║
                          │      ║                  ║
                          │      ╚════╦═════════╦══╝
                          │           ║         ║
                          │      ╔════╩═════════╩══╗
                          │      ║   🔔 DEACON      ║
                  ╔═══════▼══╗   ║   (Watchdog)     ║
                  ║          ║   ╚═══════╦══════════╝
                  ║  Rig A   ║           ║
                  ║ District ║◄──────────╬────────────────┐
                  ║          ║           ║                │
                  ╚═══╦══════╝           ║         ╔══════▼═════╗
                      ║                  ║         ║            ║
         ╔════════════╩═══╗              ║         ║   Rig B    ║
         ║                ║              ║         ║  District  ║
         ║  🏭 REFINERY   ║              ║         ║            ║
         ║  (Merge Queue) ║              ║         ╚══════╦═════╝
         ╚════════════════╝              ║                ║
                                         ║    ╔═══════════╩════╗
         ╔════════════════╗              ║    ║                ║
         ║                ║              ║    ║  🏭 REFINERY   ║
         ║  👁️ WITNESS    ║              ║    ║                ║
         ║  (Oversight)   ║              ║    ╚════════════════╝
         ║                ║              ║
         ╚════════╦═══════╝              ║    ╔════════════════╗
                  ║                      ║    ║                ║
      ┌───────────╬──────────┐          ║    ║  👁️ WITNESS    ║
      │           ║          │          ║    ║                ║
  ╔═══▼═══╗   ╔═══▼═══╗  ╔══▼══╗       ║    ╚════════╦═══════╝
  ║       ║   ║       ║  ║     ║       ║             ║
  ║ 🦨 P1 ║   ║ 🦨 P2 ║  ║ P3  ║       ║    ┌────────╬─────┐
  ║Polecat║   ║Polecat║  ║     ║       ║    │        ║     │
  ╚═══════╝   ╚═══════╝  ╚═════╝       ║ ╔══▼══╗ ╔═══▼═══╗ │
                                        ║ ║     ║ ║       ║ │
  ╔═══════╗   ╔═══════╗                ║ ║ P4  ║ ║ 🦨 P5 ║ │
  ║       ║   ║       ║                ║ ║     ║ ║Polecat║ │
  ║ 👤 C1 ║   ║ 👤 C2 ║                ║ ╚═════╝ ╚═══════╝ │
  ║ Crew  ║   ║ Crew  ║                ║                    │
  ╚═══════╝   ╚═══════╝                ║                    │
                                        ║    ╔═══════╗       │
                                        ║    ║       ║       │
         🚚═══════════════════════════►║    ║ 👤 C3 ║◄──────┘
         Convoy Route                   ║    ║ Crew  ║
                                        ║    ╚═══════╝
         ✉️ ═══════════════════════════►║
         Mail Flow                      ║
                                        ║
         📡═══════════════════════════►║
         Nudge (Real-time)              ║
```

## Map Legend

### Town Level Components (Center)
- **🎩 Mayor HQ**: Chief coordinator, initiates convoys, distributes work
- **🔔 Deacon HQ**: Daemon watchdog, runs patrol cycles, monitors health

### Rig District Components
Each Rig (project) contains:
- **🏭 Refinery**: Manages merge queue, integrates polecat work
- **👁️ Witness**: Patrol agent, oversees polecats and refinery
- **🦨 Polecats**: Ephemeral workers (P1, P2, P3...), spawned for tasks
- **👤 Crew**: Persistent workers (C1, C2, C3...), long-lived collaboration

### Communication Paths
- **🚚 Convoy Routes**: Work distribution from Mayor to rigs
- **✉️ Mail Flow**: Asynchronous messaging between agents
- **📡 Nudges**: Real-time direct messaging
- **Hook Chains**: Work assignment to individual agents (shown as arrows)

## Spatial Organization

### Town Square (Central Hub)
The Mayor and Deacon occupy the town center, with visibility across all rigs.

### Rig Districts (Peripheral)
Each rig forms its own district with:
1. **Refinery** at the district entrance (merge control)
2. **Witness** as district overseer
3. **Polecat workspaces** clustered together (ephemeral areas)
4. **Crew spaces** as permanent structures

### Work Flow Patterns

```
Convoy Creation → Distribution → Assignment → Execution → Integration
    Mayor      →  Rig District → Polecat    →  Work    →  Refinery
                                                           ↓
                                                    Main Branch
```

## Isometric Perspective Notes

The map uses isometric projection where:
- **Vertical axis**: Represents authority/oversight hierarchy
- **Horizontal spread**: Represents different rigs/projects
- **Depth**: Represents worker layers (polecats vs crew)

Buildings closer to the town center have broader visibility and coordination responsibility.

## Example: Work Journey

1. **User** tells Mayor: "Build feature X"
2. **Mayor** creates convoy in town square
3. **Convoy** routes to appropriate Rig District
4. **Witness** in rig receives work
5. **Polecat** spawned and assigned via hook
6. **Polecat** executes work in isolated worktree
7. **Refinery** merges completed work
8. **Mayor** notified of completion

## Scale Reference

- **Town**: 1 per workspace (~/ gt/)
- **Rigs**: Multiple per town (one per project)
- **Witnesses**: 1 per rig
- **Refineries**: 1 per rig
- **Polecats**: Multiple per rig (ephemeral, 0-20+)
- **Crew**: Multiple per rig (persistent, 1-5 typical)
