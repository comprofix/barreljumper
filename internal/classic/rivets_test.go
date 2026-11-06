package classic

import (
	"fmt"
	"testing"
)

func newTestStage(family int) *stage {
	rng := random{1}
	updates := 0
	return newStage(family, &rng, arcadeSkill, 0, &updates)
}

// TestRivetsLayoutMatchesTheReference compares the live records read from
// the original's Rivets stage (DESIGN section 4, M3 Rivets setup).
func TestRivetsLayoutMatchesTheReference(t *testing.T) {
	const girders = "[0 194 20][20 194 20][40 194 20][60 194 20][80 194 20][100 194 20][120 194 20][140 194 20][160 194 20][180 194 20][200 194 20][220 194 20][240 194 20]" +
		"[10 162 10][20 162 20][40 162 20][70 162 20][90 162 20][110 162 20][130 162 20][150 162 20][170 162 20][200 162 10][210 162 20][230 162 20]" +
		"[20 130 20][40 130 20][70 130 20][90 130 20][110 130 20][130 130 20][150 130 20][170 130 20][200 130 20][220 130 20]" +
		"[30 98 10][40 98 20][70 98 20][90 98 20][110 98 20][130 98 20][150 98 20][170 98 20][200 98 20][220 98 10]" +
		"[40 66 20][70 66 20][90 66 20][110 66 20][130 66 20][150 66 20][170 66 20][200 66 20]" +
		"[60 24 20][80 24 20][100 24 20][120 24 20][140 24 20][160 24 20][180 24 20]" +
		"[60 162 10][60 130 10][60 98 10][60 66 10][190 162 10][190 130 10][190 98 10][190 66 10]"
	const ladders = "[10 168 26][125 168 26][240 168 26][20 136 26][80 136 26][170 136 26][230 136 26]" +
		"[30 104 26][125 104 26][220 104 26][40 72 26][80 72 26][170 72 26][210 72 26]"
	l := rivetsLayout()
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

func TestRivetRemovalAndSupport(t *testing.T) {
	s := newTestStage(rivetsFamily)
	m := s.player
	// Standing on the Y=162 row, centre just short of the X=60 rivet.
	m.X, m.Y, m.Facing = 52, 146, 1
	s.removeRivets()
	if s.rivets[0].removed {
		t.Fatal("removed with the centre at rivet X")
	}
	m.X = 53 // centre 61 = rivet X+1
	s.removeRivets()
	if !s.rivets[0].removed || s.earned != 100 {
		t.Fatalf("walking over: removed %v, earned %d", s.rivets[0].removed, s.earned)
	}
	support := &s.girders[len(s.girders)-8]
	m.X = 60
	s.removeRivets()
	if support.right == support.left {
		t.Fatal("the support went while the player was still on it")
	}
	m.Facing = -1
	s.removeRivets()
	if support.right != support.left {
		t.Fatal("turning round did not retire the support")
	}
	if support.supports(m.X, m.Y+supportFoot) {
		t.Error("a retired rivet still supports the player")
	}
	// A jump over a rivet counts from its takeoff height.
	m.X, m.Y, m.Facing = 183, 146-10, 1
	m.jumping, m.jumpDisplacement = true, -10
	s.removeRivets()
	if !s.rivets[4].removed {
		t.Error("jumping over the X=190 rivet did not remove it")
	}
	for i := range s.rivets {
		s.rivets[i].removed = true
	}
	m.jumping = false
	if !s.completed() {
		t.Error("removing every rivet did not complete Rivets")
	}
}

func TestLostItems(t *testing.T) {
	s := newTestStage(rivetsFamily)
	m := s.player
	it := s.items[0] // (144,184)
	m.X, m.Y = it.x-1, 178
	m.jumping = true
	s.collectItems()
	if !s.items[0].active {
		t.Fatal("collected while jumping")
	}
	m.jumping = false
	s.collectItems()
	if s.items[0].active || s.earned != 300 {
		t.Fatalf("collect: active %v, earned %d", s.items[0].active, s.earned)
	}
	for level, want := range []int{300, 500, 800, 800} {
		if got := itemValue(level); got != want {
			t.Errorf("level %d item worth %d, want %d", level+1, got, want)
		}
	}
}

func TestRivetsTauntsAndSpawnsFireballs(t *testing.T) {
	s := newTestStage(rivetsFamily)
	if p := s.fireballPeriod(); p != 216 {
		t.Fatalf("fireball period %d, want 216 at level 1 Arcade", p)
	}
	if s.fireballLimit() != 5 {
		t.Errorf("Rivets fireball limit %d, want 5", s.fireballLimit())
	}
	*s.updates = 199
	*s.updates++
	s.stageRoutines()
	if !s.thrower.taunting {
		t.Fatal("no taunt at update 200")
	}
	*s.updates = 215
	*s.updates++
	before := s.rng.state
	s.stageRoutines()
	probe := random{before}
	i := probe.bucket(4)
	f := s.fireballs[0]
	if !f.active || f.x != 248-10*i || f.y != 176-32*i || f.dx != 1 || f.delay != 0 || f.emerging {
		t.Errorf("fireball %+v for draw %d with the player on the left", f, i)
	}
}

func TestThrowerContactOnBarrels(t *testing.T) {
	s := newTestStage(barrelsFamily)
	m := s.player
	m.X, m.Y = 54, 24 // top row, span 56..67 against the box X 32..67
	if !s.contact() {
		t.Error("walking into the thrower did not kill")
	}
	m.X = 66 // span 68..79
	if s.contact() {
		t.Error("contact beyond the thrower's box")
	}
}

func TestClassicScheduleProgression(t *testing.T) {
	p := NewPlay(nil, skillByID(t, "arcade"))
	var got []string
	for range 6 {
		got = append(got, fmt.Sprint(p.level+1, ":", p.stage.family))
		p.nextStage()
		p.enterStage()
	}
	// Level 2 Elevators and level 3 Conveyor/Elevators are not built yet.
	want := "[1:0 1:3 2:0 2:3 3:0 3:3]"
	if fmt.Sprint(got) != want {
		t.Errorf("screens %v, want %s", got, want)
	}
}
