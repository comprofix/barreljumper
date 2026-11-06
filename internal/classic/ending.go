package classic

import (
	"fmt"
	"math"

	"github.com/comprofix/barreljumper/internal/gfx"
	"github.com/comprofix/barreljumper/internal/scene"
)

// The opening, floor card and screen endings (DESIGN section 6, Screen
// endings, floor card and opening). Their lengths follow the reference
// (section 4, Screen endings and the height screen); the scenes are the
// rebuild's own. Times are counter ticks from the phase's start.

// Carry-off: the crane hook comes down to the thrower, he pulls the
// captive to him, and the crane hoists them both off the top.
const (
	hookTicks = 60
	grabTicks = 50
)

// Opening: the crane lowers the Barrels rows into place above the ground,
// then hoists the thrower and the captive to the top. He carries her to
// her platform and sets her on it, she walks to her place, and he walks
// back to his and beats his chest.
const (
	openRowEvery = 70
	openRowTicks = 90
	openRows     = 7   // the six Barrels rows, then the captive's platform
	openLift     = 30  // the empty hook going back up after a landing
	openLadders  = 460 // ladders, drum and hammers appear
	openHook     = 480 // the hook comes down to the thrower
	openHoist    = 520 // and lifts him and the captive
	openTop      = 680 // they reach the top
	openCarried  = 740 // he has carried her to the platform's edge
	openPlaced   = 770 // and set her on it
	openHerPlace = 810 // she has walked to her place
	openBack     = 840 // he is back at his own
)

// The thrower carries the captive this far right along the top, to the
// platform's left edge, and holds her at his side.
const (
	openCarryX         = 70
	carryAtX, carryAtY = 36, 10
)

// Rivets ending: the inner floors drop into a pile, the thrower's own floor
// last; he lands on the pile stunned; the top floor is lowered to meet the
// worker, who stands with the captive under a heart.
const (
	rivetsHisFloor = 220
	rivetsLower    = 500
	rivetsLowered  = 590
	rivetsHeart    = 600
	rivetsDrop     = 42 // how far the top floor comes down, to y 66
	fallGravity    = 0.08
)

// The opening and Rivets thrower stands at the bottom of Barrels.
var openThrower = point{60, 162}

func clamp01(v float32) float32 { return max(0, min(1, v)) }

func lerp(a, b, t float32) float32 { return a + (b-a)*t }

// smooth eases t in and out.
func smooth(t float32) float32 {
	t = clamp01(t)
	return t * t * (3 - 2*t)
}

// span is how far t is through [from, from+length], clamped.
func span(t, from, length int) float32 {
	return clamp01(float32(t-from) / float32(length))
}

// openRowStart is when row r, above the ground row, starts down.
func openRowStart(r int) int { return (r - 1) * openRowEvery }

func openRowLanded(r int) int { return openRowStart(r) + openRowTicks }

func openingSounds(t int) sound {
	var s sound
	for r := 1; r < openRows; r++ {
		if t == openRowStart(r) {
			s |= soundWinch
		}
		if t == openRowLanded(r) {
			s |= soundClank
		}
	}
	if t == openHoist {
		s |= soundWinch
	}
	if t == openPlaced {
		s |= soundLand
	}
	return s
}

func rivetsSounds(t int) sound {
	switch t {
	case rivetsHisFloor:
		return soundRumble
	case rivetsHisFloor + fallTicks(32, 162-32):
		return soundThud
	case rivetsLower:
		return soundWinch
	case rivetsLowered:
		return soundClank
	case rivetsHeart:
		return soundPoints
	}
	return 0
}

// fallTicks is how long a fall of distance takes under fallGravity.
func fallTicks(_ int, distance int) int {
	return int(math.Ceil(math.Sqrt(2 * float64(distance) / fallGravity)))
}

// fall is a drop from y to rest after t ticks.
func fall(from, rest float32, t int) float32 {
	if t <= 0 {
		return from
	}
	return min(from+0.5*fallGravity*float32(t*t), rest)
}

// captiveAt is where the captive stands on each family's screen.
func captiveAt(family int) (float32, float32) {
	switch family {
	case rivetsFamily:
		return 126, 7
	case elevatorsFamily:
		return 146, 7
	case conveyorFamily:
		return 146, 5
	}
	return 126, 7 // her shoes end 17 below, on the platform's top at 24
}

