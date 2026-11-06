package classic

import (
	"fmt"
	"testing"
)

// TestElevatorsLayoutMatchesTheReference compares the live records read
// from the original's Elevators stage (DESIGN section 4, M3 Elevators).
func TestElevatorsLayoutMatchesTheReference(t *testing.T) {
	const girders = "[0 194 20][20 194 20][40 194 20][60 194 20][80 194 20][100 194 20][120 194 20][140 194 20][160 194 20][180 194 20][200 194 20][220 194 20][240 194 20]" +
		"[0 174 30][0 138 30][0 90 30][80 160 30][70 90 10][80 90 30][150 176 30][190 170 10][210 164 10][230 158 30][240 132 20][220 126 10][200 120 10][170 114 20][160 84 30][200 76 10][220 70 10][240 64 20][220 98 20][240 98 20]" +
		"[0 50 20][20 50 20][40 50 20][60 50 20][80 50 20][100 50 20][120 50 20][140 50 20][160 50 20][180 50 30][110 24 20][130 24 20][150 24 10]" +
		"[40 194 20][40 150 20][40 108 20][120 56 20][120 150 20][120 108 20]"
	const ladders = "[10 144 30][20 96 42][80 96 64][100 96 64][240 138 20][220 104 22][170 90 24][240 70 28][200 56 20][150 30 20]"
	l := elevatorsLayout()
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

func elevatorsStage(level int) *stage {
	rng := random{1}
	updates := 0
	return newStage(elevatorsFamily, &rng, arcadeSkill, level, &updates)
}

func TestElevatorsSetup(t *testing.T) {
	s := elevatorsStage(1) // displayed level 2: one spring
	if s.oil.present || len(s.pickups) != 0 || len(s.items) != 3 || s.items[0].value != 500 {
		t.Errorf("setup: oil %v, hammers %d, items %v", s.oil.present, len(s.pickups), s.items)
	}
	if !s.springs[0].active || s.springs[1].active || s.springs[0].delay != 135 || s.springs[0].drop != 228 {
		t.Errorf("springs %+v", s.springs)
	}
	if !s.fireballs[0].active || !s.fireballs[1].active || s.fireballs[2].active {
		t.Error("want the two starting fireballs")
	}
	if elevatorsStage(5).fireballs[2].active != true {
		t.Error("displayed level 6 has a third starting fireball")
	}
	if elevatorsStage(0).springs[0].active {
		t.Error("displayed level 1 Arcade has no springs")
	}
}

func TestElevatorsMoveAndWrap(t *testing.T) {
	s := elevatorsStage(1)
	for range 3 {
		s.moveElevators()
	}
	base := len(s.girders) - 6
	if s.elevators[0].y != 193 || s.girders[base].y != 193 || s.elevators[3].y != 57 {
		t.Fatalf("after three calls: %+v", s.elevators)
	}
	s.elevators[0].y, s.elevators[3].y = 44, 200
	for range 3 {
		s.moveElevators()
	}
	if s.elevators[0].y != elevatorBottom || s.elevators[3].y != elevatorTop {
		t.Errorf("wrap: rising %d, descending %d", s.elevators[0].y, s.elevators[3].y)
	}
}

func TestPlayerRidesAnElevator(t *testing.T) {
	s := elevatorsStage(1)
	m := s.player
	base := len(s.girders) - 6
	m.X, m.Y = 38, s.girders[base+1].y-16 // on the left cell at 150
	start := m.Y
	for range 30 {
		m.step(Controls{}, true)
		s.moveElevators()
	}
	// The cell rises ten pixels; the player follows on the next update.
	if m.Dead || m.Y != start-9 {
		t.Errorf("after 30 updates on a rising cell: Y %d, want %d", m.Y, start-9)
	}
}

func TestSpringCycle(t *testing.T) {
	s := elevatorsStage(1)
	p := &s.springs[0]
	p.delay = 0
	s.moveSprings()
	if p.x != -24 {
		t.Fatal("moved on the first call")
	}
	s.moveSprings()
	if p.x != -20 || p.y != 22 {
		t.Fatalf("first move to (%d,%d), want (-20,22)", p.x, p.y)
	}
	if !p.compressing {
		t.Error("a descent to Y=22 did not start a compression")
	}
	calls, stopped := 0, false
	var before uint32
	for p.drop == 228 && calls < 2000 {
		if p.dx == 0 && !stopped {
			stopped = true
			if p.x < 228 {
				t.Fatalf("stopped at X %d before its drop point", p.x)
			}
		}
		before = s.rng.state
		s.moveSprings()
		calls++
	}
	if !stopped {
		t.Fatal("the spring never reached its drop point")
	}
	probe := random{before}
	if p.x != -24-probe.bucket(3) || p.y != 19 || p.drop != 222 || p.delay != 20 || p.dx != 1 {
		t.Errorf("respawn %+v", *p)
	}
}

func TestElevatorsEnvironmentKills(t *testing.T) {
	s := elevatorsStage(1)
	m := s.player
	s.hammer.held, s.skill = true, easySkill // no bypass applies
	for _, c := range []struct {
		x, y int
		dies bool
	}{
		{100, 50, true}, {136, 50, false}, {100, 57, false},
		{70, 178, true}, {27, 166, true}, {26, 166, false}, {60, 166, false}, {110, 166, true},
	} {
		m.X, m.Y = c.x, c.y
		if got := s.contact(); got != c.dies {
			t.Errorf("at (%d,%d) died %v, want %v", c.x, c.y, got, c.dies)
		}
	}
}

func TestElevatorsCompletion(t *testing.T) {
	s := elevatorsStage(1)
	s.player.Y = 9
	if s.completed() {
		t.Fatal("complete at Y 9")
	}
	s.player.Y = 8
	if !s.completed() {
		t.Error("not complete at Y 8")
	}
}

func TestSpringsCountInTheJumpScan(t *testing.T) {
	s := elevatorsStage(1)
	m := s.player
	for i := range s.fireballs {
		s.fireballs[i].active = false
	}
	m.X, m.Y = 100, 10
	s.springs[0] = spring{active: true, x: 110, y: 30}
	s.scanJump()
	if m.jumpCount != 1 {
		t.Errorf("count %d with a spring below, want 1", m.jumpCount)
	}
}
