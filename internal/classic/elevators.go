package classic

// Elevators machinery (DESIGN section 4, M3 Elevators).

// elevatorTop and elevatorBottom bound both shafts' travel.
const elevatorTop, elevatorBottom = 50, 194

// elevator is one moving cell; its support is one of the last six girders.
type elevator struct {
	x, y    int
	dir     int // -1 rising, 1 descending
	cadence int
}

// spring is one of the four spring slots.
type spring struct {
	active      bool
	compressing bool
	dx, dy      int // horizontal step on or off, and vertical direction
	cadence     int
	squash      int
	x, y        int
	drop        int
	delay       int
}

func elevatorsLayout() layout {
	var g []girder
	for x := 0; x <= 240; x += 20 {
		g = append(g, girder{x, x + 20, 194})
	}
	for _, c := range [][3]int{
		{0, 174, 30}, {0, 138, 30}, {0, 90, 30}, {80, 160, 30}, {70, 90, 10}, {80, 90, 30},
		{150, 176, 30}, {190, 170, 10}, {210, 164, 10}, {230, 158, 30}, {240, 132, 20}, {220, 126, 10},
		{200, 120, 10}, {170, 114, 20}, {160, 84, 30}, {200, 76, 10}, {220, 70, 10}, {240, 64, 20},
		{220, 98, 20}, {240, 98, 20},
	} {
		g = append(g, girder{c[0], c[0] + c[2], c[1]})
	}
	for x := 0; x <= 160; x += 20 {
		g = append(g, girder{x, x + 20, 50})
	}
	g = append(g, girder{180, 210, 50}, girder{110, 130, 24}, girder{130, 150, 24}, girder{150, 160, 24})
	for _, e := range elevatorStarts {
		g = append(g, girder{e.x, e.x + 20, e.y})
	}
	return layout{
		family:  elevatorsFamily,
		girders: g,
		ladders: []ladder{
			{10, 144, 30, fullLadder}, {20, 96, 42, fullLadder}, {80, 96, 64, fullLadder}, {100, 96, 64, fullLadder},
			{240, 138, 20, fullLadder}, {220, 104, 22, fullLadder}, {170, 90, 24, fullLadder}, {240, 70, 28, fullLadder},
			{200, 56, 20, fullLadder}, {150, 30, 20, fullLadder},
		},
		reserved: []ladder{{80, 2, 24, fullLadder}, {100, 2, 24, fullLadder}, {80, 26, 24, fullLadder}, {100, 26, 24, fullLadder}},
		player:   point{8, 158},
		thrower:  point{30, 16},
		items:    []point{{9, 80}, {89, 150}, {246, 54}},
	}
}

// elevatorStarts are the cells in record order: the rising left shaft,
// then the descending right one.
var elevatorStarts = [6]elevator{
	{40, 194, -1, 0}, {40, 150, -1, 0}, {40, 108, -1, 0},
	{120, 56, 1, 0}, {120, 150, 1, 0}, {120, 108, 1, 0},
}

// setupElevators enables the springs and spawns the starting fireballs.
func (s *stage) setupElevators() {
	s.elevators = elevatorStarts
	n := min(4, s.level+s.skill-1)
	for i := 0; i < n; i++ {
		delay := 55 * (i + 1)
		if n < 4 {
			delay = (140-25*n)*(i+1) + 20
		}
		s.springs[i] = spring{active: true, dx: 1, dy: 1, x: -24, y: 19, drop: 228 - 6*i, delay: delay}
	}
	s.spawnFireball(80, 74, 0)
	s.spawnFireball(230, 81, 0)
	if s.level > 4 {
		s.spawnFireball(10, 74, 0)
	}
}

func (s *stage) moveElevators() {
	base := len(s.girders) - len(s.elevators)
	for i := range s.elevators {
		e := &s.elevators[i]
		e.cadence++
		if e.cadence <= 2 {
			continue
		}
		e.cadence = 0
		e.y += e.dir
		if e.dir < 0 && e.y < elevatorTop-6 {
			e.y = elevatorBottom
		} else if e.dir > 0 && e.y > elevatorBottom+6 {
			e.y = elevatorTop
		}
		s.girders[base+i].y = e.y
	}
}

func (s *stage) moveSprings() {
	for i := range s.springs {
		p := &s.springs[i]
		if p.x >= p.drop {
			p.cadence++
		}
		if !p.active {
			continue
		}
		if p.delay > 0 {
			p.delay--
			continue
		}
		p.cadence++
		if p.cadence <= 1 {
			continue
		}
		p.cadence = 0
		p.x += 3 * p.dx
		if s.level <= 2 {
			p.x += p.dx
		}
		p.y += 3 * p.dy
		if p.x >= p.drop && p.dx != 0 {
			p.dx = 0
		}
		if p.y < 7 {
			p.y, p.dy = 7, 1
		}
		if p.compressing {
			p.squash++
			if p.squash > 7 {
				p.compressing = false
			}
			if p.y >= 34 && p.dy > 0 {
				p.y, p.dy = 34, -1
			}
		}
		if p.dx != 0 && p.x < 190 {
			if p.y >= 22 && !p.compressing && p.dy > 0 {
				p.squash, p.compressing = 0, true
			}
			continue
		}
		if p.y > 200 {
			p.x, p.y = -24-s.rng.bucket(3), 19
			p.dx, p.dy = 1, 1
			p.drop -= 6
			if p.drop < 210 {
				p.drop = 228
			}
			p.delay = 20
			if s.level > 3 {
				p.delay += 20
			}
		}
	}
}

// elevatorsKill reports the Elevators environment deaths, which no bypass
// prevents.
func (s *stage) elevatorsKill() bool {
	m := s.player
	feet := m.Y + 16
	switch {
	case m.X < 136 && m.Y >= 50 && m.Y <= 56:
		return true
	case feet >= 194:
		return true
	case feet >= 182:
		return (m.X+13 >= 40 && m.X+2 < 60) || (m.X+13 >= 120 && m.X+2 < 140)
	}
	return false
}