// carried is the thrower, captive and hook during the carry-off.
func (p *Play) carried() (t thrower, gx, gy, hx, hy float32, grabbed bool) {
	t = p.stage.thrower
	t.y += p.stage.throwerSink()
	t.holding, t.taunting = false, false
	t.pose = throwerTauntB
	cx, cy := captiveAt(p.stage.family)
	handX, handY := float32(t.x)+37, float32(t.y)+1.5
	tick := p.timer
	hy = lerp(-12, handY, smooth(span(tick, 0, hookTicks)))
	g := smooth(span(tick, hookTicks, grabTicks))
	gx, gy = lerp(cx, float32(t.x)+38, g), lerp(cy, float32(t.y)+10, g)
	grabbed = tick >= hookTicks+grabTicks
	if grabbed {
		up := span(tick, hookTicks+grabTicks, carryTicks-hookTicks-grabTicks)
		lift := up * up * (handY + 70)
		t.y -= int(lift)
		gy -= lift
		hy = handY - lift
	}
	return t, gx, gy, handX, hy, grabbed
}

// drawHook draws the crane hook on its cable from the top of the field.
func drawHook(f field, x, y float32) {
	f.line(x, -20, x, y-4, 0.8, cable)
	f.roundRect(x-2.2, y-5.5, 4.4, 3, 0.8, hazardYellow)
	f.rect(x-2.2, y-4.4, 4.4, 0.7, hazardBlack)
	f.rect(x-0.4, y-2.5, 0.8, 2.2, hammerSteel)
	f.line(x, y-0.4, x+1.4, y+1.2, 0.8, hammerSteel)
	f.line(x+1.4, y+1.2, x-0.6, y+2.2, 0.8, hammerSteel)
	f.line(x-0.6, y+2.2, x-1.2, y+1.2, 0.8, hammerSteel)
}

var heartColour = rgb(0xff4f7a)

// drawHeart draws a heart centred on x with its top at y, split when broken.
func drawHeart(f field, x, y float32, broken bool) {
	c, light := heartColour, mix(heartColour, rgb(0xffffff), 0.45)
	if !broken {
		f.circle(x-1.5, y+1.6, 1.8, c)
		f.circle(x+1.5, y+1.6, 1.8, c)
		f.poly(c, x-3.2, y+2, x+3.2, y+2, x, y+6)
		f.circle(x-1.8, y+1.1, 0.6, light)
		return
	}
	c, light = mix(c, eyeColour, 0.3), mix(c, rgb(0xffffff), 0.2)
	l, r := x-0.8, x+0.8
	f.circle(l-1.5, y+1.8, 1.8, c)
	f.poly(c, l-3.2, y+2.2, l, y+1.4, l+0.6, y+3, l, y+4.4, l, y+6.2)
	f.circle(r+1.5, y+1.4, 1.8, c)
	f.poly(c, r+3.2, y+1.8, r, y+1, r+0.6, y+2.6, r, y+4, r, y+5.8)
	f.circle(l-1.8, y+1.3, 0.6, light)
}

// drawStars circles three stars about x, y.
func drawStars(f field, x, y float32, tick int) {
	spin := float32(tick) * 0.08
	for i := range 3 {
		a := spin + float32(i)*2.094
		sx, sy := x+4.5*cos(a), y+1.4*sin(a)
		r := float32(1.3)
		f.poly(hazardYellow, sx, sy-r, sx+0.35*r, sy-0.35*r, sx+r, sy, sx+0.35*r, sy+0.35*r,
			sx, sy+r, sx-0.35*r, sy+0.35*r, sx-r, sy, sx-0.35*r, sy-0.35*r)
		f.circle(sx, sy, 0.3, rgb(0xffffff))
	}
}

// barrelsRow is the row, bottom first, of Barrels girder i.
func barrelsRow(i int) int {
	if i < 13 {
		return 0
	}
	return min((i-13)/12+1, 5)
}

