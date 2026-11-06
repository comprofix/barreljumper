package classic

import (
	"fmt"
	"testing"
)

// TestConveyorLayoutMatchesTheReference compares the live records read
// from the original's Conveyor stage (DESIGN section 4, M3 Conveyor). The
// retracting ladders are the two records after the fixed ten.
func TestConveyorLayoutMatchesTheReference(t *testing.T) {
	const girders = "[0 194 20][20 194 20][40 194 20][60 194 20][80 194 20][100 194 20][120 194 20][140 194 20][160 194 20][180 194 20][200 194 20][220 194 20][240 194 20]" +
		"[0 156 20][20 156 20][40 156 20][60 156 20][80 156 20][100 156 20][120 156 20][140 156 20][160 156 20][180 156 20][200 156 20][220 156 20][240 156 20]" +
		"[10 118 20][30 118 20][70 118 20][80 118 20][100 118 20][120 118 20][140 118 20][160 118 20][170 118 20][210 118 20][230 118 20]" +
		"[0 80 20][20 80 20][40 80 20][60 80 20][80 80 20][100 80 20][140 80 20][160 80 20][180 80 20][200 80 20][220 80 20][240 80 20]" +
		"[0 42 20][20 42 20][40 42 20][60 42 20][80 42 20][100 42 20][120 42 20][140 42 20][160 42 20][180 42 20][200 42 20][220 42 20][240 42 20]" +
		"[110 22 20][120 22 20][140 22 20]"
	const ladders = "[20 162 32][90 162 32][160 162 32][230 162 32][30 86 32][90 86 32][160 86 32][220 86 32][70 124 32][180 124 32][20 56 24][230 56 24]"
	l := conveyorLayout()
	got := ""
	for _, g := range l.girders {
		got += fmt.Sprintf("[%d %d %d]", g.left, g.y, g.right-g.left)
	}
	if got != girders {
		t.Errorf("girders\n got %s\nwant %s", got, girders)
	}
	got = ""
	for _, r := range l.ladders {
		got += fmt.Sprintf("[%d %d %d]", r.x, r.top, r.height)
	}
	if got != ladders {
		t.Errorf("ladders\n got %s\nwant %s", got, ladders)
	}
}

func conveyorStage(level int) *stage {
	rng := random{1}
	updates := 0
	return newStage(conveyorFamily, &rng, arcadeSkill, level, &updates)
}

func TestBeltsCarryThePlayerAndThrower(t *testing.T) {
	s := conveyorStage(2)
	m := s.player
	m.X, m.Y = 100, 156-16 // standing on belt 0
	x, tx := m.X, s.thrower.x
	for range 3 {
		s.moveBelts()
	}
	if m.X != x+2 || s.thrower.x != tx+2 {
		t.Errorf("after three calls: player moved %d, thrower %d; want 2 each", m.X-x, s.thrower.x-tx)
	}
	m.jumping = true
	x = m.X
	s.moveBelts()
	if m.X != x {
		t.Error("a jumping player was carried")
	}
}

func TestBeltCadenceWrapsAsASignedByte(t *testing.T) {
	b := belt{cadence: 126}
	var moves []bool
	for range 5 {
		b.cadence++
		moves = append(moves, b.cadence%3 != 0)
	}
	// 127, -128, -127 all move; -126 is the next skip.
	if fmt.Sprint(moves) != "[true true true false true]" {
		t.Errorf("moves across the wrap %v", moves)
	}
}

func TestBeltsReverse(t *testing.T) {
	s := conveyorStage(2)
	b := &s.belts[1]
	moves := 0
	for b.dir == 1 {
		s.moveBelts()
		if b.cadence%3 != 0 {
			moves++
		}
	}
	if moves != 91 { // from counter 120 past the 210 threshold
		t.Errorf("belt 1 reversed after %d moves, want 91", moves)
	}
	if s.belts[2].dir != 1 || s.belts[3].dir != -1 {
		t.Error("the Y=80 belts should never reverse")
	}
}

