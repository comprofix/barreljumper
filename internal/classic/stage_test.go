package classic

import (
	"fmt"
	"testing"

	"github.com/comprofix/barreljumper/internal/config"
)

// replayHazards runs the hazard half of an update with an idle player and
// no contact, as the reference probe ran it with the invincibility toggle.
func replayHazards(s *stage) {
	*s.updates++
	s.player.step(Controls{}, s.skill != easySkill)
	s.collectHammer()
	s.moveFireballs()
	s.release()
	s.moveBarrels()
}

func activeBarrels(s *stage) string {
	out := ""
	for i, b := range s.barrels {
		if b.active {
			out += fmt.Sprintf("[%d %v %d %d,%d %d %d]", i, b.oil, b.kind, b.x, b.y, b.dx, b.ladder)
		}
	}
	return out
}

// TestBarrelsMatchTheReferenceTrace checks the background DOSBox trace of
// Classic Arcade level 1 from RNG state 29039 (DESIGN section 4, Barrel
// release): releases, a ladder descent, the special barrel landing, and
// the drum igniting with its fireball. Before ignition no cosmetic draw
// shares the stream, so the RNG state is compared up to then.
func TestBarrelsMatchTheReferenceTrace(t *testing.T) {
	type check struct {
		update, phase, counter, released int
		taunting                         bool
		rng                              uint32
		barrels                          string
	}
	checks := []check{
		{2, 1, 0, 1, false, 29039, "[0 true 1 41,22 0 -1]"},
		{43, 2, 0, 1, false, 2555700081, "[0 true 1 41,48 0 -1]"},
		{84, 1, 0, 1, false, 1290118531, "[0 true 1 41,64 0 -1]"},
		{125, 0, 0, 2, false, 341522469, "[0 true 1 41,87 0 -1][1 false 0 60,29 1 -1]"},
		{166, 1, 0, 2, true, 1729670883, "[0 true 1 41,108 0 -1][1 false 0 106,33 1 11]"},
		{250, 2, 8, 2, false, 336944309, "[0 true 1 41,152 0 -1][1 false 0 86,82 -1 5]"},
		{316, 1, 33, 2, false, 1454093282, "[0 true 0 41,183 -1 -1][1 false 0 167,94 1 -1]"},
		{328, 0, 4, 3, false, 0, "[1 false 0 185,95 1 -1][2 false 0 66,29 1 -1]"},
	}
	rng := random{29039}
	updates := 0
	s := newStage(barrelsFamily, &rng, arcadeSkill, 0, &updates)
	next := 0
	for u := 1; next < len(checks); u++ {
		replayHazards(s)
		c := checks[next]
		if u != c.update {
			continue
		}
		next++
		tr := s.thrower
		got := check{u, tr.phase, tr.counter, tr.released, tr.taunting, rng.state, activeBarrels(s)}
		if c.rng == 0 {
			got.rng = 0 // the drum's cosmetic flame draws start at ignition
		}
		if got != c {
			t.Errorf("update %d:\n got %+v\nwant %+v", u, got, c)
		}
	}
	if !s.oil.burning {
		t.Fatal("the oil barrel did not ignite the drum at update 328")
	}
	f := s.fireballs[0]
	if !f.active || f.x != 19 || f.y != 164 || f.dx != 1 || f.dy != -1 || !f.emerging || f.delay != 100 {
		t.Errorf("fireball after ignition %+v", f)
	}
}

// TestBarrelRowProbe: the reference picks a barrel's row by the last cell
// of rows 1 and 3, so a barrel at the low right end of row 3 stays on it.
func TestBarrelRowProbe(t *testing.T) {
	g := barrelsGirders()
	for _, c := range []struct{ y, first int }{
		{97, 37},  // right end of row 3 (cell 48 at y 108)
		{161, 13}, // right end of row 1
		{86, 37},  // left end of row 3
		{183, 0},
		{29, 61},
	} {
		if first, _ := barrelRow(g, c.y); first != c.first {
			t.Errorf("barrel at y %d scans row from %d, want %d", c.y, first, c.first)
		}
	}
}

func TestThrowerThreshold(t *testing.T) {
	for _, c := range []struct{ level, skill, want int }{
		{0, arcadeSkill, 40}, {0, easySkill, 50}, {0, hardSkill, 30}, {5, arcadeSkill, 30}, {20, hardSkill, 16},
	} {
		if got := newThrower(point{30, 6}, c.level, c.skill).threshold; got != c.want {
			t.Errorf("level %d skill %d threshold %d, want %d", c.level, c.skill, got, c.want)
		}
	}
}