// drawOpening draws the opening scene on the Barrels screen.
func (p *Play) drawOpening(f field, tick int64) {
	s, t := p.stage, p.timer
	if t >= openLadders {
		for _, l := range s.layout.reserved {
			drawLadder(f, l, true)
		}
		for _, l := range s.ladders {
			drawLadder(f, l, false)
		}
	}
	// Each piece comes down on the crane hook, slung from two shackles on
	// the beam; once it is down the empty hook goes back up.
	for r := range openRows {
		start := openRowStart(r)
		if t < start {
			continue
		}
		cells := openPiece(s, r)
		drop := float32(0)
		if r > 0 {
			drop = (1 - smooth(span(t, start, openRowTicks))) * -210
		}
		for _, g := range cells {
			drawBeam(f, float32(g.left), float32(g.y)+drop, float32(g.right-g.left))
		}
		if r == 0 {
			continue
		}
		hx, hy, a, b := rig(cells, drop)
		switch landed := openRowLanded(r); {
		case t < landed:
			f.line(hx, hy+1.6, a.x, a.y, 0.6, cable)
			f.line(hx, hy+1.6, b.x, b.y, 0.6, cable)
			drawShackle(f, a)
			drawShackle(f, b)
			drawHook(f, hx, hy)
		case t < landed+openLift:
			drawHook(f, hx, lerp(hy, -12, smooth(span(t, landed, openLift))))
		}
	}
	if t >= openLadders {
		for _, h := range s.pickups {
			drawPickup(f, float32(h.x), float32(h.y))
		}
		drawDrum(f, s.oil, tick)
	}
	// The thrower and captive, hoisted to the top.
	th := thrower{x: openThrower.x, y: openThrower.y, pose: throwerReady}
	if (t/40)%2 == 1 && t < openHook {
		th.pose = throwerTauntA
	}
	gx, gy := float32(th.x)+42, float32(th.y)+15
	handX := float32(th.x) + 37
	hookY := float32(-12)
	if t >= openHook {
		th.pose = throwerTauntB
		hookY = lerp(-12, float32(th.y)+1.5, smooth(span(t, openHook, openHoist-openHook)))
		gx, gy = float32(th.x)+carryAtX, float32(th.y)+carryAtY
	}
	top := s.layout.thrower.y + s.throwerSink()
	if t >= openHoist {
		up := smooth(span(t, openHoist, openTop-openHoist))
		th.x = int(lerp(float32(openThrower.x), float32(s.layout.thrower.x), up))
		th.y = int(lerp(float32(openThrower.y), float32(top), up))
		handX = float32(th.x) + 37
		hookY = float32(th.y) + 1.5
		gx, gy = float32(th.x)+carryAtX, float32(th.y)+carryAtY
	}
	if t >= openTop {
		// The hook goes back up; he carries her along the top under his arm.
		hookY = lerp(float32(th.y)+1.5, -12, smooth(span(t, openTop, 60)))
		home := float32(s.layout.thrower.x)
		x := lerp(home, openCarryX, smooth(span(t, openTop, openCarried-openTop)))
		// He walks, lifting a leg each step; the other stays on the girder.
		step := (t / 8) % 2
		th.pose = throwerPickUp
		if t >= openPlaced {
			// Then walks back to his place.
			x = lerp(openCarryX, home, smooth(span(t, openPlaced+5, openBack-openPlaced-5)))
			th.pose = throwerReady
		}
		th.x = int(x)
		if walking := (t < openCarried || t > openPlaced+5) && t < openBack; walking {
			th.stride = 1 - 2*step
		}
		gx, gy = x+carryAtX, float32(th.y)+carryAtY
		if t >= openCarried {
			// He lifts her onto the platform's edge, and she walks to her place.
			cx, cy := captiveAt(barrelsFamily)
			edge := float32(barrelsGoal.left)
			lift := smooth(span(t, openCarried, openPlaced-openCarried))
			gx = lerp(openCarryX+carryAtX, edge, lift)
			gy = lerp(float32(top)+carryAtY, cy, lift)
			if t >= openPlaced {
				gx = lerp(edge, cx, span(t, openPlaced, openHerPlace-openPlaced))
				gy = cy
			}
		}
	}
	if t >= openBack {
		th.pose = throwerTauntA + (t/12)%2
	}
	// While he holds her she is behind him, his fist at her waist.
	held := t >= openHook && t < openPlaced
	if held {
		drawCaptive(f, gx, gy, tick)
	}
	drawThrower(f, th)
	if !held {
		// She walks the platform to her place, a leg lifting each step.
		stride := 0
		if t >= openPlaced && t < openHerPlace {
			stride = 1 - 2*((t/6)%2)
		}
		drawCaptiveStep(f, gx, gy, tick, stride)
	}
	if hookY > -12 {
		drawHook(f, handX, hookY)
	}
}

