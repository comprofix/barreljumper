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
	switch p.phase {
	case floorCard:
		p.drawFloorCard(f)
	case opening:
		p.drawOpening(f, ctx.Tick)
	case rivetsDone:
		p.drawRivetsEnd(f, ctx.Tick)
	default:
		p.drawScreen(f, ctx.Tick)
	}
	drawSidebar(f)
	p.drawText(ctx, f)
}

// drawScreen draws a screen in play, and its heart and carry-off after a
// clear.
func (p *Play) drawScreen(f field, tick int64) {
	s := p.stage
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
	if s.family == barrelsFamily {
		drawGirder(f, girder{barrelsGoal.left, barrelsGoal.left + 50, barrelsGoal.y})
	}
	cx, cy := captiveAt(s.family)
	if p.phase != carryOff {
		drawCaptive(f, cx, cy, tick)
	}
	for _, r := range s.rivets {
		if !r.removed {
			drawRivet(f, float32(r.x), float32(r.y))
		}
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
	if p.phase == carryOff {
		th, gx, gy, hx, hy, grabbed := p.carried()
		drawCaptive(f, gx, gy, tick) // behind him, held
		drawThrower(f, th)
		drawHook(f, hx, hy)
		drawHeart(f, cx+16, cy+1, grabbed)
	} else {
		drawThrower(f, s.thrower)
	}
	if p.phase == cheering {
		drawHeart(f, cx+16, cy+1, false)
	}
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
		drawSmash(f, s.smash)
	}
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
// hi-vis vest over a work shirt, jeans and boots, walking, climbing or
// jumping. Like the other figures he is lit from the left.
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
		p.drawFallen(f, cx, y, facing)
		return
	}
	f.ellipse(cx, y+16, 6, 1, shadowColour)
	vest := vestOrange
	if dying {
		vest = hurt
	}
	climbing := m.ladder != nil && !m.grounded
	stride := float32(0)
	switch {
	case m.jumping || m.falling:
		stride = 2.5
	case climbing:
		// Hands and feet change on the rung sounds' 4-pixel steps.
		stride = float32(footfall(m.Y)&1)*2 - 1
	default:
		// The feet change on the footstep sounds' 4-pixel steps.
		stride = float32(footfall(m.X)&1)*2.4 - 1.2
	}
	at := func(dx float32) float32 { return cx + dx*facing }
	walking := !climbing && !m.jumping && !s.hammer.held
	if walking {
		// The far arm swings opposite the near one, mostly behind him.
		drawLimb(f, at(-2.4), y+6, at(-3.8-stride*0.9), y+9.6, true)
	}
	// Legs and boots.
	if climbing {
		drawLeg(f, cx-2.2, y+10.4, 3.6+stride, false)
		drawLeg(f, cx+2.2, y+10.4, 3.6-stride, false)
		drawBoot(f, cx-2.2, y+13.6+stride, 0)
		drawBoot(f, cx+2.2, y+13.6-stride, 0)
	} else {
		back, front := cx-stride*facing, cx+stride*facing
		drawLeg(f, back, y+10.4, 4, true)
		drawBoot(f, back+0.3*facing, y+14, facing)
		drawLeg(f, front, y+10.4, 4, false)
		drawBoot(f, front+0.3*facing, y+14, facing)
	}
	// A belt with a brass buckle.
	f.rect(cx-3.9, y+10, 7.8, 0.9, beltColour)
	if !climbing {
		f.rect(at(1.6)-0.5, y+10.05, 1, 0.8, boltHead)
	}
	drawVest(f, cx, y, at, vest, !climbing)
	// Arms out when jumping, swinging when walking.
	switch {
	case m.jumping:
		drawLimb(f, cx-3.3, y+6, cx-5.8, y+2.8, false)
		drawLimb(f, cx+3.3, y+6, cx+5.8, y+2.8, false)
	case walking:
		drawLimb(f, at(3), y+6, at(4.2+stride*0.9), y+9.6, false)
	case s.hammer.held:
		drawLimb(f, at(-0.6), y+6.2, cx+2*facing, y+7, false)
	}
	drawHead(f, cx, y, at, climbing)
	if climbing {
		// Up the rails, hands over his hat.
		drawLimb(f, cx-3.2, y+6, cx-4, y+1.5+stride, false)
		drawLimb(f, cx+3.2, y+6, cx+4, y+1.5-stride, false)
	}
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
		drawHammer(f, hx, float32(h.y), hw, hh, cx+2*facing, y+7, h.forward, h.warning())
		drawFist(f, cx+2*facing, y+7)
	}
}

