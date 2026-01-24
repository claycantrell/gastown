# Gas Town Building Sprite Specifications

This document defines the visual specifications for rendering Gas Town buildings using PixiJS.

## Building Types

Gas Town consists of five primary building types, each representing a different agent role:

1. **Mayor** - Town coordinator and global overseer
2. **Witness** - Health monitor and worker supervisor
3. **Deacon** - Configuration and state manager
4. **Refinery** - Merge queue processor
5. **Polecat** - Worker agents (multiple instances)

## Visual Design

### Building Dimensions

All buildings use simple geometric shapes for initial implementation:

| Building | Shape | Width | Height | Color |
|----------|-------|-------|--------|-------|
| Mayor | Hexagon | 80px | 80px | #facc15 (yellow) |
| Witness | Triangle | 70px | 70px | #4ade80 (green) |
| Deacon | Square | 60px | 60px | #60a5fa (blue) |
| Refinery | Pentagon | 75px | 75px | #f87171 (red) |
| Polecat | Circle | 50px | 50px | #a78bfa (purple) |

### Layout Positioning

Buildings are arranged in a Gas Town layout:

```
        [Mayor]
        (400, 100)
           |
    +------+------+
    |             |
[Witness]    [Deacon]
(200, 250)  (600, 250)

   [Refinery]
   (400, 400)

[Polecat] [Polecat] [Polecat]
(150,550) (400,550) (650,550)
```

### Rendering Properties

- **Stroke**: 3px solid white (#ffffff)
- **Fill**: Building-specific color (see table above)
- **Opacity**: 0.9
- **Shadow**: None (static rendering, no effects)

### Labels

Each building should have a text label below it:
- **Font**: 'Arial', sans-serif
- **Size**: 14px
- **Color**: #eee (light gray)
- **Position**: Centered below building, 10px gap

## Canvas Specifications

- **Canvas Size**: 800px × 650px
- **Background**: #1a1a2e (dark blue-gray, matching convoy dashboard)
- **Border**: 1px solid #0f3460

## PixiJS Implementation Notes

1. Use `PIXI.Graphics` for drawing shapes
2. Use `PIXI.Text` for labels
3. Create a container for each building (shape + label)
4. Static rendering - no animation or interaction in initial version
5. Render at application startup, no continuous render loop needed

## Future Enhancements (Not in Scope)

- Animation (pulsing, activity indicators)
- Connection lines between buildings
- Status indicators (health, activity)
- Interactive tooltips
- Sprite assets (replace geometric shapes)
