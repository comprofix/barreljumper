package classic

// Thrower poses. The renderer chooses art for them; they carry no rules.
const (
	throwerReady = iota
	throwerPickUp
	throwerThrow
	throwerTauntA
	throwerTauntB
)

// thrower is the Barrels release state machine verified against the
// reference in DESIGN section 4, Barrel release.
type thrower struct {
	x, y                      int
	phase, counter, threshold int
	taunting, holding         bool
	lastRegular               bool
	released                  int // barrels created this stage
	pose                      int
}

func newThrower(at point, level, skill int) thrower {
	threshold := max(16, 50-2*level-10*skill)
	return thrower{x: at.x, y: at.y, phase: 2, threshold: threshold, counter: threshold - 1, holding: true, pose: throwerReady}
}

func (s *stage) release() {
	t := &s.thrower
	if t.taunting {
		s.taunt()
		return
	}
	t.counter++
	if t.counter <= t.threshold {
		return
	}
	t.counter = 0
	switch t.phase {
	case 0:
		t.phase, t.pose = 1, throwerReady
		if s.rng.bucket(6) == 0 {
			t.taunting, t.pose = true, throwerTauntA
		}
	case 1:
		if !t.holding {
			t.phase, t.pose, t.holding = 2, throwerPickUp, true
		} else {
			t.phase, t.pose = 0, throwerThrow
			s.createBarrel(regularBarrel)
		}
	case 2:
		t.phase, t.pose = 1, throwerReady
		switch {
		case t.released == 0 && s.level == 0:
			s.createBarrel(verticalBarrel)
		case t.released == 0:
			s.createBarrel(driftBarrel)
		case s.rng.bucket(7-s.skill) == 0 && t.lastRegular:
			s.createBarrel(verticalBarrel + s.rng.bucket(3))
		}
	}
}

func (s *stage) taunt() {
	t := &s.thrower
	t.counter++
	if t.counter <= 24 {
		return
	}
	t.counter = 0
	t.phase++
	t.pose = throwerTauntA
	if t.phase%2 == 0 {
		t.pose = throwerTauntB
	}
	if t.phase >= 4 {
		t.taunting = false
		t.phase, t.counter, t.pose = 1, t.threshold, throwerReady
	}
}

// createBarrel releases the held barrel into the first free slot.
func (s *stage) createBarrel(kind int) {
	t := &s.thrower
	if !t.holding {
		return
	}
	slot := -1
	for i := range s.barrels {
		if !s.barrels[i].active {
			slot = i
			break
		}
	}
	if slot < 0 {
		return
	}
	b := barrel{active: true, kind: kind, ladder: -1, bounceMax: 0}
	if kind == regularBarrel {
		b.x, b.y, b.dx = t.x+30, 29, 1
		t.lastRegular = true
	} else {
		b.x, b.y, b.fast = t.x+11, t.y+16, 1
		if kind != verticalBarrel {
			b.dx = 2
		}
		t.lastRegular = false
	}
	b.bounceDir = 1
	b.oil = t.released == 0 || t.released%(14-s.skill) == 0
	s.barrels[slot] = b
	t.holding = false
	t.released++
}
