package classic

// fireball is one of the eight fireball slots (DESIGN section 4, Fireballs).
// Its origin is the top-left of a 16-pixel-tall body.
type fireball struct {
	active        bool
	x, y, dx, dy  int
	want          int // ladder intent: 1 down, -1 up, 0 none
	emerging      bool
	arc           int
	delay         int
	cadence, anim int
	reconsider    int
}

// oilDrum is the Barrels oil drum. Contact kills whether or not it burns.
type oilDrum struct {
	x, y    int
	burning bool
}

func (s *barrelsStage) fireballLimit() int {
	return min(8, s.skill+(s.level+1)/2+3)
}

// spawnFireball starts a fireball. A nonzero dir makes it emerge from the
// drum in that direction after a delay.
func (s *barrelsStage) spawnFireball(x, y, dir int) {
	limit := s.fireballLimit()
	slot := -1
	for i := 0; i < limit; i++ {
		if !s.fireballs[i].active {
			slot = i
			break
		}
	}
	if slot < 0 {
		return
	}
	// The reference draws an animation frame and phase here, then replaces
	// the phase; both draws stay in the stream.
	s.rng.bucket(3)
	s.rng.bucket(4)
	f := fireball{active: true, x: x + dir, y: y}
	if dir != 0 {
		f.emerging, f.dx, f.dy, f.delay = true, dir, -1, 100
	} else if s.player.X < 160 {
		f.dx = 1
	} else {
		f.dx = -1
	}
	s.fireballs[slot] = f
}

func (s *barrelsStage) moveFireballs() {
	period := max(120, 370-s.level)
	for i := range s.fireballs {
		f := &s.fireballs[i]
		if !f.active {
			continue
		}
		if f.delay > 0 {
			f.delay--
			continue
		}
		if f.emerging {
			s.emergeFireball(f)
			continue
		}
		f.cadence++
		if f.cadence <= 1 && f.dy <= 0 {
			continue
		}
		f.cadence = 0
		f.anim = (f.anim + 1) % 4
		if f.dy == 0 && f.want != 0 {
			s.engageFireballLadder(f)
		}
		if f.y < 19 && f.dy != 0 {
			f.dy = 1
		}
		if f.dy != 0 {
			if f.anim%2 == 0 {
				f.y += f.dy
			}
		} else {
			s.walkFireball(f, period)
		}
		s.supportFireball(f)
	}
}

func (s *barrelsStage) emergeFireball(f *fireball) {
	move := f.dy < 0
	if f.dy > 0 {
		f.cadence++
		move = f.cadence > 1
	}
	if move {
		f.cadence = 0
		f.x += f.dx
		f.y += f.dy
		f.arc += f.dy
		f.anim = (f.anim + 1) % 2
	}
	if f.arc < -6 {
		f.dy, f.cadence = 1, 1
	}
	if f.arc == 0 {
		f.emerging = false
	}
}

func (s *barrelsStage) engageFireballLadder(f *fireball) {
	for _, l := range barrelsLadders {
		bottom := l.top + l.height
		if f.x+8 <= l.x+2 || f.x+8 >= l.x+8 || f.y+16 < l.top-6 || bottom < f.y+16 {
			continue
		}
		if f.y+16 < l.top+l.height/2 {
			if f.want == 1 {
				f.dy = 1
			}
		} else if f.want == -1 {
			f.dy = -1
		} else {
			f.dy, f.want = 0, 0
		}
		if f.dy != 0 {
			f.dx, f.x, f.cadence = 0, l.x-2, 0
			f.y += f.want
		}
		return
	}
}

func (s *barrelsStage) walkFireball(f *fireball, period int) {
	f.x += f.dx
	f.reconsider++
	switch {
	case f.x < 0:
		f.x, f.dx = 0, 1
	case f.x > 248:
		f.x, f.dx = 248, -1
	case f.reconsider > period:
		f.reconsider = 0
		f.dx = -1
		if s.rng.bucket(2) != 0 {
			f.dx = 1
		}
		if s.rng.bucket(5) < 3 {
			switch {
			case f.y < s.player.Y-4:
				f.want = 1
			case f.y > s.player.Y+12:
				f.want = -1
			default:
				f.want = 0
			}
		} else if s.rng.bucket(s.level+6) == 0 {
			f.want = 0
		} else if s.rng.bucket(2) != 0 {
			f.want = 1
		} else {
			f.want = -1
		}
	}
}

// supportFireball keeps a fireball on its girder and ends ladder travel.
// Without support it turns back.
func (s *barrelsStage) supportFireball(f *fireball) {
	if f.y <= 0 || f.y >= 184 {
		f.dy = -f.dy
	}
	if f.dy == 0 {
		f.y++
	}
	for _, g := range s.girders {
		if f.x+13 <= g.left || f.x+5 > g.right {
			continue
		}
		foot := f.y + 16
		if f.dy <= 0 {
			if foot < g.y || foot > g.y+2 {
				continue
			}
		} else if foot < g.y-2 || foot > g.y {
			continue
		}
		f.y = g.y - 16
		if f.dy != 0 {
			f.dy, f.want, f.reconsider = 0, 0, 100
			f.dx = 1
			if s.rng.bucket(2) != 0 {
				f.dx = -1
			}
		}
		return
	}
	f.dx = -f.dx
}
