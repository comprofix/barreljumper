package classic

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/barreljumper/internal/scene"
)

// Placeholder smooth artwork for M2 (DESIGN section 2, M2 artwork): each
// shape is sized to its collision rectangle, so what the player sees is
// what the rules test. Final characters and palette come in M5.
var (
	skyColour      = color.RGBA{16, 20, 32, 255}
	girderColour   = color.RGBA{196, 74, 70, 255}
	girderEdge     = color.RGBA{238, 132, 112, 255}
	ladderColour   = color.RGBA{86, 170, 190, 255}
	reservedColour = color.RGBA{58, 96, 112, 255}
	barrelColour   = color.RGBA{176, 112, 56, 255}
	barrelBand     = color.RGBA{110, 66, 34, 255}
	specialColour  = color.RGBA{78, 132, 220, 255}
	fireColour     = color.RGBA{255, 168, 48, 255}
	fireCore       = color.RGBA{255, 236, 140, 255}
	scaredColour   = color.RGBA{110, 160, 255, 255}
	drumColour     = color.RGBA{70, 110, 190, 255}
	flameColour    = color.RGBA{255, 120, 40, 255}
	throwerColour  = color.RGBA{134, 84, 52, 255}
	throwerFace    = color.RGBA{214, 170, 120, 255}
	playerColour   = color.RGBA{251, 217, 124, 255}
	playerDead     = color.RGBA{179, 62, 62, 255}
	hammerHead     = color.RGBA{214, 214, 224, 255}
	hammerHandle   = color.RGBA{150, 104, 60, 255}
	goalColour     = color.RGBA{236, 140, 200, 255}
	rivetColour    = color.RGBA{250, 220, 90, 255}
	springColour   = color.RGBA{120, 220, 140, 255}
	pieColour      = color.RGBA{240, 200, 150, 255}
	beltMark       = color.RGBA{255, 240, 200, 255}
	itemColour     = color.RGBA{200, 120, 230, 255}
	sidebarColour  = color.RGBA{10, 12, 20, 255}
)

// field maps reference coordinates onto the screen: a 260 x 200 field
// beside a 60-pixel sidebar, as in the reference's 320 x 200 layout.
type field struct {
	dst              *ebiten.Image
	left, top, scale float32
}

func newField(dst *ebiten.Image) field {
	size := dst.Bounds().Size()
	scale := min(float32(size.X)/320, float32(size.Y)/200)
	return field{dst, (float32(size.X) - 320*scale) / 2, (float32(size.Y) - 200*scale) / 2, scale}
}

func (f field) rect(x, y, w, h float32, c color.Color) {
	vector.FillRect(f.dst, f.left+x*f.scale, f.top+y*f.scale, w*f.scale, h*f.scale, c, true)
}

func (f field) circle(x, y, r float32, c color.Color) {
	vector.FillCircle(f.dst, f.left+x*f.scale, f.top+y*f.scale, r*f.scale, c, true)
}

func (f field) line(x0, y0, x1, y1, w float32, c color.Color) {
	vector.StrokeLine(f.dst, f.left+x0*f.scale, f.top+y0*f.scale, f.left+x1*f.scale, f.top+y1*f.scale, w*f.scale, c, true)
}

