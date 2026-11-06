package classic

// Classic skill words (DESIGN section 4, Reference build).
const (
	easySkill = iota
	arcadeSkill
	hardSkill
)

// barrelsGoal is the platform the thrower stands beside. It is outside the
// player's support scan; reaching its height completes the stage.
var barrelsGoal = girder{110, 130, 24}

// Stage outcomes reported by tick.
const (
	stagePlaying = iota
	stageDied
	stageCleared
)

// tauntPeriod is how often, in shared updates, the thrower taunts on every
// family except Barrels.
const tauntPeriod = 200

// popup is a score shown briefly where it was earned. It is drawn only.
type popup struct {
	points, x, y, age int
}

// lostItem is a collectable on the screens that have them.
type lostItem struct {
	active bool
	x, y   int
	value  int
}

// rivet is one of the eight removable Rivets supports. Its support is the
// matching girder among the last eight records.
type rivet struct {
	x, y    int
	removed bool
	facing  int // the player's facing when it was removed
}

// stage holds the rules of one Classic screen attempt. It knows nothing
// of drawing, sound or the frame rate.
type stage struct {
	rng        *random
	family     int
	skill      int
	level      int // zero-based
	layout     layout
	player     *Movement
	girders    []girder
	ladders    []ladder
	barrels    [16]barrel
	thrower    thrower
	oil        oilDrum
	fireballs  [8]fireball
	pickups    []hammerPickup
	hammer     heldHammer
	smash      smash
	items      []lostItem
	rivets     []rivet
	elevators  [6]elevator
	springs    [4]spring
	belts      [4]belt
	retractors [2]retractor
	pies       [10]pie
	bonus      int
	// updates is the shared eligible-update counter. Stage entry keeps it,
	// so the first bonus decrease after a restart can come early, and the
	// periodic taunts and spawns keep their phase across screens.
	updates *int
	earned  int // points earned this attempt, collected by the session
	popups  []popup
	sounds  sound // events since the session last took them
}

func stageBonus(level int) int {
	return min(5000+1000*level, 8000)
}

// itemValue is the lost item award for a zero-based level.
func itemValue(level int) int {
	return [...]int{300, 500, 800}[min(level, 2)]
}

func newStage(family int, rng *random, skill, level int, updates *int) *stage {
	l := newLayout(family)
	s := &stage{
		rng: rng, family: family, skill: skill, level: level, updates: updates,
		layout:  l,
		player:  newMovement(&l),
		girders: l.girders,
		ladders: l.ladders,
		oil:     l.oil,
		bonus:   stageBonus(level),
	}
	for _, h := range l.hammers {
		s.pickups = append(s.pickups, hammerPickup{true, h.x, h.y})
	}
	for _, it := range l.items {
		s.items = append(s.items, lostItem{true, it.x, it.y, itemValue(level)})
	}
	switch family {
	case barrelsFamily:
		s.thrower = newThrower(l.thrower, level, skill)
	case rivetsFamily:
		s.thrower = thrower{x: l.thrower.x, y: l.thrower.y, phase: 2, threshold: 32}
		for _, x := range []int{60, 190} {
			for _, y := range rivetRows {
				s.rivets = append(s.rivets, rivet{x: x, y: y})
			}
		}
	case elevatorsFamily:
		s.thrower = thrower{x: l.thrower.x, y: l.thrower.y, phase: 2, threshold: 32}
		s.setupElevators()
	case conveyorFamily:
		s.thrower = thrower{x: l.thrower.x, y: l.thrower.y, phase: 2, threshold: 32}
		s.setupConveyor()
	}
	return s
}

// update runs one reference counter tick in the gameplay loop order of
// DESIGN section 4. eligible is false on the ticks AUTO skips.
func (s *stage) update(c Controls, eligible bool) int {
	s.agePopups()
	if eligible {
		*s.updates++
		if *s.updates%120 == 0 {
			s.bonus -= 100
			if s.bonus <= 0 {
				s.bonus = 0
				return stageDied
			}
		}
		// The low-bonus warning (DESIGN section 4, Clock and player motion).
		if s.bonus < lowBonus && *s.updates%lowBonusEvery == 0 {
			s.sounds |= soundWarn
		}
		if s.smash.active {
			s.updateSmash()
		} else {
			m := s.player
			x, y, jumping := m.X, m.Y, m.jumping
			m.step(c, s.skill != easySkill)
			switch {
			case m.jumping && !jumping:
				s.sounds |= soundJump
			case jumping && !m.jumping && !m.Dead:
				s.sounds |= soundLand
			case m.ladder != nil && footfall(y) != footfall(m.Y):
				s.sounds |= alternate(m.Y, soundClimb, soundClimbB)
			case !m.jumping && m.ladder == nil && footfall(x) != footfall(m.X):
				s.sounds |= alternate(m.X, soundStep, soundStepB)
			}
			if s.player.apex {
				s.scanJump()
			}
			if s.player.landed {
				s.award(jumpAward(s.player.jumpCount), s.player.X, s.player.Y+10)
			}
			if s.player.Dead {
				return stageDied
			}
			s.collectHammer()
			s.moveFireballs()
			s.stageRoutines()
		}
		if s.hammer.held {
			s.swingHammer()
		} else {
			s.collectItems()
		}
	}
	if s.completed() {
		return stageCleared
	}
	if s.hammer.held {
		s.hitWithHammer()
	}
	if s.contact() {
		return stageDied
	}
	return stagePlaying
}

