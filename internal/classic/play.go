package classic

import (
	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/input"
	"github.com/comprofix/barreljumper/internal/scene"
)

// Play phases.
const (
	playing = iota
	dying
	clearing
	gameOver
)

// Presentation timing in reference counter ticks (about 72.8 per second).
// The death times follow the reference death sequence with sound muted.
const (
	deathFreezeTicks = 73  // 1.0 s still
	deathSpinTicks   = 140 // 20 frames of 7 ticks
	deathHoldTicks   = 146 // 2.0 s
	clearTicks       = 109 // 1.5 s
	gameOverTicks    = 218 // 3.0 s
)

// Play is one Classic game, following the Classic schedule (DESIGN
// section 4, Screen progression). Families not yet built in M3 are passed
// over until they arrive.
type Play struct {
	skill     config.Skill
	skillWord int
	rng       random
	score     int
	reserves  int
	level     int
	index     int // position in the level's schedule row
	nextLife  int
	updates   int
	stage     *stage
	phase     int
	timer     int
	// The nominal PIT clock, sampled each 60 Hz frame.
	clockPhase int
	counter    uint16
}

// NewPlay starts a game at skill, seeded from the scene's random source.
func NewPlay(ctx *scene.Context, skill config.Skill) *Play {
	seed := uint32(0)
	if ctx != nil && ctx.RNG != nil {
		seed = ctx.RNG.Uint32()
	}
	p := &Play{skill: skill, skillWord: skillWord(skill), rng: random{seed}, nextLife: 7000}
	p.reserves = 3
	if p.skillWord == easySkill {
		p.reserves = 5
	}
	if ctx != nil && ctx.StartLevel > 1 {
		p.level = ctx.StartLevel - 1
	}
	p.beginTurn()
	return p
}

func skillWord(skill config.Skill) int {
	switch skill.ID {
	case "easy":
		return easySkill
	case "hard":
		return hardSkill
	}
	return arcadeSkill
}

// beginTurn consumes a reserve and enters the current screen.
func (p *Play) beginTurn() {
	p.reserves--
	p.enterStage()
}

// builtFamilies are the families M3 has implemented so far.
var builtFamilies = map[int]bool{barrelsFamily: true, rivetsFamily: true}

func (p *Play) enterStage() {
	for !builtFamilies[p.family()] {
		p.nextStage()
	}
	p.stage = newStage(p.family(), &p.rng, p.skillWord, p.level, &p.updates)
	p.phase, p.timer = playing, 0
}

func (p *Play) family() int { return classicSchedule[p.level%5][p.index] }

// nextStage advances the schedule; the row's end starts the next level.
func (p *Play) nextStage() {
	p.index++
	if p.index >= len(classicSchedule[p.level%5]) {
		p.level++
		p.index = 0
	}
}

// Update implements scene.Scene.
func (p *Play) Update(ctx *scene.Context) error {
	in := ctx.Input
	switch {
	case in.JustPressed(input.Pause) || in.JustPressed(input.Back):
		ctx.Cut(scene.NewPause(p))
		return nil
	case in.JustPressed(input.Menu):
		ctx.Switch(scene.NewMainMenu())
		return nil
	case in.JustPressed(input.Restart):
		ctx.Switch(NewPlay(ctx, p.skill))
		return nil
	}
	axis := func(negative, positive input.Action) int {
		if in.Held(negative) == in.Held(positive) {
			return 0
		}
		if in.Held(negative) {
			return -1
		}
		return 1
	}
	p.advance(Controls{
		Horizontal: axis(input.Left, input.Right),
		Vertical:   axis(input.Up, input.Down),
		Jump:       in.Held(input.Confirm),
	})
	if p.phase == gameOver && p.timer >= gameOverTicks {
		ctx.Switch(scene.AfterGame(ctx, p.skill, p.score))
	}
	return nil
}

// advance runs the counter ticks that fall in one 60 Hz frame.
func (p *Play) advance(c Controls) {
	const pitFrequency = 1193182
	const divisorPerFrame = 16384 * 60
	p.clockPhase += pitFrequency
	for p.clockPhase >= divisorPerFrame {
		p.clockPhase -= divisorPerFrame
		p.counter++
		p.tick(c, autoUpdateEligible(p.counter))
	}
}

func (p *Play) tick(c Controls, eligible bool) {
	switch p.phase {
	case playing:
		result := p.stage.update(c, eligible)
		p.collect()
		switch result {
		case stageDied:
			p.stage.dropHammer()
			p.phase, p.timer = dying, 0
		case stageCleared:
			p.phase, p.timer = clearing, 0
		}
	case dying:
		p.timer++
		if p.timer >= deathFreezeTicks+deathSpinTicks+deathHoldTicks {
			if p.reserves <= 0 {
				p.phase, p.timer = gameOver, 0
				return
			}
			p.beginTurn()
		}
	case clearing:
		p.timer++
		p.stage.agePopups()
		if p.timer >= clearTicks {
			p.nextStage()
			p.enterStage()
		}
	case gameOver:
		p.timer++
	}
}

// collect moves points earned on the screen into the score and awards
// extra lives at 7000, 50,000 and every further 50,000.
func (p *Play) collect() {
	if p.stage.earned == 0 {
		return
	}
	p.score = min(p.score+p.stage.earned, config.MaxScore)
	p.stage.earned = 0
	for p.score >= p.nextLife {
		p.reserves++
		if p.nextLife == 7000 {
			p.nextLife = 50000
		} else {
			p.nextLife += 50000
		}
	}
}
