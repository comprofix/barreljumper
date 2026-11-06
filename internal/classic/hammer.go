package classic

// hammerPickup is a collectable hammer: an 10 x 8 box at its origin.
type hammerPickup struct {
	active bool
	x, y   int
}

// heldHammer is the swinging hammer (DESIGN section 4, Hammer). Its head
// origin is recomputed from the player each eligible update.
type heldHammer struct {
	held         bool
	x, y         int
	counter, age int
	forward      bool
}

// Hammer timing from the reference swing routine.
const (
	hammerPhaseCalls = 7
	hammerWarnAge    = 48
	hammerExpiryAge  = 64
)

// smash is the destruction animation that replaces ordinary updates.
type smash struct {
	active       bool
	x, y         int
	frame, timer int
	points       int
	kind         int // what was smashed, for the debris drawn
}

// What a smash destroyed. The kind is drawn only; it carries no rules.
const (
	smashedBarrel = iota
	smashedFlame
	smashedPie
)

func (s *stage) collectHammer() {
	m := s.player
	if m.jumpDisplacement > 16 || !m.jumping {
		return
	}
	for i := range s.pickups {
		h := &s.pickups[i]
		if h.active && h.x < m.X+14 && h.x+10 > m.X+4 && h.y < m.Y+16 && h.y+8 > m.Y {
			h.active = false
			s.hammer = heldHammer{held: true, x: h.x, y: h.y}
			s.sounds |= soundHammer
			m.Hammer = true
			return
		}
	}
}

func (s *stage) swingHammer() {
	m, h := s.player, &s.hammer
	if m.jumping && m.jumpDisplacement < -4 {
		h.x, h.y = m.X+6, m.Y-9
		if m.Facing > 0 {
			h.x = m.X + 2
		}
		return
	}
	h.counter++
	if h.counter > hammerPhaseCalls {
		h.counter = 0
		if !m.falling {
			h.forward = !h.forward
			if h.forward {
				s.sounds |= soundSwing
			}
		}
		if !h.forward {
			h.counter++
		} else {
			h.age++
		}
		if h.age > hammerExpiryAge {
			s.dropHammer()
		}
	}
	if s.smash.active && h.age != 0 {
		return
	}
	if !h.forward {
		h.x, h.y = m.X+6, m.Y-9
		if m.Facing > 0 {
			h.x = m.X + 2
		}
		return
	}
	h.x, h.y = m.X-9, m.Y+5
	if m.Facing > 0 {
		h.x = m.X + 14
	}
}

func (s *stage) dropHammer() {
	s.hammer.held = false
	s.player.Hammer = false
}

// HammerWarning reports the expiry flash for the renderer.
func (h heldHammer) warning() bool {
	return h.age > hammerWarnAge && h.age%4 < 2
}

// hitWithHammer destroys at most one target under the hammer head,
// fireballs before barrels.
func (s *stage) hitWithHammer() {
	h := s.hammer
	left, top, w, ht := h.x, h.y, 12, 6
	if h.forward {
		w, ht = 10, 10
		if s.player.Facing > 0 {
			left += 2
		}
	}
	for i := range s.fireballs {
		f := &s.fireballs[i]
		if f.active && f.x+2 < left+w && f.x+14 > left && f.y < top+ht && f.y+16 > top {
			f.active = false
			s.startSmash(f.x, f.y, smashedFlame)
			return
		}
	}
	for i := range s.barrels {
		b := &s.barrels[i]
		right := 14
		if b.kind != regularBarrel {
			right = 18
		}
		if b.active && b.x+2 < left+w && b.x+right > left && b.y < top+ht && b.y+11 > top {
			b.active = false
			s.startSmash(b.x, b.y, smashedBarrel)
			return
		}
	}
	for i := range s.pies {
		p := &s.pies[i]
		if p.active && p.x < left+w && p.x+18 > left && p.y < top+ht && p.y+8 > top {
			p.active = false
			s.startSmash(p.x, p.y, smashedPie)
			return
		}
	}
}

func (s *stage) startSmash(x, y, kind int) {
	s.smash = smash{active: true, x: x + 2, y: y + 1, kind: kind}
	s.sounds |= soundSmash
}

// updateSmash runs the destruction animation. The award is drawn on every
// call; the last draw is paid when the animation ends.
func (s *stage) updateSmash() {
	k := &s.smash
	k.points = [...]int{300, 500, 800}[s.rng.bucket(3)]
	k.timer++
	if k.timer <= 6 {
		return
	}
	if k.frame < 11 {
		k.timer = 0
	} else {
		k.timer = 4
	}
	k.frame++
	if k.frame > 12 {
		k.active = false
		x := s.player.X - 12
		if s.player.Facing > 0 {
			x = s.player.X + 18
		}
		s.award(k.points, x, s.player.Y+10)
	}
}
