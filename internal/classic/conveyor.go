package classic

// Conveyor machinery (DESIGN section 4, M3 Conveyor).

// belt is one conveyor. Belts 0 and 1 reverse after threshold moves.
type belt struct {
	x, y, width int
	dir         int
	cadence     int8 // a signed byte: its wrap changes the move pattern
	counter     int
	threshold   int
}

// retractor drives one of the two retracting ladders.
type retractor struct {
	ladder    int // index into the stage's ladders
	extension int
	dir       int
	waiting   bool
	counter   int
}

// pie rides one belt.
type pie struct {
	active bool
	x, y   int
	belt   int
}

const (
	retractorBase = 48 // the ladder top at full extension
	pieLeft       = -20
	pieRight      = 266
)

var conveyorBelts = [4]belt{
	{0, 156, 260, 1, 0, 0, 360},
	{0, 42, 260, 1, 0, 120, 210},
	{0, 80, 120, 1, 0, 0, 200},
	{140, 80, 120, -1, 0, 0, 200},
}

func conveyorLayout() layout {
	var g []girder
	row := func(y int, xs ...int) {
		for _, x := range xs {
			g = append(g, girder{x, x + 20, y})
		}
	}
	full := []int{0, 20, 40, 60, 80, 100, 120, 140, 160, 180, 200, 220, 240}
	row(194, full...)
	row(156, full...)
	row(118, 10, 30, 70, 80, 100, 120, 140, 160, 170, 210, 230)
	row(80, 0, 20, 40, 60, 80, 100, 140, 160, 180, 200, 220, 240)
	row(42, full...)
	row(22, 110, 120, 140)
	return layout{
		family:  conveyorFamily,
		girders: g,
		ladders: []ladder{
			{20, 162, 32, fullLadder}, {90, 162, 32, fullLadder}, {160, 162, 32, fullLadder}, {230, 162, 32, fullLadder},
			{30, 86, 32, fullLadder}, {90, 86, 32, fullLadder}, {160, 86, 32, fullLadder}, {220, 86, 32, fullLadder},
			{70, 124, 32, fullLadder}, {180, 124, 32, fullLadder},
			{20, 56, 24, fullLadder}, {230, 56, 24, fullLadder},
		},
		reserved: []ladder{{80, 0, 42, fullLadder}, {100, 0, 42, fullLadder}, {}, {}},
		player:   point{42, 176},
		thrower:  point{125, 8},
		oil:      oilDrum{x: 122, y: 68, present: true, burning: true},
		hammers:  []point{{125, 128}, {10, 90}},
		items:    []point{{131, 184}, {69, 108}, {232, 108}},
	}
}

func (s *stage) setupConveyor() {
	s.belts = conveyorBelts
	s.retractors = [2]retractor{{ladder: 10, extension: 8, dir: 1}, {ladder: 11, extension: 8, dir: 1}}
}

func (s *stage) moveBelts() {
	m := s.player
	for i := range s.belts {
		b := &s.belts[i]
		b.cadence++
		if b.cadence%3 == 0 {
			continue
		}
		for j := range s.pies {
			p := &s.pies[j]
			if !p.active || p.belt != i {
				continue
			}
			p.x += b.dir
			switch i {
			case 0, 1:
				p.active = p.x >= pieLeft && p.x <= pieRight
			case 2:
				p.active = p.x < s.oil.x+3
			case 3:
				p.active = p.x >= s.oil.x+3
			}
		}
		foot := m.Y + supportFoot
		if !m.falling && m.ladder == nil && !m.jumping && b.x < m.X+14 && b.x+b.width >= m.X &&
			(b.y-2 == foot || b.y-1 == foot) {
			m.X += b.dir
		}
		if i == 1 {
			s.thrower.x += b.dir
		}
		if i < 2 {
			b.counter++
			if b.counter > b.threshold {
				b.counter, b.dir = 0, -b.dir
			}
		}
	}
}

func (s *stage) moveRetractors() {
	for i := range s.retractors {
		r := &s.retractors[i]
		l := &s.ladders[r.ladder]
		r.counter++
		if r.waiting {
			if r.counter > 280 {
				r.counter, r.waiting = 0, false
			}
		} else if r.counter > 8 {
			r.counter = 0
			if r.dir < 0 {
				r.counter = 4
			}
			r.extension += r.dir
			switch {
			case r.extension < 0:
				r.extension, r.dir, r.waiting = 0, 1, true
			case r.extension > 16:
				r.extension, r.dir, r.waiting, r.counter = 16, -1, true, 276
			}
			l.top, l.height = retractorBase+r.extension, 32-r.extension
		}
		l.kind = fullLadder
		if r.extension > 8 {
			l.kind = lowerSegment
		}
	}
}

// pieLimit is how many pie slots the spawn may use.
func (s *stage) pieLimit() int {
	return min(10, 5+s.level+s.skill)
}

// piePeriod is the pie spawn interval in shared updates.
func (s *stage) piePeriod() int {
	return max(80, 180-24*(s.level+s.skill))
}

func (s *stage) spawnPie() {
	slot := -1
	for i := 0; i < s.pieLimit(); i++ {
		if !s.pies[i].active {
			slot = i
			break
		}
	}
	if slot < 0 {
		return
	}
	p := &s.pies[slot]
	p.active = true
	if s.rng.bucket(2) == 0 {
		p.belt = 0
	} else {
		p.belt = s.rng.bucket(2) + 2
	}
	b := s.belts[p.belt]
	p.y = b.y - 7
	switch {
	case p.belt == 2:
		p.x = pieLeft
	case p.belt == 3 || b.dir < 0:
		p.x = pieRight
	default:
		p.x = pieLeft
	}
}

// spawnDrumFireball sends a fireball out of the burning drum.
func (s *stage) spawnDrumFireball() {
	dir := -1
	if s.rng.bucket(2) == 0 {
		dir = 1
	}
	s.spawnFireball(s.oil.x, s.oil.y-6, dir)
}