func TestBonusFallsAndRunsOut(t *testing.T) {
	if stageBonus(0) != 5000 || stageBonus(2) != 7000 || stageBonus(9) != 8000 {
		t.Fatal("stage bonus table")
	}
	rng := random{1}
	updates := 118
	s := newStage(barrelsFamily, &rng, arcadeSkill, 0, &updates)
	s.update(Controls{}, true)
	if s.bonus != 5000 {
		t.Fatalf("bonus %d before the 120th update", s.bonus)
	}
	s.update(Controls{}, false) // a skipped tick does not count
	s.update(Controls{}, true)
	if s.bonus != 4900 {
		t.Fatalf("bonus %d at the 120th update", s.bonus)
	}
	s.bonus, updates = 100, 239
	if got := s.update(Controls{}, true); got != stageDied || s.bonus != 0 {
		t.Errorf("bonus run out: outcome %d, bonus %d", got, s.bonus)
	}
}

func TestReachingTheGoalClearsTheStage(t *testing.T) {
	rng := random{1}
	updates := 0
	s := newStage(barrelsFamily, &rng, arcadeSkill, 0, &updates)
	s.player.X, s.player.Y = 120, barrelsGoal.y-supportFoot+1
	if got := s.update(Controls{}, false); got == stageCleared {
		t.Fatal("cleared one pixel short of the goal")
	}
	s.player.Y--
	if got := s.update(Controls{}, false); got != stageCleared || s.earned != 5000 {
		t.Errorf("at the goal: outcome %d, earned %d", got, s.earned)
	}
}

// TestEveryClearCreditsTheBonus checks that each family's clear adds the
// remaining bonus once (DESIGN section 4, M5 natural clears).
func TestEveryClearCreditsTheBonus(t *testing.T) {
	for _, tc := range []struct {
		name   string
		family int
		finish func(s *stage)
	}{
		{"barrels", barrelsFamily, func(s *stage) { s.player.X, s.player.Y = 120, barrelsGoal.y-supportFoot }},
		{"rivets", rivetsFamily, func(s *stage) {
			for i := range s.rivets {
				s.rivets[i].removed = true
			}
		}},
		{"elevators", elevatorsFamily, func(s *stage) { s.player.Y = 24 - supportFoot - 1 }},
		{"conveyor", conveyorFamily, func(s *stage) { s.player.Y = 42 - supportFoot - 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rng := random{1}
			updates := 0
			s := newStage(tc.family, &rng, arcadeSkill, 0, &updates)
			s.bonus = 3000
			tc.finish(s)
			if got := s.update(Controls{}, false); got != stageCleared || s.earned != 3000 {
				t.Errorf("outcome %d, earned %d; want a clear crediting 3000", got, s.earned)
			}
		})
	}
}

func TestContact(t *testing.T) {
	rng := random{1}
	updates := 0
	s := newStage(barrelsFamily, &rng, arcadeSkill, 0, &updates)
	m := s.player // (40,176) facing right: span 42..53
	s.barrels[0] = barrel{active: true, x: 51, y: 176, ladder: -1}
	if !s.contact() {
		t.Error("barrel overlapping the player's right edge missed")
	}
	s.barrels[0].x = 52
	if s.contact() {
		t.Error("barrel just past the player's span hit")
	}
	s.barrels[0].x = 31
	if !s.contact() {
		t.Error("barrel overlapping the player's left edge missed")
	}
	s.barrels[0].x = 30
	if s.contact() {
		t.Error("barrel just short of the player's span hit")
	}
	s.barrels[0].x = 51
	s.hammer.held, s.skill = true, easySkill
	if s.contact() {
		t.Error("Easy with a hammer did not guard")
	}
	s.skill = arcadeSkill
	if !s.contact() {
		t.Error("Arcade with a hammer guarded")
	}
	s.barrels[0].active, s.hammer.held = false, false
	m.X, m.Y = s.oil.x, s.oil.y-6
	if !s.contact() {
		t.Error("the drum did not kill while unlit")
	}
}