// Worker colours that only the standing worker uses.
var (
	workShirt  = rgb(0x34507a)
	hairColour = rgb(0x5a3a22)
	beltColour = rgb(0x3a2614)
	white      = rgb(0xffffff)
)

// drawLimb draws an arm from the shoulder: a rolled sleeve, the forearm
// and the hand. A far arm is drawn darker.
func drawLimb(f field, sx, sy, ex, ey float32, far bool) {
	shirt, arm := workShirt, skin
	if far {
		shirt, arm = mix(workShirt, eyeColour, 0.35), mix(skin, eyeColour, 0.3)
	}
	mx, my := sx+(ex-sx)*0.42, sy+(ey-sy)*0.42
	f.line(mx, my, ex, ey, 1.45, arm)
	f.line(sx, sy, mx, my, 2, shirt)
	f.circle(mx, my, 1, mix(shirt, white, 0.15)) // the rolled cuff
	f.circle(ex, ey, 0.95, arm)
	f.circle(ex+0.3, ey+0.3, 0.45, mix(arm, eyeColour, 0.2))
}

// drawFist draws the hand closed on the hammer handle.
func drawFist(f field, x, y float32) {
	f.circle(x, y, 1.25, skin)
	f.circle(x+0.35, y+0.4, 0.6, skinShade)
	f.circle(x-0.4, y-0.4, 0.4, mix(skin, white, 0.35))
}

// drawLeg draws a jeans leg hanging from top, shaded on its right with a
// turned-up cuff. A far leg is drawn darker.
func drawLeg(f field, lx, top, length float32, far bool) {
	base := denim
	if far {
		base = denimDark
	}
	f.rect(lx-1.3, top, 2.6, length, base)
	f.rect(lx+0.6, top, 0.7, length, mix(base, eyeColour, 0.3))
	f.rect(lx-1.3, top, 0.5, length, mix(base, white, 0.18))
	f.rect(lx-1.3, top+length*0.45, 2.6, 0.25, mix(base, eyeColour, 0.2)) // the knee
	f.rect(lx-1.4, top+length-0.8, 2.8, 0.8, mix(base, white, 0.12))      // the cuff
}

// drawBoot draws a work boot with its toe toward facing, or from behind
// when facing is zero.
func drawBoot(f field, bx, by, facing float32) {
	sole := mix(boot, eyeColour, 0.6)
	if facing == 0 {
		f.roundRect(bx-1.5, by, 3, 2, 0.7, boot)
		f.rect(bx-1.5, by+1.6, 3, 0.45, sole)
		f.rect(bx-0.9, by+0.3, 0.6, 0.9, mix(boot, white, 0.18))
		return
	}
	f.roundRect(bx-1.7, by, 3.4, 1.8, 0.7, boot)
	f.circle(bx+1.1*facing, by+1, 0.85, mix(boot, white, 0.12)) // the steel toe
	f.rect(bx-1.8, by+1.4, 3.6, 0.45, sole)
	f.rect(bx-1.1, by+0.25, 1.4, 0.35, mix(boot, white, 0.25)) // laces
}

