package classic

// playerSpan returns the player's horizontal collision interval for the
// current facing (DESIGN section 4, Contact and jump award boundaries).
func playerSpan(m *Movement) (left, right int) {
	if m.Facing > 0 {
		return m.X + 2, m.X + 13
	}
	return m.X + 4, m.X + 15
}

// contact reports an ordinary fatal contact. Holding a hammer on Easy skips
// every ordinary contact, including the drum.
func (s *stage) contact() bool {
	m := s.player
	if s.family == elevatorsFamily && s.elevatorsKill() {
		return true
	}
	if s.hammer.held && s.skill == easySkill {
		return false
	}
	bottom := 14
	if m.ladder != nil {
		bottom = 12
	}
	left, right := playerSpan(m)
	top := m.Y // ordinary pose descriptors have a zero collision offset
	o := s.oil
	if o.present && left <= o.x+15 && right >= o.x+2 && m.Y+16 <= o.y+11 && m.Y+15 >= o.y+8 {
		return true
	}
	for _, b := range s.barrels {
		if !b.active {
			continue
		}
		width := 12
		if b.kind != regularBarrel {
			width = 16
		}
		if b.x+2 <= right && b.x+width > left && b.y+1 <= m.Y+bottom && top < b.y+10 {
			return true
		}
	}
	for _, f := range s.fireballs {
		if f.active && f.x+3 <= right && f.x+12 >= left && f.y+2 <= m.Y+bottom-3 && top < f.y+14 {
			return true
		}
	}
	for _, p := range s.springs {
		if p.active && p.x+2 <= right && p.x+11 >= left && p.y+2 <= m.Y+14 && p.y+13 >= m.Y {
			return true
		}
	}
	// The thrower, on every family (DESIGN section 4, Thrower contact).
	t := s.thrower
	return right >= t.x+2 && left <= t.x+37 && m.Y+14 >= t.y && t.y+31 >= m.Y
}

// scanJump records nearby hazards at the top of a jump (DESIGN section 4,
// jump awards). It runs once, while the count is empty.
func (s *stage) scanJump() {
	m := s.player
	if m.jumpCount != 0 {
		return
	}
	for _, b := range s.barrels {
		if b.active && b.x-8 <= m.X+16 && b.x+24 >= m.X && b.y <= m.Y+30 && b.y+10 >= m.Y+14 {
			m.jumpCount++
		}
	}
	for _, p := range s.springs {
		if p.active && p.x-6 < m.X+16 && p.x+22 >= m.X && p.y >= m.Y+16 && p.y < m.Y+30 {
			m.jumpCount++
		}
	}
	for _, f := range s.fireballs {
		if f.active && f.x-8 < m.X+16 && f.x+24 >= m.X && f.y >= m.Y+12 && f.y < m.Y+30 {
			m.jumpCount += s.rng.bucket(3) + 1
		}
	}
}

// jumpAward converts a landed jump count into points.
func jumpAward(count int) int {
	switch {
	case count == 1:
		return 100
	case count == 2:
		return 300
	case count > 2:
		return 800
	}
	return 0
}
