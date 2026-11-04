package classic

const (
	supportFoot = 15
	ladderFoot  = 16
)

// Controls carries one sample of the prototype's movement actions.
type Controls struct {
	Horizontal, Vertical int
	Jump                 bool
}

// Movement is the headless Barrels movement prototype. Coordinates are in
// the original logical grid. Advance receives input at the skeleton's 60 Hz.
type Movement struct {
	X, Y                            int
	Dead                            bool
	surfaces                        []girder
	grounded                        bool
	ladder                          *ladder
	movePhase                       int
	jump                            jumpControl
	jumping, falling                bool
	jumpDisplacement, jumpDirection int
	jumpStep                        int
	fall                            walkingFall
	clockPhase                      int
	counter                         uint16
}

func NewMovement() *Movement {
	return &Movement{
		X: 40, Y: 176, grounded: true, surfaces: barrelsGirders(),
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
	// DESIGN's M1 comparison establishes takeoff before ladder attachment.
	if m.jump.takeoff(c.Jump, m.grounded && !m.jumping && m.ladder == nil) {
		m.jumping, m.grounded = true, false
		m.jumpDisplacement, m.jumpDirection, m.jumpStep = 0, c.Horizontal, -1
		m.movePhase = 0
		m.fall = walkingFall{}
	}
	if !m.jumping && m.ladder == nil && c.Vertical != 0 {
		for i := range barrelsLadders {
			l := &barrelsLadders[i]
			if x, ok := l.attach(m.X, m.Y+ladderFoot, c.Vertical < 0); ok {
				m.X, m.ladder, m.movePhase = x, l, 0
				m.grounded = false
				break
			}
		}
	}
	if m.ladder != nil {
		m.climb(c.Vertical)
		if m.ladder != nil {
			return
		}
	} else if m.jumping {
		m.stepJump()
		return
	} else if !m.falling && c.Horizontal != 0 {
		m.X = groundX(m.X, c.Horizontal, arcade && m.movePhase == 0, false)
		m.movePhase = (m.movePhase + 1) % 6
	}
	// Walking descent precedes contact, including the landing call. A lower
	// girder step can therefore suppress walking for one recovery call.
	m.Y = m.fall.advance(m.Y)
	for _, g := range m.surfaces {
		if y, hit, fatal := m.fall.landOn(g, m.X, m.Y, supportFoot); hit {
			m.Y, m.grounded, m.Dead = y, true, fatal
			m.falling = false
			return
		}
	}
	m.grounded, m.falling = false, true
}

func (m *Movement) climb(direction int) {
	if direction == 0 {
		return
	}
	m.movePhase++
	if m.movePhase < 5 {
		return
	}
	m.movePhase = 0
	foot, attached := m.ladder.climbLimit(m.Y + ladderFoot + 2*direction)
	m.Y = foot - ladderFoot
	if !attached {
		m.ladder = nil
	}
}

func (m *Movement) stepJump() {
	displacement := m.jumpDisplacement
	if displacement%2 != 0 || displacement > 7 {
		m.X += m.jumpDirection
	}
	m.movePhase++
	if displacement > -10 {
		m.movePhase++
	}
	if m.movePhase >= 2 {
		m.movePhase = 0
		m.Y += m.jumpStep
		m.jumpDisplacement += m.jumpStep
		if m.jumpDisplacement <= -16 {
			m.jumpStep = 1
		}
		if m.jumpDisplacement == 14 && m.jumpStep > 0 {
			m.jumpStep = 2
		}
	}
	// The signed threshold is evaluated after vertical motion. It also
	// rebounds near the apex; it is not an absolute-height comparison.
	if m.X < 0 || m.X > maxPlayerX {
		m.X = min(max(m.X, 0), maxPlayerX)
		if m.jumpDisplacement < 15 {
			m.jumpStep = -m.jumpStep
			if m.X == 0 {
				m.jumpDirection = 1
			} else {
				m.jumpDirection = -1
			}
		} else {
			m.jumpDirection = 0
		}
	}
	if m.jumpStep < 0 {
		return
	}
	for _, g := range m.surfaces {
		if y, hit := g.landingY(m.X, m.Y, supportFoot); hit {
			m.Y, m.grounded = y, true
			if m.falling {
				m.Dead = m.jumpDisplacement > 16
				m.falling = false
			} else {
				m.jumping = false
			}
			m.jumpDisplacement = 0
			return
		}
	}
	m.grounded, m.falling = false, true
}