// Draw implements scene.Scene.
func (p *Play) Draw(ctx *scene.Context, dst *ebiten.Image) {
	dst.Fill(sidebarColour)
	f := newField(dst)
	f.rect(0, 0, 260, 200, skyColour)
	s := p.stage
	for _, l := range s.layout.reserved {
		drawLadder(f, l, reservedColour)
	}
	for _, l := range s.ladders {
		drawLadder(f, l, ladderColour)
	}
	for _, g := range s.girders {
		drawGirder(f, g)
	}
	switch s.family {
	case barrelsFamily:
		drawGirder(f, girder{barrelsGoal.left, barrelsGoal.left + 50, barrelsGoal.y})
		f.rect(126, 6, 10, 17, goalColour) // the captive waiting on the goal platform
	case rivetsFamily:
		f.rect(126, 7, 10, 17, goalColour)
		for _, r := range s.rivets {
			if !r.removed {
				f.rect(float32(r.x)+2, float32(r.y)-2, 6, 9, rivetColour)
			}
		}
	}
	for _, p := range s.pies {
		if p.active {
			f.rect(float32(p.x)+2, float32(p.y)+2, 10, 4, pieColour) // X+2..X+11, Y+2..Y+5
		}
	}
	for _, b := range s.belts {
		// Direction marks along the belt surface.
		for x := b.x + 4; x < b.x+b.width-4; x += 16 {
			tip := float32(x) + 3*float32(b.dir)
			f.line(float32(x), float32(b.y)+2, tip, float32(b.y)+3.5, 0.8, beltMark)
			f.line(tip, float32(b.y)+3.5, float32(x), float32(b.y)+5, 0.8, beltMark)
		}
	}
	for _, p := range s.springs {
		if p.active {
			h := float32(12)
			if p.compressing {
				h = 8
			}
			f.rect(float32(p.x)+2, float32(p.y)+14-h, 10, h, springColour) // X+2..X+11, Y+2..Y+13
		}
	}
	for _, it := range s.items {
		if it.active {
			f.rect(float32(it.x)+1, float32(it.y), 11, 10, itemColour) // X+1..X+11, Y..Y+9
		}
	}
	drawThrower(f, s.thrower)
	if s.oil.present {
		drawDrum(f, s.oil, ctx.Tick)
	}
	for _, h := range s.pickups {
		if h.active {
			f.rect(float32(h.x)+3, float32(h.y)+3, 2, 5, hammerHandle)
			f.rect(float32(h.x), float32(h.y), 10, 4, hammerHead)
		}
	}
	for _, b := range s.barrels {
		if b.active {
			drawBarrel(f, b)
		}
	}
	for _, fb := range s.fireballs {
		if fb.active {
			drawFireball(f, fb, s.hammer.held, ctx.Tick)
		}
	}
	p.drawPlayer(f)
	if s.smash.active {
		r := float32(2 + s.smash.frame)
		f.circle(float32(s.smash.x)+5, float32(s.smash.y)+5, r, fireCore)
		f.circle(float32(s.smash.x)+5, float32(s.smash.y)+5, r*0.6, skyColour)
	}
	p.drawText(ctx, f)
}

func drawGirder(f field, g girder) {
	w := float32(g.right - g.left)
	f.rect(float32(g.left), float32(g.y), w, 7, girderColour)
	f.rect(float32(g.left), float32(g.y), w, 1.5, girderEdge)
	for x := g.left; x < g.right; x += 10 {
		f.line(float32(x), float32(g.y)+1, float32(x)+5, float32(g.y)+6, 0.8, girderEdge)
		f.line(float32(x)+5, float32(g.y)+6, float32(x)+10, float32(g.y)+1, 0.8, girderEdge)
	}
}

func drawLadder(f field, l ladder, c color.Color) {
	if l.height == 0 {
		return
	}
	x, top, h := float32(l.x), float32(l.top), float32(l.height)
	f.rect(x, top, 1, h, c)
	f.rect(x+7, top, 1, h, c)
	for y := top + 2; y < top+h; y += 4 {
		f.rect(x, y, 8, 1, c)
	}
}

func drawThrower(f field, t thrower) {
	x, y := float32(t.x), float32(t.y)
	f.rect(x+2, y+8, 36, 24, throwerColour) // the boss contact box is X+2..X+37, Y..Y+31
	f.circle(x+20, y+8, 9, throwerColour)
	f.circle(x+20, y+9, 5, throwerFace)
	switch t.pose {
	case throwerPickUp:
		f.rect(x-2, y+14, 6, 12, throwerColour)
	case throwerThrow:
		f.rect(x+36, y+12, 8, 6, throwerColour)
	case throwerTauntA:
		f.rect(x, y+2, 6, 10, throwerColour)
		f.rect(x+34, y+14, 6, 10, throwerColour)
	case throwerTauntB:
		f.rect(x, y+14, 6, 10, throwerColour)
		f.rect(x+34, y+2, 6, 10, throwerColour)
	}
	if t.holding {
		f.rect(x+11, y+16, 14, 10, specialColour)
	}
}

func drawDrum(f field, o oilDrum, tick int64) {
	x, y := float32(o.x), float32(o.y)
	f.rect(x+2, y+8, 14, 22, drumColour) // contact: X+2..X+15 at the drum's lip, Y+8..Y+11
	f.rect(x+2, y+14, 14, 2, sidebarColour)
	if o.burning {
		h := float32(4 + (tick/4)%3)
		f.rect(x+4, y+8-h, 10, h, flameColour)
	}
}

