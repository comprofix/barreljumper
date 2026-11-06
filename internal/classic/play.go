package classic

import (
	"github.com/comprofix/barreljumper/internal/audio"
	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/input"
	"github.com/comprofix/barreljumper/internal/scene"
)

// Play phases.
const (
	playing = iota
	dying
	playerOut // one player of two has no reserve left
	gameOver
	opening    // the game's opening scene
	floorCard  // the floor card before a screen's play
	cheering   // the heart by the captive after a clear
	carryOff   // the thrower hoisted away with the captive
	rivetsDone // the Rivets ending
)

// Presentation timing in reference counter ticks (about 72.8 per second).
// The death times follow the reference death sequence with sound muted.
const (
	deathFreezeTicks = 73   // 1.0 s still
	deathSpinTicks   = 140  // 20 frames of 7 ticks
	deathHoldTicks   = 146  // 2.0 s
	heartTicks       = 314  // 4.31 s, the heart after a clear
	carryTicks       = 305  // 4.19 s, the captive carried off
	cardTicks        = 223  // 3.07 s, the floor card
	rivetsEndTicks   = 1063 // 14.6 s, the Rivets ending
	openingTicks     = 910  // 12.5 s, the reference opening's length
	gameOverTicks    = 218  // 3.0 s, shorter than the reference (owner decision)
	playerOutTicks   = 655  // 9.0 s, the "Game Over Player n" message
)

// player is one player's place in the game. Each keeps their own score,
// lives and schedule position (DESIGN section 4, M4 session and turns).
type player struct {
	skill     config.Skill
	skillWord int
	score     int
	reserves  int
	level     int
	index     int // position in the level's schedule row
	nextLife  int
}

// Play is one Classic game, following the Classic schedule (DESIGN
// section 4, Screen progression).
type Play struct {
	*player // the current player
	players []player
	current int
	chosen  config.Skill // the skill the game was started with, for a restart
	rng     random
	updates int
	stage   *stage
	phase   int
	sounds  sound       // effects to play at the end of the frame
	jingle  *audio.Tune // a jingle to start at the end of the frame
	timer   int
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
	p := &Play{chosen: skill, rng: random{seed}}
	// The game menu's setup: the player count, the starting level and the
	// other players' skills. The -level flag overrides the level, for tests.
	skills, level := []config.Skill{skill}, 1
	if ctx != nil && (ctx.MaxPlayers > 1 || ctx.StartLevels > 1) {
		g := ctx.Game()
		level = g.Level
		for _, id := range g.Skills[1:g.Players] {
			skills = append(skills, *config.SkillNamed(id))
		}
	}
	if ctx != nil && ctx.StartLevel > 1 {
		level = ctx.StartLevel
	}
	for _, sk := range skills {
		pl := player{skill: sk, skillWord: skillWord(sk), reserves: 3, nextLife: 7000, level: level - 1}
		if pl.skillWord == easySkill {
			pl.reserves = 5
		}
		if ctx != nil && ctx.StartScreen > 1 {
			// Past the row's end, the last screen.
			pl.index = min(ctx.StartScreen, len(classicSchedule[pl.level%5])) - 1
		}
		p.players = append(p.players, pl)
	}
	p.player = &p.players[0]
	p.beginTurn()
	if p.family() == barrelsFamily && p.index == 0 {
		p.phase, p.timer, p.jingle = opening, 0, nil
	}
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

func (p *Play) enterStage() {
	p.stage = newStage(p.family(), &p.rng, p.skillWord, p.level, &p.updates)
	p.phase, p.timer = floorCard, 0
	p.jingle = startJingle
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
		ctx.Switch(NewPlay(ctx, p.chosen))
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
	defer p.playSounds(ctx)
	if in.JustPressed(input.Confirm) {
		p.skip()
	}
	p.advance(Controls{
		Horizontal: axis(input.Left, input.Right),
		Vertical:   axis(input.Up, input.Down),
		Jump:       in.Held(input.Confirm),
	})
	if p.phase == gameOver && p.timer >= gameOverTicks {
		results := make([]scene.Result, len(p.players))
		for i, pl := range p.players {
			results[i] = scene.Result{Skill: pl.skill, Points: pl.score}
		}
		ctx.Switch(scene.AfterGame(ctx, results...))
	}
	return nil
}

// skip answers Confirm: it cuts short the opening, to the floor card, and
// the "Game Over Player n" message, to the next player's turn.
func (p *Play) skip() {
	switch p.phase {
	case opening:
		p.phase, p.timer, p.jingle = floorCard, 0, startJingle
	case playerOut:
		p.passTurn()
	}
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
		p.sounds |= p.stage.sounds
		p.stage.sounds = 0
		switch result {
		case stageDied:
			p.stage.dropHammer()
			p.phase, p.timer = dying, 0
			p.sounds |= soundHit
		case stageCleared:
			p.phase, p.timer = cheering, 0
			if p.stage.family == rivetsFamily {
				p.phase = rivetsDone
				p.sounds |= soundRumble
			}
			p.jingle = clearJingle
		}
	case dying:
		p.timer++
		switch p.timer {
		case deathFreezeTicks:
			p.sounds |= soundSpin
		case deathFreezeTicks + deathSpinTicks:
			p.sounds |= soundThud
		}
		if p.timer >= deathFreezeTicks+deathSpinTicks+deathHoldTicks {
			switch {
			case p.out():
				p.phase, p.timer = gameOver, 0
			case p.reserves <= 0:
				p.phase, p.timer = playerOut, 0
			default:
				p.passTurn()
			}
		}
	case playerOut:
		p.timer++
		if p.timer >= playerOutTicks {
			p.passTurn()
		}
	case opening:
		p.timer++
		p.sounds |= openingSounds(p.timer)
		if p.timer >= openingTicks {
			p.phase, p.timer, p.jingle = floorCard, 0, startJingle
		}
	case floorCard:
		p.timer++
		if p.timer >= cardTicks {
			p.phase, p.timer = playing, 0
		}
	case cheering:
		p.timer++
		p.stage.agePopups()
		if p.timer >= heartTicks {
			p.phase, p.timer = carryOff, 0
		}
	case carryOff:
		p.timer++
		p.stage.agePopups()
		if p.timer == carryHoist {
			p.sounds |= soundWinch
		}
		if p.timer >= carryTicks {
			p.nextStage()
			p.enterStage()
		}
	case rivetsDone:
		p.timer++
		p.stage.agePopups()
		p.sounds |= rivetsSounds(p.timer)
		if p.timer == rivetsLowered {
			p.stage.meet()
		}
		if p.timer >= rivetsEndTicks {
			p.nextStage()
			p.enterStage()
		}
	case gameOver:
		p.timer++
	}
}

// playSounds starts the frame's jingle and effects.
func (p *Play) playSounds(ctx *scene.Context) {
	if p.jingle != nil {
		ctx.Audio.PlayTune(p.jingle)
		p.jingle = nil
	}
	play(ctx.Audio, p.sounds)
	p.sounds = 0
}

// out reports whether no player has a reserve left.
func (p *Play) out() bool {
	for _, pl := range p.players {
		if pl.reserves > 0 {
			return false
		}
	}
	return true
}

// passTurn gives the next turn to the next player with a reserve, as the
// reference does after a death: one step, and one more if that player has
// none. With one player the turn stays.
func (p *Play) passTurn() {
	n := len(p.players)
	p.current = (p.current + 1) % n
	if p.players[p.current].reserves <= 0 {
		p.current = (p.current + 1) % n
	}
	p.player = &p.players[p.current]
	p.beginTurn()
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