// stageRoutines runs the family's own machinery, then the periodic taunt
// and fireball spawns of the other families.
func (s *stage) stageRoutines() {
	switch s.family {
	case barrelsFamily:
		s.release()
		s.moveBarrels()
		return
	case rivetsFamily:
		s.removeRivets()
	case elevatorsFamily:
		s.moveSprings()
		s.moveElevators()
	case conveyorFamily:
		s.moveBelts()
		s.moveRetractors()
	}
	if s.thrower.taunting {
		s.taunt()
	} else if *s.updates%tauntPeriod == 0 {
		t := &s.thrower
		t.counter, t.taunting, t.phase, t.pose = 0, true, 1, throwerTauntA
	}
	if *s.updates%s.fireballPeriod() == 0 && s.family == conveyorFamily {
		s.spawnDrumFireball()
	}
	if *s.updates%s.fireballPeriod() == 0 && s.family == rivetsFamily {
		i := s.rng.bucket(4)
		y := 176 - 32*i
		if s.player.X < 130 {
			s.spawnFireball(248-10*i, y, 0)
		} else {
			s.spawnFireball(2+10*i, y, 0)
		}
	}
	if s.family == conveyorFamily && *s.updates%s.piePeriod() == 0 {
		s.spawnPie()
	}
}

// fireballPeriod is the spawn interval on Conveyor and Rivets.
func (s *stage) fireballPeriod() int {
	return max(100, 240-24*(s.level+s.skill))
}

func (s *stage) completed() bool {
	m := s.player
	switch s.family {
	case rivetsFamily:
		for _, r := range s.rivets {
			if !r.removed {
				return false
			}
		}
		return true
	case elevatorsFamily:
		return 24-supportFoot-1 >= m.Y
	case conveyorFamily:
		return 42-supportFoot-1 >= m.Y
	default:
		if !m.jumping && barrelsGoal.y-supportFoot >= m.Y {
			m.Y--
			s.earned += s.bonus
			return true
		}
		return false
	}
}

// removeRivets removes at most one rivet, or retires one removed rivet's
// support, per update (DESIGN section 4, Rivet removal).
func (s *stage) removeRivets() {
	m := s.player
	base := len(s.girders) - len(s.rivets)
	for i := range s.rivets {
		r := &s.rivets[i]
		if !r.removed {
			if r.y-16 == m.takeoffY() && r.x+1 <= m.X+8 && r.x+8 >= m.X+8 {
				r.removed, r.facing = true, m.Facing
				s.award(100, r.x, r.y-10)
				return
			}
			continue
		}
		g := &s.girders[base+i]
		if g.right == g.left {
			continue
		}
		if r.x > m.X+12 || r.x+10 < m.X+6 || r.facing != m.Facing {
			g.right = g.left
			return
		}
	}
}

// collectItems collects at most one lost item while the player is neither
// jumping nor climbing.
func (s *stage) collectItems() {
	m := s.player
	if m.jumping || m.ladder != nil {
		return
	}
	for i := range s.items {
		it := &s.items[i]
		if it.active && it.x+1 <= m.X+15 && it.x+12 > m.X+2 && it.y < m.Y+14 && it.y+10 > m.Y+2 {
			it.active = false
			s.award(it.value, it.x, it.y+2)
			return
		}
	}
}

func (s *stage) award(points, x, y int) {
	if points == 0 {
		return
	}
	s.earned += points
	s.sounds |= soundPoints
	s.popups = append(s.popups, popup{points: points, x: x, y: y})
}

func (s *stage) agePopups() {
	kept := s.popups[:0]
	for _, p := range s.popups {
		if p.age++; p.age < popupTicks {
			kept = append(kept, p)
		}
	}
	s.popups = kept
}

// Below lowBonus the warning sounds every lowBonusEvery eligible updates.
const (
	lowBonus      = 1000
	lowBonusEvery = 34
)

// popupTicks is how long a score stays on screen, in counter ticks.
const popupTicks = 60
