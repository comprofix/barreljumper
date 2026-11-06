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

// Movement is the Barrels player. Coordinates are in the original logical
// grid. The stage calls step once per eligible reference update.
type Movement struct {
	X, Y                            int
	Dead                            bool
	Facing                          int  // 1 right, -1 left
	Hammer                          bool // blocks takeoff and ladders; narrows walking
	surfaces                        []girder
	ladders                         []ladder
	grounded                        bool
	ladder                          *ladder
	movePhase                       int
	jump                            jumpControl
	jumping, falling                bool
	jumpDisplacement, jumpDirection int
	jumpStep                        int
	fall                            walkingFall
	// Jump award hooks for the stage: apex is set on the update that turns
	// the jump down, landed on a normal landing. jumpCount is the award count.
	apex, landed bool
	jumpCount    int
}

// NewMovement returns a player at the Barrels start.
func NewMovement() *Movement {
	l := barrelsLayout()
	return newMovement(&l)
}

func newMovement(l *layout) *Movement {
	return &Movement{
		X: l.player.x, Y: l.player.y, Facing: 1, grounded: true,
		surfaces: l.girders, ladders: l.ladders,
	}
}

// takeoffY is the height the current jump or walking fall started from.
func (m *Movement) takeoffY() int {
	return m.Y - m.jumpDisplacement - m.fall.distance
}

func (m *Movement) step(c Controls, arcade bool) {
	m.apex, m.landed = false, false
	if m.Dead {
		return
	}
	// DESIGN's M1 comparison establishes takeoff before ladder attachment.
	if m.jump.takeoff(c.Jump, m.grounded && !m.jumping && m.ladder == nil && !m.Hammer) {
		m.jumping, m.grounded = true, false
		m.jumpDisplacement, m.jumpDirection, m.jumpStep = 0, c.Horizontal, -1
		m.movePhase, m.jumpCount = 0, 0
		m.fall = walkingFall{}
	}
	if !m.jumping && m.ladder == nil && c.Vertical != 0 && !m.Hammer {
		for i := range m.ladders {
			l := &m.ladders[i]
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
		m.Facing = c.Horizontal
		m.X = groundX(m.X, c.Horizontal, arcade && m.movePhase == 0, m.Hammer)
		m.movePhase = (m.movePhase + 1) % 6
	}
	// The wall clamp runs on every call that gets here, idle and falling
	// ones too, so a belt cannot carry the player through a wall (DESIGN
	// section 4, M3 Conveyor).
	if m.Hammer && !m.falling {
		m.X = min(max(m.X, 9), 236)
	} else {
		m.X = min(max(m.X, 0), maxPlayerX)
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
			m.apex = true
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
			m.Facing = m.jumpDirection
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
				m.jumping, m.landed = false, true
			}
			m.jumpDisplacement = 0
			return
		}
	}
	m.grounded, m.falling = false, true
}
