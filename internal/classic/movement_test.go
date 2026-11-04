package classic

import "testing"

func settledMovement(t *testing.T) *Movement {
	t.Helper()
	m := NewMovement()
	for range 3 {
		m.step(Controls{}, true)
	}
	if m.X != 40 || m.Y != 178 || !m.grounded {
		t.Fatalf("spawn did not settle at (40,178): %+v", m)
	}
	return m
}

// Expected positions and boundaries come from DESIGN's M1 comparisons,
// including assisted original-game replays, rather than model snapshots.
func TestBarrelsJumpReferenceVector(t *testing.T) {
	for _, direction := range []int{-1, 0, 1} {
		m := settledMovement(t)
		startX, startY := m.X, m.Y
		checks := map[int][2]int{1: {0, -1}, 10: {5, -10}, 22: {11, -16}, 36: {17, -9}, 45: {22, 0}, 46: {22, 0}, 47: {22, 0}}
		for tick := 1; tick <= 47; tick++ {
			c := Controls{Horizontal: -direction, Jump: true}
			if tick == 1 {
				c.Horizontal = direction
			}
			m.step(c, true)
			if want, ok := checks[tick]; ok && (m.X-startX != direction*want[0] || m.Y-startY != want[1] || m.grounded != (tick >= 46)) {
				t.Fatalf("direction %d tick %d: position (%d,%d), grounded %v", direction, tick, m.X, m.Y, m.grounded)
			}
			if m.jumping != (tick < 47) {
				t.Fatalf("tick %d: jumping=%v", tick, m.jumping)
			}
		}
	}
}

func TestArcadeWalkCadenceAndIdle(t *testing.T) {
	m := settledMovement(t)
	wantX := []int{40, 41, 42, 43, 44, 45, 45, 46, 47, 48, 49, 50}
	for i, want := range wantX {
		m.step(Controls{Horizontal: 1}, true)
		if m.X != want || m.Y != 178 || !m.grounded {
			t.Fatalf("walk call %d: (%d,%d), grounded=%v", i+1, m.X, m.Y, m.grounded)
		}
		for range 7 {
			m.step(Controls{}, true)
		}
	}
	m = settledMovement(t)
	for range 50 {
		m.step(Controls{Horizontal: 1}, false)
	}
	if m.X != 90 {
		t.Fatalf("Easy walk reached X=%d, want 90", m.X)
	}
}

func TestRearmedJumpWaitsForLandingStateToClear(t *testing.T) {
	m := settledMovement(t)
	m.step(Controls{Jump: true}, true)
	m.step(Controls{}, true)
	for tick := 3; tick <= 47; tick++ {
		m.step(Controls{Jump: true}, true)
	}
	if m.Y != 178 || m.jumping || !m.grounded {
		t.Fatal("held rearmed jump started before the landing state cleared")
	}
	m.step(Controls{Jump: true}, true)
	if m.Y != 177 || !m.jumping {
		t.Fatal("rearmed held jump did not start on the next eligible call")
	}
}

func TestStageEntryNeedsJumpRelease(t *testing.T) {
	m := NewMovement()
	for range 10 {
		m.step(Controls{Jump: true}, true)
	}
	if m.jumping || m.Y != 178 {
		t.Fatal("holding jump on stage entry bypassed release-to-rearm")
	}
	m.step(Controls{}, true)
	m.step(Controls{Jump: true}, true)
	if !m.jumping || m.Y != 177 {
		t.Fatal("release did not rearm jump after stage entry")
	}
}

func TestMovementClimbsAndDetaches(t *testing.T) {
	m := settledMovement(t)
	m.X = 216
	for tick := 1; tick <= 55; tick++ {
		m.step(Controls{Vertical: -1}, true)
		if want := 178 - 2*(tick/5); m.Y != want {
			t.Fatalf("climb call %d: Y=%d, want %d", tick, m.Y, want)
		}
		if tick == 7 {
			for range 8 {
				m.step(Controls{}, true)
			}
			if m.Y != 176 || m.ladder == nil {
				t.Fatal("idle changed attachment, position or climb cadence")
			}
		}
	}
	if m.ladder != nil || m.Y != 156 || !m.grounded {
		t.Fatalf("top transfer: Y=%d, attached=%v, grounded=%v", m.Y, m.ladder != nil, m.grounded)
	}
	for range 55 {
		m.step(Controls{Vertical: 1}, true)
	}
	if m.ladder != nil || m.Y != 178 || !m.grounded {
		t.Fatalf("bottom transfer: Y=%d, attached=%v, grounded=%v", m.Y, m.ladder != nil, m.grounded)
	}
}

