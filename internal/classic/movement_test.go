package classic

import "testing"

func TestBarrelsJumpReferenceVector(t *testing.T) {
	for _, direction := range []int{-1, 0, 1} {
		m := NewMovement()
		startX, startY := m.X, m.Y
		checks := map[int][2]int{1: {0, -1}, 10: {5, -10}, 22: {11, -16}, 36: {17, -9}, 45: {22, 0}, 46: {22, 0}}
		for tick := 1; tick <= 46; tick++ {
			c := Controls{Horizontal: -direction, Jump: true}
			if tick == 1 {
				c.Horizontal = direction
			}
			m.step(c, true)
			if want, ok := checks[tick]; ok && (m.X-startX != direction*want[0] || m.Y-startY != want[1] || m.grounded != (tick == 46)) {
				t.Fatalf("direction %d tick %d: position (%d,%d), grounded %v", direction, tick, m.X, m.Y, m.grounded)
			}
		}
	}
}

func TestMovementWalksAcrossGroundCells(t *testing.T) {
	m := NewMovement()
	for range 50 {
		m.step(Controls{Horizontal: 1}, true)
	}
	if m.X != 90 || m.Y != 177 || !m.grounded {
		t.Fatalf("walk ended at (%d,%d), grounded %v", m.X, m.Y, m.grounded)
	}
}

func TestMovementClimbsAndDetaches(t *testing.T) {
	m := NewMovement()
	m.X = 216
	for range 5 {
		m.step(Controls{Vertical: -1}, true)
	}
	if m.Y != 175 || m.ladder == nil {
		t.Fatalf("after five climb updates: Y=%d, attached=%v", m.Y, m.ladder != nil)
	}
	m.step(Controls{}, true)
	if m.Y != 175 || m.ladder == nil {
		t.Fatal("idle input changed ladder attachment or position")
	}
	for range 50 {
		m.step(Controls{Vertical: -1}, true)
		if m.ladder == nil {
			if m.Y+playerFoot != 172 {
				t.Fatalf("detached at foot %d, want 172", m.Y+playerFoot)
			}
			return
		}
	}
	t.Fatal("did not detach at the top of the ladder")
}

func TestMovementJumpReboundsAtWall(t *testing.T) {
	m := NewMovement()
	m.X = 0
	m.step(Controls{Horizontal: -1, Jump: true}, true)
	m.step(Controls{}, true)
	if m.X != 0 || m.jumpDirection != 1 || !m.descending {
		t.Fatalf("wall contact: X=%d, direction=%d, descending=%v", m.X, m.jumpDirection, m.descending)
	}
}

func TestMovementTransfersBetweenGirderSteps(t *testing.T) {
	m := NewMovement()
	m.surfaces = []girder{{0, 20, 161}, {20, 40, 162}, {40, 60, 161}}
	m.X, m.Y = 0, 144
	for range 16 {
		m.step(Controls{Horizontal: 1}, true)
	}
	if m.Y != 145 || !m.grounded {
		t.Fatalf("downward transfer: Y=%d, grounded=%v", m.Y, m.grounded)
	}
	for range 20 {
		m.step(Controls{Horizontal: 1}, true)
	}
	if m.Y != 144 || !m.grounded {
		t.Fatalf("upward transfer: Y=%d, grounded=%v", m.Y, m.grounded)
	}
}

func TestMovementWalkOffAndLand(t *testing.T) {
	for _, gap := range []int{16, 17} {
		m := NewMovement()
		m.surfaces = []girder{{0, 20, 100}, {20, 60, 99 + gap}}
		m.X, m.Y = 15, 83
		m.step(Controls{Horizontal: 1}, true)
		for i := 0; i < gap && !m.grounded; i++ {
			m.step(Controls{}, true)
		}
		if m.Y != 82+gap || !m.grounded || m.Dead != (gap > 16) {
			t.Fatalf("fall %d: Y=%d, grounded=%v, dead=%v", gap, m.Y, m.grounded, m.Dead)
		}
	}
}
