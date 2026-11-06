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

// popup is a score shown briefly where it was earned. It is drawn only.
type popup struct {
	points, x, y, age int
}

// barrelsStage holds the rules of one Classic Barrels attempt. It knows
// nothing of drawing, sound or the frame rate.
type barrelsStage struct {
	rng       *random
	skill     int
	level     int // zero-based
	player    *Movement
	girders   []girder
	barrels   [16]barrel
	thrower   thrower
	oil       oilDrum
	fireballs [8]fireball
	pickups   [2]hammerPickup
	hammer    heldHammer
	smash     smash
	bonus     int
	// updates is the shared eligible-update counter. Stage entry keeps it,
	// so the first bonus decrease after a restart can come early.
	updates *int
	earned  int // points earned this attempt, collected by the session
	popups  []popup
}

func stageBonus(level int) int {
	return min(5000+1000*level, 8000)
}

func newBarrelsStage(rng *random, skill, level int, updates *int) *barrelsStage {
	s := &barrelsStage{
		rng: rng, skill: skill, level: level, updates: updates,
		player:  NewMovement(),
		girders: barrelsGirders(),
		thrower: newThrower(level, skill),
		oil:     oilDrum{x: 18, y: 164},
		pickups: [2]hammerPickup{{true, 230, 152}, {true, 18, 47}},
		bonus:   stageBonus(level),
	}
	return s
}

// update runs one reference counter tick in the gameplay loop order of
// DESIGN section 4. eligible is false on the ticks AUTO skips.
func (s *barrelsStage) update(c Controls, eligible bool) int {
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
		if s.smash.active {
			s.updateSmash()
		} else {
			s.player.step(c, s.skill != easySkill)
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
			s.release()
			s.moveBarrels()
		}
		if s.hammer.held {
			s.swingHammer()
		}
	}
	if !s.player.jumping && barrelsGoal.y-supportFoot >= s.player.Y {
		s.player.Y--
		s.earned += s.bonus
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

func (s *barrelsStage) award(points, x, y int) {
	if points == 0 {
		return
	}
	s.earned += points
	s.popups = append(s.popups, popup{points: points, x: x, y: y})
}

func (s *barrelsStage) agePopups() {
	kept := s.popups[:0]
	for _, p := range s.popups {
		if p.age++; p.age < popupTicks {
			kept = append(kept, p)
		}
	}
	s.popups = kept
}

// popupTicks is how long a score stays on screen, in counter ticks.
const popupTicks = 60
