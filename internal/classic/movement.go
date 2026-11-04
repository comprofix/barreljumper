package classic

const playerFoot = 16

// Controls carries one sample of the prototype's movement actions.
type Controls struct {
	Horizontal, Vertical int
	Jump                 bool
}

// Movement is the headless Barrels movement prototype. Coordinates are in
// the original logical grid. Advance receives input at the skeleton's 60 Hz.
type Movement struct {
	X, Y                                int
	Dead                                bool
	surfaces                            []girder
	grounded                            bool
	ladder                              *ladder
	climbPhase, walkPhase               int
	jump                                jumpControl
	jumping, descending                 bool
	jumpOrigin, jumpDirection, jumpTick int
	fall                                walkingFall
	clockPhase                          int
	counter                             uint16
}

func NewMovement() *Movement {
	return &Movement{
		X: 40, Y: 177, grounded: true, surfaces: barrelsGirders(),
		jump: jumpControl{armed: true},
	}
}

// Advance samples the nominal PIT clock without tying simulation to Draw.
func (m *Movement) Advance(c Controls, arcade bool) {
	const pitFrequency = 1193182
	const divisorPerFrame = 16384 * 60
	m.clockPhase += pitFrequency
	for m.clockPhase >= divisorPerFrame {
		m.clockPhase -= divisorPerFrame
		m.counter++
		if autoUpdateEligible(m.counter) {
			m.step(c, arcade)
		}
	}
}

func (m *Movement) step(c Controls, arcade bool) {
	if m.Dead {
		return
	}
	m.walkPhase = (m.walkPhase + 1) % 6
	if !c.Jump {
		m.jump.takeoff(false, false)
	}
	if m.ladder != nil {
		m.climb(c.Vertical)
		return
	}
	if m.jumping {
		m.stepJump()
		return
	}
	if c.Vertical != 0 {
		for i := range barrelsLadders {
			l := &barrelsLadders[i]
			if x, ok := l.attach(m.X, m.Y+playerFoot, c.Vertical < 0); ok {
				m.X, m.ladder, m.climbPhase = x, l, 0
				m.grounded = false
				m.fall = walkingFall{}
				m.climb(c.Vertical)
				return
			}
		}
	}
	if m.jump.takeoff(c.Jump, m.grounded) {
		m.jumping, m.descending = true, false
		m.jumpOrigin, m.jumpDirection, m.jumpTick = m.Y, c.Horizontal, 0
		m.grounded = false
		m.stepJump()
		return
	}
	if m.grounded {
		m.X = groundX(m.X, c.Horizontal, !arcade && m.walkPhase == 0, false)
		// Check the adjacent step at the active foot and one pixel below
		// its settled position, then continue falling if support ends.
		for _, g := range m.surfaces {
			if g.supports(m.X, m.Y+playerFoot+1) || g.supports(m.X, m.Y+playerFoot+2) {
				m.Y = g.y - playerFoot - 1
				return
			}
		}
		m.grounded = false
	}
	for _, g := range m.surfaces {
		if y, hit, fatal := m.fall.landOn(g, m.X, m.Y, playerFoot); hit {
			m.Y, m.grounded, m.Dead = y, true, fatal
			return
		}
	}
	m.Y = m.fall.advance(m.Y)
}

func (m *Movement) climb(direction int) {
	if direction == 0 {
		return
	}
	m.climbPhase += 2
	if m.climbPhase < 5 {
		return
	}
	m.climbPhase -= 5
	foot, attached := m.ladder.climbLimit(m.Y + playerFoot + direction)
	m.Y = foot - playerFoot
	if !attached {
		m.ladder = nil
	}
}

func (m *Movement) stepJump() {
	if m.descending {
		for _, g := range m.surfaces {
			if y, ok := g.landingY(m.X, m.Y+1, playerFoot); ok {
				m.Y, m.jumping, m.grounded = y, false, true
				return
			}
		}
	}
	m.jumpTick++
	dy := m.Y - m.jumpOrigin
	if dy%2 != 0 || dy > 7 {
		x := m.X + m.jumpDirection
		m.X = min(max(x, 0), maxPlayerX)
		if x != m.X {
			if -dy < 15 {
				m.descending = !m.descending
				m.jumpDirection = -m.jumpDirection
			} else {
				m.jumpDirection = 0
			}
		}
	}
	if dy > -10 || m.jumpTick%2 == 0 {
		if m.descending {
			m.Y++
		} else {
			m.Y--
			m.descending = m.Y-m.jumpOrigin == -16
		}
	}
}