// drawVest draws the hi-vis vest over the work shirt, lit from the left,
// with reflective bands; the front shows the collar and an ID badge.
func drawVest(f field, cx, y float32, at func(float32) float32, vest color.RGBA, front bool) {
	dark, light := mix(vest, eyeColour, 0.3), mix(vest, white, 0.3)
	f.roundRect(cx-4, y+5, 8, 5.3, 1.5, vest)
	f.rect(cx+2.5, y+5.7, 1.5, 4.4, dark)
	f.rect(cx-4, y+5.7, 0.7, 4, light)
	f.rect(cx-3.6, y+9.9, 7.2, 0.4, dark)
	if front {
		// The shirt collar in the vest's open front.
		nx := at(0.9)
		f.poly(workShirt, nx-1.4, y+5, nx+1.4, y+5, nx, y+7)
		f.poly(mix(workShirt, white, 0.25), nx-1.4, y+5, nx-0.3, y+5, nx-0.5, y+6.1)
	}
	// Bands over the shoulders and around the body.
	band := func(bx, by, bw, bh float32) {
		f.rect(bx, by, bw, bh, vestStripe)
		f.rect(bx, by+bh-0.25, bw, 0.25, mix(vestStripe, eyeColour, 0.35))
	}
	band(cx-2.6, y+5.1, 0.8, 2.3)
	band(cx+1.8, y+5.1, 0.8, 2.3)
	band(cx-4, y+7.4, 8, 0.85)
	band(cx-4, y+9, 8, 0.75)
	f.rect(cx-3.3, y+7.5, 1.2, 0.25, white) // a glint
	if front {
		bx := at(-1.6) - 0.55
		f.rect(bx, y+8.25, 1.1, 0.75, rgb(0xe8eef4)) // the ID badge
		f.rect(bx+0.15, y+8.4, 0.5, 0.3, blueLight)
	}
}