func TestJumpAwards(t *testing.T) {
	for count, want := range []int{0, 100, 300, 800, 800} {
		if got := jumpAward(count); got != want {
			t.Errorf("count %d awards %d, want %d", count, got, want)
		}
	}
}

func TestHammerPickupSwingAndExpiry(t *testing.T) {
	rng := random{1}
	updates := 0
	s := newStage(barrelsFamily, &rng, arcadeSkill, 0, &updates)
	m := s.player
	m.X, m.Y = 222, 148
	s.collectHammer()
	if s.hammer.held {
		t.Fatal("collected without jumping")
	}
	m.jumping, m.jumpDisplacement = true, 10
	s.collectHammer()
	if !s.hammer.held || !m.Hammer || s.pickups[0].active {
		t.Fatal("jumping into the hammer did not collect it")
	}
	m.jumping, m.jumpDisplacement = false, 0
	warned := false
	calls := 0
	for s.hammer.held && calls < 2000 {
		s.swingHammer()
		calls++
		warned = warned || s.hammer.warning()
	}
	if s.hammer.held || m.Hammer {
		t.Fatal("the hammer never expired")
	}
	if !warned || s.hammer.age != hammerExpiryAge+1 {
		t.Errorf("expiry: warned %v, age %d", warned, s.hammer.age)
	}
}

func TestHammerSmash(t *testing.T) {
	rng := random{7}
	updates := 0
	s := newStage(barrelsFamily, &rng, arcadeSkill, 0, &updates)
	s.hammer = heldHammer{held: true, x: 60, y: 170}
	s.barrels[0] = barrel{active: true, x: 60, y: 168, ladder: -1}
	s.fireballs[0] = fireball{active: true, x: 58, y: 166}
	s.hitWithHammer()
	if s.fireballs[0].active == s.barrels[0].active || !s.barrels[0].active {
		t.Fatal("the hammer should take the fireball first, and only one target")
	}
	calls := 0
	for s.smash.active {
		s.updateSmash()
		calls++
	}
	if calls != 87 {
		t.Errorf("smash lasted %d calls, want 87", calls)
	}
	if e := s.earned; e != 300 && e != 500 && e != 800 {
		t.Errorf("smash awarded %d", e)
	}
}

func skillByID(t *testing.T, id string) config.Skill {
	for _, sk := range config.Skills {
		if sk.ID == id {
			return sk
		}
	}
	t.Fatalf("no skill %q", id)
	return config.Skill{}
}

func TestPlayLosesALifeThenEnds(t *testing.T) {
	p := toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	if p.reserves != 2 {
		t.Fatalf("reserves %d at the start, want 2", p.reserves)
	}
	die := func() {
		p.stage.bonus, p.updates = 100, 119
		p.tick(Controls{}, true)
		if p.phase != dying {
			t.Fatalf("phase %d after the bonus ran out", p.phase)
		}
		for i := 0; i < deathFreezeTicks+deathSpinTicks+deathHoldTicks; i++ {
			p.tick(Controls{}, true)
		}
		if p.phase == floorCard {
			toPlay(t, p)
		}
	}
	first := p.stage
	die()
	if p.phase != playing || p.reserves != 1 || p.stage == first || p.stage.bonus != 5000 {
		t.Fatalf("after a death: phase %d, reserves %d", p.phase, p.reserves)
	}
	die()
	die()
	if p.phase != gameOver {
		t.Errorf("phase %d after the last life", p.phase)
	}
	if NewPlay(nil, skillByID(t, "easy")).reserves != 4 {
		t.Error("Easy should start with five lives")
	}
}

func TestPlayClearAdvancesAndAwardsLives(t *testing.T) {
	p := toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	p.score = 6000
	p.stage.player.Y = barrelsGoal.y - supportFoot
	p.tick(Controls{}, false)
	if p.phase != cheering || p.score != 11000 || p.reserves != 3 {
		t.Fatalf("clear: phase %d, score %d, reserves %d", p.phase, p.score, p.reserves)
	}
	toPlay(t, p)
	if p.phase != playing || p.stage.family != rivetsFamily || p.level != 0 {
		t.Errorf("after the level 1 Barrels clear: family %d, level %d", p.stage.family, p.level+1)
	}
	p.stage.earned = 39000
	p.collect()
	if p.reserves != 4 || p.nextLife != 100000 {
		t.Errorf("at 50000: reserves %d, next life %d", p.reserves, p.nextLife)
	}
}