func drawBarrel(f field, b barrel) {
	x, y := float32(b.x), float32(b.y-b.bounce)
	if b.kind != regularBarrel {
		f.rect(x+2, y+1, 14, 10, specialColour) // X+2..X+16
		f.rect(x+2, y+5, 14, 1.5, barrelBand)
		return
	}
	f.circle(x+7, y+6, 5, barrelColour) // X+2..X+12, Y+1..Y+10
	angle := float32(b.frame) * 0.25
	f.line(x+7-4*(1-angle), y+2+angle*8, x+7+4*(1-angle), y+10-angle*8, 1, barrelBand)
}

func drawFireball(f field, fb fireball, scared bool, tick int64) {
	x, y := float32(fb.x), float32(fb.y)
	c := fireColour
	if scared {
		c = scaredColour
	}
	flicker := float32((tick / 3) % 2)
	f.rect(x+3, y+6, 10, 9, c) // X+3..X+12, Y+2..Y+14
	f.circle(x+8, y+6, 4+flicker, c)
	f.circle(x+8, y+9, 2.5, fireCore)
}

func (p *Play) drawPlayer(f field) {
	s := p.stage
	m := s.player
	left, right := playerSpan(m)
	c := playerColour
	if p.phase == dying || m.Dead {
		c = playerDead
	}
	x, y := float32(left), float32(m.Y)
	w := float32(right - left + 1)
	if p.phase == dying && p.timer >= deathFreezeTicks && p.timer < deathFreezeTicks+deathSpinTicks {
		// The spin: a rotating bar in place of the body.
		phase := float32(((p.timer-deathFreezeTicks)/7)%4) / 4 * 6.283
		cx, cy := x+w/2, y+7
		f.line(cx-7*cos(phase), cy-7*sin(phase), cx+7*cos(phase), cy+7*sin(phase), 4, c)
		return
	}
	f.rect(x, y+4, w, 11, c)
	f.circle(x+w/2, y+3, 3.5, c)
	eye := x + w - 3
	if m.Facing < 0 {
		eye = x + 2
	}
	f.rect(eye, y+2, 1.2, 1.2, skyColour)
	if s.hammer.held && p.phase == playing {
		h := s.hammer
		hw, hh := float32(12), float32(6)
		hx := float32(h.x)
		if h.forward {
			hw, hh = 10, 10
			if m.Facing > 0 {
				hx += 2
			}
		}
		head := hammerHead
		if h.warning() {
			head = flameColour
		}
		f.line(x+w/2, y+7, hx+hw/2, float32(h.y)+hh/2, 1.5, hammerHandle)
		f.rect(hx, float32(h.y), hw, hh, head)
	}
}

func (p *Play) drawText(ctx *scene.Context, f field) {
	font, sc := ctx.Fonts.Small, ctx.Scheme
	if font == nil {
		return
	}
	at := func(x, y float32) (int, int) {
		return int(f.left + x*f.scale), int(f.top + y*f.scale)
	}
	line := func(row int, label, value string) {
		x, y := at(266, float32(10+row*30))
		font.Draw(f.dst, label, x, y, sc.Dim)
		font.Draw(f.dst, value, x, y+font.LineHeight()+2, sc.Text)
	}
	score := "SCORE"
	if len(p.players) > 1 {
		score = fmt.Sprint("PLAYER ", p.current+1)
		other := 1 - p.current
		line(5, fmt.Sprint("PLAYER ", other+1), fmt.Sprint(p.players[other].score))
	}
	line(0, score, fmt.Sprint(p.score))
	line(1, "BONUS", fmt.Sprint(p.stage.bonus))
	line(2, "LIVES", fmt.Sprint(p.reserves))
	line(3, "LEVEL", fmt.Sprint(p.level+1))
	for _, pu := range p.stage.popups {
		x, y := at(float32(pu.x), float32(pu.y-pu.age/6))
		font.Draw(f.dst, fmt.Sprint(pu.points), x, y, sc.Highlight)
	}
	message := ""
	switch p.phase {
	case clearing:
		message = "STAGE CLEAR"
	case playerOut:
		message = fmt.Sprint("GAME OVER PLAYER ", p.current+1)
	case gameOver:
		message = "GAME OVER"
	}
	if message != "" {
		x, y := at(130, 100)
		font.DrawShadowed(f.dst, message, x-font.Width(message)/2, y, sc.Title, sc.Shadow)
	}
}

func cos(a float32) float32 { return float32(math.Cos(float64(a))) }
func sin(a float32) float32 { return float32(math.Sin(float64(a))) }
