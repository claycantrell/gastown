package web

import (
	"math"
	"sort"
)

// MapData represents data for the 2D map view.
type MapData struct {
	Buildings []BuildingPosition
	Polecats  []PolecatPosition
}

// BuildingPosition represents a building (convoy) on the map.
type BuildingPosition struct {
	ID         string  // Convoy ID
	Title      string
	Status     string  // "open" or "closed"
	WorkStatus string  // "complete", "active", etc.
	X          float64 // Map coordinates (0-100%)
	Y          float64
	Progress   int // % completion
}

// PolecatPosition represents a polecat worker on the map.
type PolecatPosition struct {
	Name      string  // e.g., "nux", "dag"
	Rig       string
	SessionID string
	X         float64
	Y         float64
	Status    string // Activity status
}

// LayoutMap calculates positions for convoys and polecats on a 2D map.
// Uses a simple grid layout with grouping by status.
func LayoutMap(convoys []ConvoyRow, polecats []PolecatRow) MapData {
	buildings := layoutBuildings(convoys)
	polecatPositions := layoutPolecats(polecats)

	return MapData{
		Buildings: buildings,
		Polecats:  polecatPositions,
	}
}

// layoutBuildings positions buildings in a grid, grouped by status.
// Open convoys on left, closed on right, sorted by last activity.
func layoutBuildings(convoys []ConvoyRow) []BuildingPosition {
	if len(convoys) == 0 {
		return nil
	}

	// Separate open and closed convoys
	var openConvoys, closedConvoys []ConvoyRow
	for _, c := range convoys {
		if c.Status == "closed" {
			closedConvoys = append(closedConvoys, c)
		} else {
			openConvoys = append(openConvoys, c)
		}
	}

	// Sort by last activity (most recent first)
	sortByActivity := func(rows []ConvoyRow) {
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].LastActivity.LastActivity.After(rows[j].LastActivity.LastActivity)
		})
	}
	sortByActivity(openConvoys)
	sortByActivity(closedConvoys)

	var buildings []BuildingPosition

	// Layout open convoys (left side: 10-40%)
	const openX = 25.0
	const closedX = 75.0
	const startY = 15.0
	const endY = 85.0

	buildings = append(buildings, positionColumn(openConvoys, openX, startY, endY)...)
	buildings = append(buildings, positionColumn(closedConvoys, closedX, startY, endY)...)

	return buildings
}

// positionColumn arranges convoys vertically in a column.
func positionColumn(convoys []ConvoyRow, x, startY, endY float64) []BuildingPosition {
	if len(convoys) == 0 {
		return nil
	}

	positions := make([]BuildingPosition, len(convoys))
	spacing := (endY - startY) / float64(len(convoys))

	for i, convoy := range convoys {
		y := startY + float64(i)*spacing
		positions[i] = BuildingPosition{
			ID:         convoy.ID,
			Title:      convoy.Title,
			Status:     convoy.Status,
			WorkStatus: convoy.WorkStatus,
			X:          x,
			Y:          y,
			Progress:   (convoy.Completed * 100) / max(convoy.Total, 1),
		}
	}

	return positions
}

// layoutPolecats positions polecats in a row at the bottom.
// Distributes them evenly across the width.
func layoutPolecats(polecats []PolecatRow) []PolecatPosition {
	if len(polecats) == 0 {
		return nil
	}

	const yPosition = 95.0 // Bottom of map
	const startX = 15.0
	const endX = 85.0

	positions := make([]PolecatPosition, len(polecats))

	if len(polecats) == 1 {
		positions[0] = PolecatPosition{
			Name:      polecats[0].Name,
			Rig:       polecats[0].Rig,
			SessionID: polecats[0].SessionID,
			X:         50.0, // Center single polecat
			Y:         yPosition,
			Status:    string(polecats[0].LastActivity.ColorClass),
		}
		return positions
	}

	spacing := (endX - startX) / float64(len(polecats)-1)

	for i, polecat := range polecats {
		x := startX + float64(i)*spacing
		positions[i] = PolecatPosition{
			Name:      polecat.Name,
			Rig:       polecat.Rig,
			SessionID: polecat.SessionID,
			X:         x,
			Y:         yPosition,
			Status:    string(polecat.LastActivity.ColorClass),
		}
	}

	return positions
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// distance calculates Euclidean distance between two points.
func distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
