package classic

// Barrel kinds. A special barrel falls through the girders until it reaches
// the ground row, then rolls as a regular barrel.
const (
	regularBarrel = iota
	verticalBarrel
	driftBarrel
	wanderBarrel // a drift barrel that can also reverse at random
)

// barrelGround is the special-barrel landing origin on the ground row.
const barrelGround = 183

// barrel is one of the sixteen Barrels hazard slots (DESIGN section 4,
// Hazards). Coordinates are the reference origin; a barrel rests 11 pixels
// above the girder surface it rolls on.
type barrel struct {
	active, oil bool
	kind        int
	fast        int // special descent: one extra pixel while set
	sub, frame  int // roll animation counters; frame gates the bounce
	ladder      int // descending ladder index, or -1
	falling     bool
	cadence     int
	x, y, dx    int
	// The bounce is drawn only, but it draws from the random stream.
	bounce, bounceMax, bounceDir int
}

// barrelRows are the first and last live girder cells of each Barrels row,
// bottom to top.
var barrelRows = [6][2]int{{0, 12}, {13, 24}, {25, 36}, {37, 48}, {49, 60}, {61, 72}}

// barrelRowProbe is the cell whose height selects each row: the reference
// tests the last cell of rows 1 and 3, which fall to the right, and the first
// cell of the others.
var barrelRowProbe = [6]int{0, 24, 25, 48, 49, 61}

// barrelRow returns the girder range scanned for a barrel at y: the highest
// row whose probe cell lies below it, or the ground row.
func barrelRow(girders []girder, y int) (first, last int) {
	for k := len(barrelRows) - 1; k > 0; k-- {
		if girders[barrelRowProbe[k]].y > y {
			return barrelRows[k][0], barrelRows[k][1]
		}
	}
	return barrelRows[0][0], barrelRows[0][1]
}

func (b *barrel) startBounce(height int) {
	b.bounceMax, b.bounce, b.bounceDir = height, 0, 1
}

// moveBarrels advances every active slot in slot order. A special barrel
// that reaches the ground ends the pass for the remaining slots, as in the
// reference.
func (s *stage) moveBarrels() {
	for i := range s.barrels {
		b := &s.barrels[i]
		if !b.active {
			continue
		}
		if b.kind != regularBarrel {
			if s.moveSpecialBarrel(b) {
				return
			}
			continue
		}
		first, last := barrelRow(s.girders, b.y)
		switch {
		case b.falling:
			s.fallBarrel(b, first, last)
		case b.ladder >= 0:
			s.descendBarrel(b, first, last)
		default:
			s.rollBarrel(b, first, last)
		}
	}
}

// moveSpecialBarrel reports whether the barrel landed on the ground row.
func (s *stage) moveSpecialBarrel(b *barrel) bool {
	b.cadence++
	if b.cadence <= 2 {
		return false
	}
	if s.level%2 != 0 {
		b.fast = 1
	}
	b.cadence = 0
	b.y += b.fast + 1
	if b.kind >= driftBarrel {
		b.x += b.dx
	}
	if b.x >= 242 || b.x <= 0 {
		b.dx = -b.dx
	} else if s.rng.bucket(80) == 0 && b.kind == wanderBarrel {
		b.dx = -b.dx
	}
	b.sub = (b.sub + 1) % 3
	if b.sub == 0 {
		b.frame = (b.frame + 1) % 4
		if b.frame == 0 {
			b.fast = (b.fast + 1) % 2
		}
	}
	if b.y < barrelGround+1 {
		return false
	}
	b.y, b.kind, b.dx = barrelGround, regularBarrel, -1
	b.startBounce(4)
	return true
}

func (s *stage) fallBarrel(b *barrel, first, last int) {
	b.cadence++
	if b.cadence <= 2 {
		return
	}
	b.cadence = 0
	b.x += b.dx
	b.y += 2
	for _, g := range s.girders[first : last+1] {
		if b.x+10 >= g.left+2 && b.x+2 < g.left+20 && b.y+11 > g.y {
			b.y = g.y - 11
			b.dx = -b.dx // Classic Barrels always reverses on landing
			b.falling = false
			if b.y > s.player.Y+36 && !b.oil && (b.x < 10 || b.x > 250) {
				b.dx = -b.dx
			}
			b.startBounce(3)
			return
		}
	}
}

func (s *stage) descendBarrel(b *barrel, first, last int) {
	b.cadence++
	if b.cadence <= 2 {
		return
	}
	b.cadence = 0
	b.y += 2
	b.sub = (b.sub + 1) % 3
	if b.sub == 0 {
		b.frame = (b.frame + 1) % 4
	}
	top := s.ladders[b.ladder].top
	for _, g := range s.girders[first : last+1] {
		if b.y > top && b.x+10 >= g.left && b.x+2 < g.left+20 &&
			b.y+11 >= g.y && b.y+11 < g.y+2 {
			b.y = g.y - 11
			b.dx = -b.dx
			b.ladder = -1
			return
		}
	}
}

func (s *stage) rollBarrel(b *barrel, first, last int) {
	b.cadence++
	if b.cadence <= 1 {
		return
	}
	b.cadence = 0
	b.sub = (b.sub + 1) % 3
	if b.sub == 0 {
		b.frame += b.dx
		if b.frame < 0 {
			b.frame = 3
		} else if b.frame > 3 {
			b.frame = 0
		}
	}
	b.x += 3 * b.dx
	if b.oil && b.x < s.oil.x+6 && b.x > s.oil.x+2 && b.y > s.oil.y && b.y < s.oil.y+30 {
		s.oil.burning = true
		s.spawnFireball(s.oil.x, s.oil.y, -b.dx)
		if s.skill == hardSkill || s.level >= 5 {
			s.spawnFireball(s.oil.x, s.oil.y, b.dx)
		}
		b.active = false
	}
	// The slot keeps updating for the rest of this call after removal.
	if b.x < -10 || b.x > 260 || b.y < 0 || b.y > 194 {
		b.active = false
	}
	supported := false
	for _, g := range s.girders[first : last+1] {
		if b.x+10 >= g.left+2 && b.x+2 < g.left+20 {
			b.y = g.y - 11
			supported = true
			break
		}
	}
	if !supported && b.x >= 0 && b.x < 250 {
		b.falling = true
	}
	if b.bounceMax != 0 && b.frame == 0 {
		b.bounce += b.bounceDir
		if b.bounce > b.bounceMax {
			b.bounce, b.bounceDir = b.bounceMax, -1
		} else if b.bounce == 0 {
			// The reference leaves the height unchanged when it bounces again.
			if s.rng.bucket(4) != 0 {
				b.bounceMax = 0
			}
			b.bounceDir = 1
		}
	}
	if b.falling || b.y >= s.player.Y {
		return
	}
	// The reference scan stops short of the top ladder, which no rolling
	// barrel can reach.
	for j, l := range s.ladders[:len(s.ladders)-1] {
		if b.x+8 > l.x+1 && b.x+8 < l.x+10 && b.y == l.top-17 {
			if s.rng.bucket(max(6, 10-s.level)) == 0 {
				b.x, b.ladder = l.x-4, j
				b.bounceMax, b.bounce = 0, 0
			}
			return
		}
	}
}
