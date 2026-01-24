package polecat

import (
	"testing"
)

func TestCalculatePosition(t *testing.T) {
	tests := []struct {
		name     string
		poleName string
		state    State
		rigIndex int
		wantY    float64 // Expected base Y (allowing for scatter)
	}{
		{
			name:     "working state should be in middle zone",
			poleName: "toast",
			state:    StateWorking,
			rigIndex: 0,
			wantY:    300,
		},
		{
			name:     "done state should be in bottom zone",
			poleName: "shadow",
			state:    StateDone,
			rigIndex: 0,
			wantY:    500,
		},
		{
			name:     "stuck state should be in upper zone",
			poleName: "copper",
			state:    StateStuck,
			rigIndex: 0,
			wantY:    150,
		},
		{
			name:     "active state (legacy) should be in middle zone",
			poleName: "ash",
			state:    StateActive,
			rigIndex: 0,
			wantY:    300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := CalculatePosition(tt.poleName, tt.state, tt.rigIndex)

			// Check Y position is within scatter range of expected base
			scatterRange := 20.0
			if pos.Y < tt.wantY-scatterRange || pos.Y > tt.wantY+scatterRange {
				t.Errorf("Y position %f is outside expected range [%f, %f]",
					pos.Y, tt.wantY-scatterRange, tt.wantY+scatterRange)
			}
		})
	}
}

func TestCalculatePosition_Deterministic(t *testing.T) {
	// Same name and state should always produce same position
	pos1 := CalculatePosition("toast", StateWorking, 0)
	pos2 := CalculatePosition("toast", StateWorking, 0)

	if pos1.X != pos2.X || pos1.Y != pos2.Y {
		t.Errorf("Position should be deterministic: got %v and %v", pos1, pos2)
	}

	// Different names should produce different positions (very likely)
	pos3 := CalculatePosition("shadow", StateWorking, 0)
	if pos1.X == pos3.X && pos1.Y == pos3.Y {
		t.Errorf("Different names should produce different positions (hash collision unlikely)")
	}
}

func TestCalculatePosition_RigIndex(t *testing.T) {
	// Different rig indices should produce different X positions
	pos0 := CalculatePosition("toast", StateWorking, 0)
	pos1 := CalculatePosition("toast", StateWorking, 1)
	pos2 := CalculatePosition("toast", StateWorking, 2)

	// Base X should be 100, 300, 500 respectively (with scatter)
	// Difference should be ~200px
	diff01 := pos1.X - pos0.X
	diff12 := pos2.X - pos1.X

	// Allow for some variance due to scatter, but should be close to 200
	if diff01 < 180 || diff01 > 220 {
		t.Errorf("X position difference between rig 0 and 1 should be ~200, got %f", diff01)
	}
	if diff12 < 180 || diff12 > 220 {
		t.Errorf("X position difference between rig 1 and 2 should be ~200, got %f", diff12)
	}
}

func TestBuildRigIndex(t *testing.T) {
	tests := []struct {
		name string
		rigs []string
		want map[string]int
	}{
		{
			name: "single rig",
			rigs: []string{"gastown"},
			want: map[string]int{"gastown": 0},
		},
		{
			name: "multiple rigs sorted alphabetically",
			rigs: []string{"roxas", "gastown", "beads"},
			want: map[string]int{
				"beads":    0,
				"gastown":  1,
				"roxas":    2,
			},
		},
		{
			name: "already sorted",
			rigs: []string{"alpha", "beta", "gamma"},
			want: map[string]int{
				"alpha": 0,
				"beta":  1,
				"gamma": 2,
			},
		},
		{
			name: "empty input",
			rigs: []string{},
			want: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildRigIndex(tt.rigs)

			if len(got) != len(tt.want) {
				t.Errorf("BuildRigIndex() returned %d entries, want %d", len(got), len(tt.want))
			}

			for rig, wantIdx := range tt.want {
				gotIdx, ok := got[rig]
				if !ok {
					t.Errorf("BuildRigIndex() missing rig %q", rig)
					continue
				}
				if gotIdx != wantIdx {
					t.Errorf("BuildRigIndex()[%q] = %d, want %d", rig, gotIdx, wantIdx)
				}
			}
		})
	}
}

func TestBuildRigIndex_StableSort(t *testing.T) {
	// Same input should always produce same output
	rigs := []string{"zulu", "alpha", "mike", "bravo"}

	idx1 := BuildRigIndex(rigs)
	idx2 := BuildRigIndex(rigs)

	if len(idx1) != len(idx2) {
		t.Fatal("Index sizes differ")
	}

	for rig, i1 := range idx1 {
		i2, ok := idx2[rig]
		if !ok {
			t.Errorf("Rig %q missing in second index", rig)
		}
		if i1 != i2 {
			t.Errorf("Index for %q changed: %d vs %d", rig, i1, i2)
		}
	}
}