// openPiece is the girders the crane lowers as piece r of the opening:
// a Barrels row, bottom first, then the captive's platform.
func openPiece(s *stage, r int) []girder {
	if r == openRows-1 {
		return []girder{{barrelsGoal.left, barrelsGoal.left + 50, barrelsGoal.y}}
	}
	var cells []girder
	for i, g := range s.girders {
		if barrelsRow(i) == r {
			cells = append(cells, g)
		}
	}
	return cells
}

type spot struct{ x, y float32 }

// rig is where the crane holds a piece lowered by drop: the hook above
// its middle and the two shackles a quarter in from each end, on the
// beam's top where it runs under them.
func rig(cells []girder, drop float32) (hx, hy float32, a, b spot) {
	left, right := float32(cells[0].left), float32(cells[len(cells)-1].right)
	top := func(x float32) float32 {
		for _, g := range cells {
			if x < float32(g.right) {
				return float32(g.y) + drop
			}
		}
		return float32(cells[len(cells)-1].y) + drop
	}
	ax, bx := lerp(left, right, 0.25), lerp(left, right, 0.75)
	if len(cells) == 1 {
		ax, bx = left+8, right-8
	}
	a, b = spot{ax, top(ax)}, spot{bx, top(bx)}
	hx = (ax + bx) / 2
	hy = min(a.y, b.y) - (bx-ax)*0.35
	return hx, hy, a, b
}

// drawShackle draws the shackle a sling is fixed to on a beam's top.
func drawShackle(f field, at spot) {
	f.circle(at.x, at.y-0.6, 0.9, hammerDark)
	f.circle(at.x, at.y-0.6, 0.45, hammerSteel)
	f.rect(at.x-1.2, at.y-0.2, 2.4, 0.7, hammerDark)
}

// rivetFaller reports whether girder g drops in the Rivets ending, and its
// row, bottom first.
func rivetFaller(g girder) (int, bool) {
	if g.left < 70 || g.left >= 190 {
		return 0, false
	}
	for r, y := range rivetRows {
		if g.y == y {
			return r, true
		}
	}
	return 0, false
}

// rivetsFall is when a row starts to drop.
func rivetsFall(r int) int {
	if r == len(rivetRows)-1 {
		return rivetsHisFloor
	}
	return r * 6
}

// meet places the worker beside the lowered captive for the Rivets ending.
func (s *stage) meet() {
	m := s.player
	cx, cy := captiveAt(rivetsFamily)
	m.X, m.Y, m.Facing = int(cx)-16, int(cy)+rivetsDrop-1, 1
	m.ladder, m.jumping, m.falling, m.grounded, m.Dead = nil, false, false, true, false
	s.hammer.held = false
}

// drawRivetsEnd draws the Rivets ending.
func (p *Play) drawRivetsEnd(f field, tick int64) {
	s, t := p.stage, p.timer
	for _, l := range s.ladders {
		if l.x < 70 || l.x >= 190 {
			drawLadder(f, l, false)
		}
	}
	lower := smooth(span(t, rivetsLower, rivetsLowered-rivetsLower)) * rivetsDrop
	for _, g := range s.girders {
		x, w := float32(g.left), float32(g.right-g.left)
		if r, ok := rivetFaller(g); ok {
			rest := float32(186 - 8*r)
			drawBeam(f, x, fall(float32(g.y), rest, t-rivetsFall(r)), w)
			continue
		}
		if g.y == 24 {
			drawBeam(f, x, float32(g.y)+lower, w)
			continue
		}
		drawGirder(f, g)
	}
	cx, cy := captiveAt(rivetsFamily)
	drawCaptive(f, cx, cy+lower, tick)
	// The thrower beats his chest, falls with his floor and lands stunned.
	th := s.thrower
	th.y += s.throwerSink()
	th.holding, th.taunting = false, false
	th.pose = throwerTauntA + (t/14)%2
	rest := float32(162 - 32)
	if t >= rivetsHisFloor {
		y := fall(float32(th.y), rest, t-rivetsHisFloor)
		th.y = int(y)
		th.pose = throwerTauntA + (t/4)%2
		if y >= rest {
			th.pose = throwerReady
		}
	}
	drawThrower(f, th)
	if float32(th.y) >= rest {
		drawStars(f, float32(th.x)+20, float32(th.y)-1, t)
	}
	if t >= rivetsLowered {
		p.drawPlayer(f)
	}
	if t >= rivetsHeart {
		drawHeart(f, cx-1, cy+lower-8, false)
	}
}