func TestBrokenLadderCapsAndExits(t *testing.T) {
	for _, tc := range []struct {
		name                               string
		ladder, y, direction, calls, wantY int
		attached                           bool
	}{
		{"lower cap", 16, 176, -1, 30, 170, true},
		{"upper cap", 15, 156, 1, 30, 166, true},
		{"lower exit", 16, 176, 1, 5, 178, false},
		{"upper exit", 15, 156, -1, 15, 150, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := settledMovement(t)
			m.X, m.Y, m.ladder, m.grounded = 96, tc.y, &barrelsLadders[tc.ladder], false
			for range tc.calls {
				m.step(Controls{Vertical: tc.direction}, true)
			}
			if m.Y != tc.wantY || (m.ladder != nil) != tc.attached || m.grounded == tc.attached {
				t.Fatalf("Y=%d, attached=%v, grounded=%v", m.Y, m.ladder != nil, m.grounded)
			}
		})
	}
}

func TestJumpTakesPrecedenceOverLadder(t *testing.T) {
	m := settledMovement(t)
	m.X = 216
	m.step(Controls{Vertical: -1, Jump: true}, true)
	if !m.jumping || m.ladder != nil || m.Y != 177 {
		t.Fatalf("simultaneous jump/up: jumping=%v, attached=%v, Y=%d", m.jumping, m.ladder != nil, m.Y)
	}
}

func TestMovementJumpReboundsAtWalls(t *testing.T) {
	for _, x := range []int{0, maxPlayerX} {
		m := settledMovement(t)
		m.X = x
		direction := -1
		if x != 0 {
			direction = 1
		}
		for range 2 {
			m.step(Controls{Horizontal: direction, Jump: true}, true)
		}
		if m.X != x || m.Y != 176 || m.jumpDirection != -direction || m.jumpStep != 1 {
			t.Fatalf("wall %d: X=%d, Y=%d, direction=%d, vertical=%d", x, m.X, m.Y, m.jumpDirection, m.jumpStep)
		}
	}
	for _, start := range []int{11, 233} {
		m := settledMovement(t)
		m.X = start
		direction, wall := -1, 0
		if start == 233 {
			direction, wall = 1, maxPlayerX
		}
		for range 25 {
			m.step(Controls{Horizontal: direction, Jump: true}, true)
		}
		if m.X != wall || m.Y != 163 || m.jumpDirection != -direction || m.jumpStep != -1 {
			t.Fatalf("apex wall from %d: X=%d, Y=%d, direction=%d, vertical=%d", start, m.X, m.Y, m.jumpDirection, m.jumpStep)
		}
	}
}

func TestAcceleratedJumpFallAndLateWall(t *testing.T) {
	m := settledMovement(t)
	m.X, m.Y, m.jumping, m.falling, m.grounded = 244, 156, true, true, false
	m.jumpDisplacement, m.jumpStep = 13, 1
	for i, want := range []int{157, 159, 161} {
		m.step(Controls{}, true)
		if m.Y != want || m.jumpStep != 2 {
			t.Fatalf("fall call %d: Y=%d, vertical=%d", i+1, m.Y, m.jumpStep)
		}
	}
	m.jumpDirection = 1
	m.step(Controls{}, true)
	if m.X != 244 || m.jumpDirection != 0 || m.jumpStep != 2 {
		t.Fatal("late wall contact rebounded instead of clearing horizontal travel")
	}
}

func TestMovementTransfersBetweenGirderSteps(t *testing.T) {
	for _, tc := range []struct{ y, wantY int }{{148, 149}, {122, 121}} {
		m := settledMovement(t)
		m.X, m.Y = 70, tc.y
		for range 8 {
			m.step(Controls{Horizontal: 1}, true)
		}
		if m.X != 76 || m.Y != tc.wantY || m.falling != (tc.y == 148) {
			t.Fatalf("step from Y=%d: X=%d, Y=%d, falling=%v", tc.y, m.X, m.Y, m.falling)
		}
		if m.falling {
			m.step(Controls{Horizontal: 1}, true)
			if m.X != 76 || m.Y != 149 || !m.grounded {
				t.Fatal("descending step did not recover support without walking")
			}
		}
	}
}

func TestMovementFallContactIncludesLandingCall(t *testing.T) {
	for _, distance := range []int{16, 17} {
		m := settledMovement(t)
		m.X, m.Y, m.falling, m.grounded = 244, 179-distance, true, false
		for range distance {
			m.step(Controls{}, true)
		}
		if m.Y != 178 || !m.grounded || m.Dead != (distance == 17) {
			t.Fatalf("fall %d: Y=%d, grounded=%v, dead=%v", distance, m.Y, m.grounded, m.Dead)
		}
	}
	m := settledMovement(t)
	m.X, m.Y = 234, 156
	for range 25 {
		m.step(Controls{Horizontal: 1}, true)
	}
	if m.X != 236 || m.Y != 178 || !m.Dead {
		t.Fatalf("edge fall: (%d,%d), dead=%v", m.X, m.Y, m.Dead)
	}
}
