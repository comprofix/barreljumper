package classic

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/barreljumper/internal/scene"
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
	drawBackdrop(f)
	s := p.stage
	tick := ctx.Tick
	if s.family == elevatorsFamily {
		drawCables(f)
	}
	for _, l := range s.layout.reserved {
		drawLadder(f, l, true)
	}
	for _, l := range s.ladders {
		drawLadder(f, l, false)
	}
	cells := len(s.girders)
	if s.family == elevatorsFamily {
		cells -= len(s.elevators)
	}
	for i, g := range s.girders {
		drawGirder(f, g)
		if i >= cells {
			drawElevatorCell(f, g)
		}
	}
	for _, b := range s.belts {
		if b.width > 0 {
			drawBelt(f, b, tick)
		}
	}
	switch s.family {
	case barrelsFamily:
		drawGirder(f, girder{barrelsGoal.left, barrelsGoal.left + 50, barrelsGoal.y})
		drawCaptive(f, 126, 6, tick)
	case rivetsFamily:
		drawCaptive(f, 126, 7, tick)
		for _, r := range s.rivets {
			if !r.removed {
				drawRivet(f, float32(r.x), float32(r.y))
			}
		}
	case elevatorsFamily:
		drawCaptive(f, 146, 7, tick)
	case conveyorFamily:
		drawCaptive(f, 146, 5, tick)
	}
	for i, it := range s.items {
		if it.active {
			drawItem(f, i, float32(it.x), float32(it.y))
		}
	}
	for _, h := range s.pickups {
		if h.active {
			drawPickup(f, float32(h.x), float32(h.y))
		}
	}
	if s.oil.present {
		drawDrum(f, s.oil, tick)
	}
	drawThrower(f, s.thrower)
	for _, p := range s.pies {
		if p.active {
			drawPie(f, float32(p.x), float32(p.y))
		}
	}
	for _, p := range s.springs {
		if p.active {
			drawSpring(f, p)
		}
	}
	for _, b := range s.barrels {
		if b.active {
			drawBarrel(f, b)
		}
	}
	for _, fb := range s.fireballs {
		if fb.active {
			drawFireball(f, fb, s.hammer.held, tick)
		}
	}
	p.drawPlayer(f)
	if s.smash.active {
		drawSmash(f, float32(s.smash.x), float32(s.smash.y), s.smash.frame)
	}
	drawSidebar(f)
	p.drawText(ctx, f)
}

// drawSidebar frames the score panel right of the field.
func drawSidebar(f field) {
	f.rect(260, 0, 60, 200, sidebarColour)
	f.rect(260, 0, 1, 200, sidebarEdge)
	for i := float32(0); i < 60; i += 6 {
		f.poly(hazardYellow, 261+i, 196, 264+i, 196, 267+i, 200, 264+i, 200)
		f.poly(hazardYellow, 261+i, 0, 264+i, 0, 267+i, 4, 264+i, 4)
	}
}

// drawPlayer draws the worker in the player's span, Y..Y+15: hard hat,
// hi-vis vest, jeans and boots, walking, climbing or jumping.
func (p *Play) drawPlayer(f field) {
	s := p.stage
	m := s.player
	left, right := playerSpan(m)
	x, y := float32(left), float32(m.Y)
	w := float32(right - left + 1)
	cx := x + w/2
	dying := p.phase == dying || m.Dead
	facing := float32(m.Facing)
	if dying && p.timer >= deathFreezeTicks && p.timer < deathFreezeTicks+deathSpinTicks {
		// The spin: the worker turns on the spot, facing each way in turn.
		if (p.timer-deathFreezeTicks)/7%2 == 1 {
			facing = -facing
		}
	}
	if dying && p.timer >= deathFreezeTicks+deathSpinTicks {
		// Flat on the girder, hat knocked off.
		f.roundRect(cx-7, y+11, 14, 4, 1.5, vestOrange)
		f.rect(cx-1, y+11, 1, 4, vestStripe)
		f.roundRect(cx+5, y+11.5, 5, 3, 1, denim)
		f.circle(cx-8, y+13, 2.5, skin)
		f.ellipse(cx-12, y+14, 3, 1.5, hatYellow)
		return
	}
	f.ellipse(cx, y+16, 6, 1, shadowColour)
	body := vestOrange
	if dying {
		body = hurt
	}
	climbing := m.ladder != nil && !m.grounded
	stride := float32(0)
	switch {
	case m.jumping || m.falling:
		stride = 2.5
	case climbing:
		stride = float32((m.Y/2)%2)*2 - 1
	default:
		stride = float32((m.X/3)%3) - 1
	}
	// Legs and boots.
	if climbing {
		f.rect(cx-3.5, y+10, 2.6, 4+stride, denim)
		f.rect(cx+0.9, y+10, 2.6, 4-stride, denim)
		f.rect(cx-3.8, y+13.5+stride, 3, 2, boot)
		f.rect(cx+0.8, y+13.5-stride, 3, 2, boot)
	} else {
		back, front := cx-1.3-stride*facing, cx-1.3+stride*facing
		f.rect(back, y+10, 2.6, 4.5, denimDark)
		f.rect(front, y+10, 2.6, 4.5, denim)
		f.rect(back-0.3+facing*0.6, y+14, 3.2, 1.6, boot)
		f.rect(front-0.3+facing*0.6, y+14, 3.2, 1.6, boot)
	}
	// Torso in the vest, reflective stripes across it.
	f.roundRect(cx-4, y+5, 8, 6, 1.5, body)
	f.rect(cx-4, y+7.5, 8, 0.9, vestStripe)
	f.rect(cx-4, y+9.5, 8, 0.9, vestStripe)
	// Arms: up the rails when climbing, swinging when walking.
	switch {
	case climbing:
		f.line(cx-3.5, y+6, cx-4, y+1.5+stride, 1.6, skin)
		f.line(cx+3.5, y+6, cx+4, y+1.5-stride, 1.6, skin)
	case m.jumping:
		f.line(cx-3.5, y+6, cx-5.5, y+3, 1.6, skin)
		f.line(cx+3.5, y+6, cx+5.5, y+3, 1.6, skin)
	case !s.hammer.held:
		f.line(cx+3.5*facing, y+6, cx+(3.5+stride*0.8)*facing, y+10, 1.6, skin)
	}
	// Head and hard hat.
	f.circle(cx, y+3.2, 2.6, skin)
	if !climbing {
		f.rect(cx+1.2*facing-0.5, y+2.6, 1, 1, eyeColour)
		f.rect(cx-2.3*facing-0.4, y+3.2, 0.8, 1.6, skinShade)
	}
	f.ellipse(cx, y+1.4, 3.2, 2, hatYellow)
	f.rect(cx-3.2+facing*0.8, y+1.6, 6.4, 0.9, hatShade)
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
		drawHammer(f, hx, float32(h.y), hw, hh, cx+2*facing, y+7, h.warning())
		f.circle(cx+2*facing, y+7, 1.2, skin)
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