// drawFloorCard draws the card before a screen's play: the building with a
// floor for each screen of the level, the floors climbed lit, the worker on
// the current one, and the thrower with the captive on the roof.
func (p *Play) drawFloorCard(f field) {
	n := len(classicSchedule[p.level%5])
	h := min(float32(30), 130/float32(n))
	base := float32(186)
	left, width := float32(80), float32(100)
	drawBeam(f, 40, base, 180)
	top := base - h*float32(n)
	for k := range n {
		y := base - h*float32(k+1)
		// Window panels between the columns, lit once climbed.
		glass := mix(skylineNear, skyTop, 0.3)
		switch {
		case k < p.index:
			glass = mix(windowLight, skylineNear, 0.55)
		case k == p.index:
			glass = windowLight
		}
		for wx := left + 8; wx < left+width-8; wx += 14 {
			f.rect(wx, y+7, 10, h-9, glass)
			f.rect(wx, y+7, 10, 1, mix(glass, rgb(0xffffff), 0.3))
		}
		f.rect(left+2, y+7, 4, h-7, steelDark)
		f.rect(left+width-6, y+7, 4, h-7, steelDark)
		f.rect(left+2, y+7, 1, h-7, steel)
		f.rect(left+width-6, y+7, 1, h-7, steel)
		drawBeam(f, left, y, width)
		if k == p.index {
			// The worker on his floor, and a marker beside it.
			drawWorkerIcon(field{f.dst, f.left + (left+width/2)*f.scale, f.top + (y+7)*f.scale, f.scale * min(1, (h-8)/16)})
			if p.timer/18%2 == 0 {
				ay := y + h/2 + 3
				f.poly(hazardYellow, left+width+10, ay-3, left+width+10, ay+3, left+width+5, ay)
			}
		}
	}
	// The thrower and captive on the roof, at half size.
	roof := field{f.dst, f.left + (left+18)*f.scale, f.top + (top-16)*f.scale, f.scale * 0.5}
	drawThrower(roof, thrower{pose: throwerTauntA + p.timer/16%2})
	roof.left += 46 * roof.scale
	roof.top += 15 * roof.scale
	drawCaptive(roof, 0, 0, int64(p.timer))
}

// drawWorkerIcon draws the worker standing, facing right, with the field's
// origin at the centre of his hard hat's top.
func drawWorkerIcon(f field) {
	at := func(dx float32) float32 { return dx }
	f.ellipse(0, 16, 6, 1, shadowColour)
	drawLeg(f, -1, 10.4, 4, true)
	drawBoot(f, -0.7, 14, 1)
	drawLeg(f, 1, 10.4, 4, false)
	drawBoot(f, 1.3, 14, 1)
	f.rect(-3.9, 10, 7.8, 0.9, beltColour)
	drawVest(f, 0, 0, at, vestOrange, true)
	drawLimb(f, 3, 6, 4.4, 9.6, false)
	drawHead(f, 0, 0, at, false)
}

// drawCardText writes the floor card's lines: the player, the floor and
// its height.
func (p *Play) drawCardText(ctx *scene.Context, f field) {
	small, large, sc := ctx.Fonts.Small, ctx.Fonts.Large, ctx.Scheme
	if large == nil {
		large = small
	}
	centre := func(font gfx.Font, s string, y float32, c int) {
		x := int(f.left+130*f.scale) - font.Width(s)/2
		font.DrawShadowed(f.dst, s, x, int(f.top+y*f.scale), c, sc.Shadow)
	}
	n := len(classicSchedule[p.level%5])
	centre(large, fmt.Sprint("PLAYER ", p.current+1), 10, sc.Title)
	centre(small, fmt.Sprintf("FLOOR %d OF %d   %d M UP", p.index+1, n, 25*(p.index+1)), 30, sc.Text)
}
