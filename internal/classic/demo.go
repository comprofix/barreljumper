package classic

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/scene"
)

// The attract demo (DESIGN section 6, Title, attract and music): the first
// Barrels screen played by a simple pilot from a fixed seed, after its
// floor card. It is original presentation, not the reference's demo.

// demoSeed is the random seed the demo plays from, chosen for a run that
// shows the screen well (see TestDemoRun).
const demoSeed = 14

// demoTicks is the longest the demo plays: 90 seconds.
const demoTicks = 90 * config.TPS

// demoRoute is the full ladders the pilot climbs, bottom to top, by their
// index in barrelsLadders.
var demoRoute = []int{0, 4, 6, 5, 7, 16}

// The pilot's judgement, in pixels: it jumps a hazard whose gap closes to
// between demoJumpNear and demoJumpFar, and climbs only when no barrel on
// its girder is within demoClimbClear of the ladder.
const demoJumpNear, demoJumpFar, demoClimbClear = 12, 18, 60

// Demo is the game playing itself behind the title.
type Demo struct {
	p      *Play
	ticks  int
	next   int  // the route's next ladder
	died   bool // the death sequence has started
	ending bool // the screen has been cleared
	over   bool
}

// NewDemo starts the demo.
func NewDemo(*scene.Context) scene.Scene { return newDemo(demoSeed) }

func newDemo(seed uint32) *Demo {
	p := NewPlay(nil, *config.SkillNamed("arcade"))
	p.rng = random{seed}
	p.enterStage()
	return &Demo{p: p}
}

// Over implements scene.DemoOver: after a death, a clear or the time limit.
func (d *Demo) Over() bool { return d.over }

// Update implements scene.Scene.
func (d *Demo) Update(ctx *scene.Context) error {
	d.step()
	d.p.playSounds(ctx)
	return nil
}

// step runs one frame of the demo.
func (d *Demo) step() {
	if d.over {
		return
	}
	d.ticks++
	p := d.p
	p.advance(d.pilot())
	switch p.phase {
	case dying:
		d.died = true
	case cheering, carryOff:
		d.ending = true
	case floorCard:
		// The death sequence or the carry-off has run its course.
		d.over = d.over || d.died || d.ending
	}
	if p.phase == playerOut || p.phase == gameOver || d.ticks >= demoTicks {
		d.over = true
	}
}

// pilot chooses the demo's controls: walk to the route's next ladder and
// climb it, and jump a barrel or fireball coming along the girder.
func (d *Demo) pilot() Controls {
	p := d.p
	if p.phase != playing {
		return Controls{}
	}
	s := p.stage
	m := s.player
	if m.ladder != nil {
		return Controls{Vertical: -1}
	}
	if m.jumping || m.falling {
		return Controls{}
	}
	if d.danger() {
		return Controls{Jump: true}
	}
	if d.next >= len(demoRoute) {
		return Controls{Vertical: -1}
	}
	l := barrelsLadders[demoRoute[d.next]]
	foot := m.Y + ladderFoot
	if foot <= l.top+2 {
		// Above this ladder already: on to the next.
		d.next++
		return Controls{}
	}
	if _, ok := l.attach(m.X, foot, true); ok && !m.Hammer {
		if d.clearToClimb(l) {
			return Controls{Vertical: -1}
		}
		return Controls{}
	}
	dx := l.x - 4 - m.X
	switch {
	case dx > 0:
		return Controls{Horizontal: 1}
	case dx < 0:
		return Controls{Horizontal: -1}
	}
	return Controls{}
}

// danger reports a barrel or fireball on the player's girder about to
// reach them.
func (d *Demo) danger() bool {
	s := d.p.stage
	m := s.player
	left, right := playerSpan(m)
	near := func(x, w, y, h, dx int) bool {
		if y+h < m.Y+6 || y > m.Y+15 {
			return false
		}
		gap := -1
		switch {
		case x >= right && dx <= 0:
			gap = x - right
		case x+w <= left && dx >= 0:
			gap = left - (x + w)
		}
		return gap >= demoJumpNear && gap <= demoJumpFar
	}
	for _, b := range s.barrels {
		if b.active && b.ladder < 0 && near(b.x, 14, b.y, 12, b.dx) {
			return true
		}
	}
	for _, f := range s.fireballs {
		if f.active && near(f.x, 16, f.y, 16, f.dx) {
			return true
		}
	}
	return false
}

// clearToClimb reports no barrel rolling toward the ladder's foot along
// the player's girder, and none on the ladder.
func (d *Demo) clearToClimb(l ladder) bool {
	s := d.p.stage
	m := s.player
	for _, b := range s.barrels {
		if !b.active {
			continue
		}
		if b.ladder >= 0 && b.x-l.x < 12 && l.x-b.x < 12 && b.y < m.Y+16 {
			return false
		}
		if b.ladder < 0 && b.y+12 >= m.Y && b.y <= m.Y+15 {
			if b.x < l.x && b.dx > 0 && l.x-b.x < demoClimbClear || b.x > l.x && b.dx < 0 && b.x-l.x < demoClimbClear {
				return false
			}
		}
	}
	return true
}

// Draw implements scene.Scene: the screen, with DEMO flashing in the
// sidebar.
func (d *Demo) Draw(ctx *scene.Context, dst *ebiten.Image) {
	d.p.Draw(ctx, dst)
	if ctx.Tick/30%2 == 0 {
		f := newField(dst)
		small, large, sc := ctx.Fonts.Small, ctx.Fonts.Large, ctx.Scheme
		if large == nil {
			large = small
		}
		const word = "DEMO"
		large.DrawShadowed(dst, word, int(f.left+290*f.scale)-large.Width(word)/2, int(f.top+170*f.scale), sc.Title, sc.Shadow)
	}
}
