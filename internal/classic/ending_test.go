package classic

import "testing"

// toPlay ticks through the opening, floor card or screen ending until play
// begins, and drops the sounds they made.
func toPlay(t *testing.T, p *Play) *Play {
	t.Helper()
	for range 5000 {
		if p.phase == playing {
			p.sounds = 0
			return p
		}
		p.tick(Controls{}, false)
	}
	t.Fatalf("phase %d never reached play", p.phase)
	return p
}

// runFor ticks n times and reports the phase reached.
func runFor(p *Play, n int) int {
	for range n {
		p.tick(Controls{}, false)
	}
	return p.phase
}

// TestPresentationTimings follows the reference lengths: the opening, the
// floor card, the heart and carry-off after a clear, and the Rivets ending.
func TestPresentationTimings(t *testing.T) {
	p := NewPlay(nil, skillByID(t, "arcade"))
	if p.phase != opening {
		t.Fatalf("phase %d at the start, want the opening", p.phase)
	}
	if got := runFor(p, openingTicks-1); got != opening {
		t.Fatalf("phase %d before the opening's end", got)
	}
	if got := runFor(p, 1); got != floorCard || p.jingle != startJingle {
		t.Fatalf("phase %d after the opening, want the floor card", got)
	}
	if runFor(p, cardTicks-1) != floorCard || runFor(p, 1) != playing {
		t.Fatalf("the floor card did not last %d ticks", cardTicks)
	}
	p.stage.player.Y = barrelsGoal.y - supportFoot
	if got := runFor(p, 1); got != cheering {
		t.Fatalf("phase %d after a clear, want the heart", got)
	}
	if runFor(p, heartTicks) != carryOff {
		t.Fatal("no carry-off after the heart")
	}
	if runFor(p, carryTicks) != floorCard || p.stage.family != rivetsFamily {
		t.Fatalf("after the carry-off: phase %d, family %d", p.phase, p.stage.family)
	}
	toPlay(t, p)
	for i := range p.stage.rivets {
		p.stage.rivets[i].removed = true
	}
	if got := runFor(p, 1); got != rivetsDone || p.sounds&soundRumble == 0 {
		t.Fatalf("phase %d after the Rivets clear, want its ending", got)
	}
	runFor(p, rivetsLowered)
	if m := p.stage.player; m.Y+18 != 66 || m.X+16 != 126 {
		t.Errorf("worker at %d,%d beside the lowered captive", m.X, m.Y)
	}
	if runFor(p, rivetsEndTicks-rivetsLowered-1) != rivetsDone || runFor(p, 1) != floorCard || p.level != 1 {
		t.Fatalf("after the Rivets ending: phase %d, level %d", p.phase, p.level+1)
	}
}

// TestFloorCardAfterADeath checks the card comes before the restarted screen.
func TestFloorCardAfterADeath(t *testing.T) {
	p := toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	p.stage.bonus, p.updates = 100, 119
	p.tick(Controls{}, true)
	if got := runFor(p, deathFreezeTicks+deathSpinTicks+deathHoldTicks); got != floorCard {
		t.Fatalf("phase %d after a death, want the floor card", got)
	}
}