// drawHead draws the head and hard hat: in profile toward facing, or from
// behind when climbing.
func drawHead(f field, cx, y float32, at func(float32) float32, back bool) {
	f.rect(cx-1, y+4.6, 2, 1, skinShade) // the neck
	if back {
		f.circle(cx-2.5, y+3.4, 0.7, skinShade)
		f.circle(cx+2.5, y+3.4, 0.7, skinShade)
		f.circle(cx, y+3.2, 2.6, skin)
		f.ellipse(cx, y+3.4, 2.5, 1.9, hairColour)
		f.ellipse(cx-0.8, y+2.9, 0.9, 0.6, mix(hairColour, white, 0.15))
	} else {
		f.circle(cx, y+3.2, 2.6, skinShade)
		f.ellipse(at(0.4), y+3.2, 2.3, 2.5, skin)
		f.ellipse(at(1.2), y+3.8, 0.8, 0.6, mix(skin, rgb(0xe07060), 0.25)) // cheek
		f.ellipse(at(-1.7), y+3.1, 1, 1.4, hairColour)                      // sideburn
		f.circle(at(-0.9), y+3.6, 0.75, skinShade)                          // ear
		f.circle(at(-0.9), y+3.6, 0.4, skin)
		f.ellipse(at(1.3), y+3, 0.55, 0.62, white) // eye
		f.circle(at(1.5), y+3.05, 0.33, eyeColour)
		f.line(at(0.7), y+2.2, at(1.9), y+2.35, 0.35, hairColour) // brow
		f.circle(at(2.45), y+3.6, 0.62, skin)                     // nose
		f.circle(at(2.35), y+3.95, 0.35, skinShade)
		f.line(at(1.4), y+4.65, at(2.1), y+4.55, 0.3, mix(skinShade, eyeColour, 0.35)) // mouth
	}
	// The hard hat: a dome shaded underneath, a ridge, a shine and a brim
	// peaked toward facing.
	f.ellipse(cx, y+1.4, 3.2, 2, hatShade)
	f.ellipse(cx-0.3, y+1.1, 2.9, 1.7, hatYellow)
	f.rect(cx-0.4, y-0.55, 0.8, 2.4, mix(hatYellow, white, 0.25))
	f.rect(cx+0.2, y-0.4, 0.2, 2.2, hatShade)
	f.ellipse(cx-1.5, y+0.4, 1, 0.45, mix(hatYellow, white, 0.6))
	l, r := cx-3.4, cx+3.4
	if !back {
		l, r = min(at(-2.8), at(4.2)), max(at(-2.8), at(4.2))
	}
	f.roundRect(l, y+1.7, r-l, 0.95, 0.45, hatYellow)
	f.rect(l+0.3, y+2.3, r-l-0.6, 0.35, hatShade)
	f.rect(l+0.4, y+1.75, r-l-0.8, 0.2, mix(hatYellow, white, 0.4))
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
	case floorCard:
		p.drawCardText(ctx, f)
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

// drawFallen draws the worker flat on his back after the death spin, in
// the hero's own colours and shading: head toward where he faced, the hard
// hat knocked off beside it, one arm flung back, boots up, and stars
// circling his head.
func (p *Play) drawFallen(f field, cx, y, facing float32) {
	d := facing // the side the head lies on
	at := func(dx float32) float32 { return cx + dx*d }
	// rect draws a box given by its edges along the body, either way round.
	rect := func(dx0, dx1, y0, h float32, c color.Color) {
		x0, x1 := at(dx0), at(dx1)
		f.rect(min(x0, x1), y0, abs32(x1-x0), h, c)
	}
	f.ellipse(cx, y+16, 10, 1.2, shadowColour)
	// Legs along the girder, the far one darker, and boots toes up.
	rect(-1, -8, y+13.4, 1.8, denimDark)
	rect(-1, -8, y+11.8, 2, denim)
	rect(-1, -8, y+11.8, 0.5, mix(denim, rgb(0xffffff), 0.25))
	rect(-8, -10.4, y+10.6, 4.8, boot)
	rect(-8, -8.8, y+10.6, 4.8, mix(boot, rgb(0xffffff), 0.15))
	rect(-8, -10.4, y+15, 0.6, mix(boot, eyeColour, 0.5))
	// The other arm along his side, under the vest.
	f.line(at(1), y+15, at(-3), y+15.3, 1.5, skin)
	f.circle(at(-3.4), y+15.3, 0.8, skin)
	// The torso in the vest, shaded on its lower side, stripes across it.
	vx := min(at(5), at(-2))
	f.roundRect(vx, y+11, 7, 4.6, 1.4, vestOrange)
	f.rect(vx, y+14.2, 7, 1.2, mix(vestOrange, eyeColour, 0.25))
	rect(2.6, 3.5, y+11, 4.6, vestStripe)
	rect(0.4, 1.3, y+11, 4.6, vestStripe)
	// The near arm flung back over his head.
	f.line(at(4), y+11.8, at(7.5), y+9.2, 1.6, skin)
	f.circle(at(7.8), y+8.9, 0.9, skin)
	// The head, cheek shaded, eyes crossed out.
	hx := at(7.4)
	f.rect(min(at(5), at(6)), y+12.6, 1, 1.4, skinShade)
	f.circle(hx, y+13.2, 2.5, skin)
	f.ellipse(hx, y+14.4, 2.1, 1, skinShade)
	f.ellipse(hx, y+11.4, 2.2, 0.9, rgb(0x5a3a22)) // hair
	for _, ex := range []float32{-0.9, 0.9} {
		ex += hx
		f.line(ex-0.5, y+12.4, ex+0.5, y+13.4, 0.35, eyeColour)
		f.line(ex+0.5, y+12.4, ex-0.5, y+13.4, 0.35, eyeColour)
	}
	f.ellipse(hx, y+14.3, 0.6, 0.3, rgb(0x9a3040))
	// The hard hat lying upturned beside him, brim and shine like his own.
	tx := at(13.4)
	f.ellipse(tx, y+14.4, 3.2, 1.7, hatYellow)
	f.rect(tx-3.4, y+14.2, 6.8, 0.9, hatShade)
	f.ellipse(tx-1, y+15.1, 1.2, 0.4, mix(hatYellow, rgb(0xffffff), 0.5))
	// Three stars circling above his head.
	spin := float32(p.timer) * 0.08
	for i := range 3 {
		a := spin + float32(i)*2.094
		sx, sy := hx+3.4*cos(a), y+9+1.2*sin(a)
		r := float32(1.1)
		f.poly(hazardYellow, sx, sy-r, sx+0.35*r, sy-0.35*r, sx+r, sy, sx+0.35*r, sy+0.35*r,
			sx, sy+r, sx-0.35*r, sy+0.35*r, sx-r, sy, sx-0.35*r, sy-0.35*r)
		f.circle(sx, sy, 0.3, rgb(0xffffff))
	}
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
