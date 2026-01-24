package polecat

import (
	"hash/fnv"
)

// Position represents a 2D coordinate on the desert map.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// CalculatePosition determines the 2D map position for a polecat based on its
// name, rig, and state. The desert map uses:
//
//	X-axis: Rig distribution (horizontal zones, 200px apart)
//	Y-axis: Work state progression (vertical layers)
//	  - Working: Y=300 (middle)
//	  - Done:    Y=500 (bottom)
//	  - Stuck:   Y=150 (top, between ready and working)
//	  - Other:   Y=100 (top, ready/idle zone)
//
// Each polecat gets a deterministic scatter based on its name (±30px horizontally,
// ±20px vertically) to avoid exact overlap while keeping positions stable.
func CalculatePosition(name string, state State, rigIndex int) Position {
	// Base X: Rig spacing (200px apart) starting at x=100
	baseX := float64(rigIndex*200 + 100)

	// Deterministic scatter based on polecat name
	// Use hash to generate consistent pseudo-random values
	h := fnv.New32a()
	h.Write([]byte(name))
	hash := h.Sum32()

	// Extract two pseudo-random values from the hash
	xSeed := float64(hash&0xFFFF) / 65535.0       // First 16 bits
	ySeed := float64((hash>>16)&0xFFFF) / 65535.0 // Next 16 bits

	// Horizontal scatter: ±30px
	scatterX := (xSeed * 60) - 30

	// Base Y: State-based vertical position
	var baseY float64
	switch state {
	case StateWorking, StateActive:
		baseY = 300 // Middle zone: actively working
	case StateDone:
		baseY = 500 // Bottom zone: completed work
	case StateStuck:
		baseY = 150 // Upper-middle zone: stuck/needs help
	default:
		baseY = 100 // Top zone: ready/waiting
	}

	// Vertical scatter: ±20px
	scatterY := (ySeed * 40) - 20

	return Position{
		X: baseX + scatterX,
		Y: baseY + scatterY,
	}
}

// BuildRigIndex creates a mapping from rig name to rig index for position calculation.
// Rigs are sorted alphabetically and assigned sequential indices.
func BuildRigIndex(rigs []string) map[string]int {
	// Sort rigs alphabetically for consistent positioning
	sortedRigs := make([]string, len(rigs))
	copy(sortedRigs, rigs)
	// Simple insertion sort (adequate for small rig counts)
	for i := 1; i < len(sortedRigs); i++ {
		key := sortedRigs[i]
		j := i - 1
		for j >= 0 && sortedRigs[j] > key {
			sortedRigs[j+1] = sortedRigs[j]
			j--
		}
		sortedRigs[j+1] = key
	}

	// Build index map
	index := make(map[string]int, len(sortedRigs))
	for i, rig := range sortedRigs {
		index[rig] = i
	}
	return index
}