func TestRetractingLadderCycle(t *testing.T) {
	s := conveyorStage(2)
	r := &s.retractors[0]
	l := &s.ladders[r.ladder]
	run := func(until func() bool) int {
		calls := 0
		for !until() {
			s.moveRetractors()
			calls++
		}
		return calls
	}
	// One step every nine calls from 8 to 16, then a further step clamps
	// at 16 and starts the short wait.
	if n := run(func() bool { return r.waiting }); n != 9*9 || r.extension != 16 ||
		l.top != 64 || l.height != 16 || l.kind != lowerSegment {
		t.Fatalf("clamped at 16 after %d calls: %+v, ladder %+v", n, *r, *l)
	}
	// A five-call wait, then the first shrink nine calls later.
	if n := run(func() bool { return r.extension < 16 }); n != 5+9 {
		t.Errorf("left 16 after %d calls, want 14", n)
	}
	// Shrinking steps every five calls; the step below 0 clamps and waits.
	if n := run(func() bool { return r.waiting }); n != 16*5 || r.extension != 0 ||
		l.top != retractorBase || l.height != 32 || l.kind != fullLadder {
		t.Errorf("clamped at 0 after %d calls: %+v, ladder %+v", n, *r, *l)
	}
	if n := run(func() bool { return !r.waiting }); n != 277 {
		t.Errorf("waited %d calls at full extension, want 277", n)
	}
}

func TestPieSpawnAndDespawn(t *testing.T) {
	s := conveyorStage(2)
	if s.pieLimit() != 8 || s.piePeriod() != 108 {
		t.Fatalf("limit %d, period %d", s.pieLimit(), s.piePeriod())
	}
	for range 40 {
		before := s.rng.state
		s.spawnPie()
		probe := random{before}
		p := s.pies[0]
		want := pie{true, pieLeft, 149, 0}
		if probe.bucket(2) != 0 {
			want.belt = probe.bucket(2) + 2
			want.y, want.x = 73, pieLeft
			if want.belt == 3 {
				want.x = pieRight
			}
		}
		if p != want {
			t.Fatalf("spawned %+v, want %+v", p, want)
		}
		s.pies[0].active = false
	}
	s.pies[0] = pie{true, s.oil.x + 2, 73, 2}
	s.belts[2].cadence = 0
	s.moveBelts()
	if s.pies[0].active {
		t.Error("a belt 2 pie at the drum was not removed")
	}
}

func TestPieContactHitAndJump(t *testing.T) {
	s := conveyorStage(2)
	for i := range s.fireballs {
		s.fireballs[i].active = false
	}
	m := s.player
	m.X, m.Y = 100, 140
	s.pies[0] = pie{true, 111, 149 - 9, 0} // X+2 = 113 against the span 102..113
	if !s.contact() {
		t.Error("pie contact missed")
	}
	s.pies[0].x = 112
	if s.contact() {
		t.Error("pie past the span hit")
	}
	s.pies[0] = pie{true, 100, 160, 0}
	s.scanJump()
	if m.jumpCount != 1 {
		t.Errorf("jump count %d over a pie, want 1", m.jumpCount)
	}
	s.hammer = heldHammer{held: true, x: 100, y: 160}
	s.hitWithHammer()
	if s.pies[0].active || !s.smash.active {
		t.Error("the hammer did not smash the pie")
	}
}

func TestConveyorCompletionAndDrumFireballs(t *testing.T) {
	s := conveyorStage(2)
	s.player.Y = 27
	if s.completed() {
		t.Fatal("complete below the top row")
	}
	s.player.Y = 26
	if !s.completed() {
		t.Error("standing on the Y=42 row did not complete Conveyor")
	}
	before := s.rng.state
	s.spawnDrumFireball()
	probe := random{before}
	dir := -1
	if probe.bucket(2) == 0 {
		dir = 1
	}
	f := s.fireballs[0]
	if !f.active || f.x != s.oil.x+dir || f.y != s.oil.y-6 || !f.emerging || f.delay != 0 {
		t.Errorf("drum fireball %+v, direction %d", f, dir)
	}
}
