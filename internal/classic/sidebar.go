package classic

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/scene"
)

// The sidebar is a site board right of the field (DESIGN section 6,
// Sidebar): a painted board in a riveted steel frame, with the logo, a
// window on the thrower and the captive, inset plates for the scores, the
// bonus and the level, hard hats for the lives, and a worker hopping
// barrels at the foot. It is drawing only.

// The board's span inside its frame, in field units.
const (
	boardLeft  = 264
	boardRight = 316
	boardTop   = 4
	boardFoot  = 196
)

// The windows and plates, in field units: left edge, width and height.
const (
	paneLeft  = 267
	paneWidth = 46
	plateLeft = 266
	plateW    = 48
	plateH    = 15
)

// The panes' tops and the first plate's: the portrait window, the
// plates, then the vignette window at the foot.
const (
	portraitTop = 35
	portraitH   = 35
	platesTop   = 75
	plateStep   = 16.5
	vignetteTop = 160
	vignetteH   = 33
)

var (
	boardTopColour  = rgb(0x1e2236)
	boardFootColour = rgb(0x12141f)
	boardSeam       = rgb(0x0e1018)
	plateFace       = rgb(0x080a10)
	plateRim        = rgb(0x3a3f58)
)

// drawSidebar draws the board and everything on it, then its text.
func (p *Play) drawSidebar(ctx *scene.Context, f field) {
	tick := ctx.Tick
	drawBoard(f)
	drawLogoAt(f, "BARREL", 290, 7, 1.2, tick)
	drawLogoAt(f, "JUMPER", 290, 17.4, 1.2, tick-12)
	drawHazardBand(f, 28.4)
	drawPortrait(f, tick)
	rows := p.sidebarRows(bestScore(ctx, p.player.skill.ID))
	for i := range rows {
		drawPlate(f, plateTop(i), plateH)
	}
	drawHats(f, plateTop(len(rows)-1)+8.6, p.reserves)
	drawVignette(f, tick)
	p.drawSidebarText(ctx, f, rows)
}

// sidebarRow is one plate's label and value.
type sidebarRow struct {
	label, value string
	labelColour  string // a palette name, or "" for the scheme's dim
	valueColour  string // a palette name, or "" for the scheme's text
	hats         bool   // the value line holds the lives as hard hats
}

// bestScore is the top of skill's high score table, or 0.
func bestScore(ctx *scene.Context, skill string) int {
	if ctx.Scores == nil {
		return 0
	}
	if t := ctx.Scores.Table(skill); len(t) > 0 {
		return t[0].Score
	}
	return 0
}

// sidebarRows lists the plates top to bottom: the high score, which rises
// with any player's score that passes it, each player's score, the bonus,
// and the level over the lives.
func (p *Play) sidebarRows(best int) []sidebarRow {
	for _, pl := range p.players {
		best = max(best, pl.score)
	}
	rows := []sidebarRow{{label: "HI SCORE", value: fmt.Sprint(best)}}
	if len(p.players) == 1 {
		rows = append(rows, sidebarRow{label: "SCORE", value: fmt.Sprint(p.score)})
	} else {
		for i, pl := range p.players {
			r := sidebarRow{label: fmt.Sprint("PLAYER ", i+1), value: fmt.Sprint(pl.score)}
			if i == p.current {
				r.labelColour = "Gold"
			}
			rows = append(rows, r)
		}
	}
	bonus := sidebarRow{label: "BONUS", value: fmt.Sprint(p.stage.bonus), valueColour: "Amber"}
	if p.stage.bonus < lowBonus {
		bonus.valueColour = "Red"
	}
	return append(rows, bonus, sidebarRow{label: fmt.Sprint("LEVEL ", p.level+1), hats: true})
}

func plateTop(i int) float32 { return platesTop + float32(i)*plateStep }

// drawBoard paints the board with its plank seams, frames it with steel
// columns and beams, and frames the two windows.
func drawBoard(f field) {
	const bands = 24
	for i := range bands {
		t := float32(i) / (bands - 1)
		y := boardTop + (boardFoot-boardTop)*float32(i)/bands
		f.rect(boardLeft, y, boardRight-boardLeft, (boardFoot-boardTop)/bands+0.5, mix(boardTopColour, boardFootColour, t))
	}
	for x := float32(boardLeft + 13); x < boardRight; x += 13 {
		f.rect(x, boardTop, 0.4, boardFoot-boardTop, boardSeam)
		f.rect(x+0.4, boardTop, 0.3, boardFoot-boardTop, mix(boardTopColour, rgb(0xffffff), 0.05))
	}
	// A shadow under the top beam and inside the columns, for depth.
	f.rect(boardLeft, boardTop, boardRight-boardLeft, 2, color.RGBA{0, 0, 0, 90})
	f.rect(boardLeft, boardTop, 1.2, boardFoot-boardTop, color.RGBA{0, 0, 0, 70})
	f.rect(boardRight-1.2, boardTop, 1.2, boardFoot-boardTop, color.RGBA{0, 0, 0, 50})
	drawColumn(f, 260)
	drawColumn(f, 316)
	drawBeam(f, 260, -3, 60)
	drawBeam(f, 260, 193, 60)
	drawPaneFrame(f, portraitTop, portraitH)
	drawPaneFrame(f, vignetteTop, vignetteH)
	drawPlate(f, 5.4, 21.4) // the logo's
}

// drawColumn draws a steel I-beam column four units wide down the side of
// the board, lit from the left, with rivets down both flanges.
func drawColumn(f field, x float32) {
	f.rect(x, 0, 4, 200, steelDark)
	f.rect(x, 0, 1.2, 200, steel)
	f.rect(x, 0, 0.5, 200, steelLight)
	f.rect(x+2.8, 0, 1.2, 200, steel)
	f.rect(x+3.6, 0, 0.4, 200, steelEdge)
	f.rect(x+1.6, 0, 0.8, 200, mix(steelDark, steelEdge, 0.5))
	for y := float32(6); y < 196; y += 10 {
		f.circle(x+0.7, y, 0.4, boltColour)
		f.circle(x+3.3, y, 0.4, boltColour)
	}
}

// drawHazardBand draws a band of hazard stripes across the board.
func drawHazardBand(f field, y float32) {
	c := clip(f, boardLeft, y, boardRight, y+2.4)
	c.rect(boardLeft, y, boardRight-boardLeft, 2.4, hazardBlack)
	for x := float32(boardLeft - 4); x < boardRight; x += 4 {
		c.poly(hazardYellow, x, y+2.4, x+2, y+2.4, x+4.4, y, x+2.4, y)
	}
	f.rect(boardLeft, y+2.4, boardRight-boardLeft, 0.5, color.RGBA{0, 0, 0, 110})
}

// clip returns a field that draws only inside x0, y0 to x1, y1: the same
// mapping, clipped.
func clip(f field, x0, y0, x1, y1 float32) field {
	r := image.Rect(int(f.left+x0*f.scale), int(f.top+y0*f.scale),
		int(f.left+x1*f.scale+0.999), int(f.top+y1*f.scale+0.999))
	return field{f.dst.SubImage(r).(*ebiten.Image), f.left, f.top, f.scale}
}

// drawPaneFrame draws a steel window frame round a pane, bolted at the
// corners.
func drawPaneFrame(f field, top, h float32) {
	l, r, b := float32(paneLeft-1.6), float32(paneLeft+paneWidth+1.6), top+h+1.6
	f.rect(l+0.6, top-1, r-l, b-top+1.6+0.6, color.RGBA{0, 0, 0, 120})
	f.rect(l, top-1.6, r-l, b-top+1.6, steelDark)
	f.rect(l, top-1.6, r-l, 0.5, steelLight)
	f.rect(l, top-1.6, 0.5, b-top+1.6, steel)
	f.rect(l, b-0.5, r-l, 0.5, steelEdge)
	f.rect(r-0.5, top-1.6, 0.5, b-top+1.6, steelEdge)
	for _, c := range [][2]float32{{l + 0.8, top - 0.8}, {r - 0.8, top - 0.8}, {l + 0.8, b - 0.8}, {r - 0.8, b - 0.8}} {
		f.circle(c[0], c[1], 0.45, boltColour)
	}
}

// drawPlate draws a dark display plate h tall sunk into the board:
// shadowed along its top, lit along its foot, with a screw at each end.
func drawPlate(f field, y, h float32) {
	f.rect(plateLeft-0.6, y-0.6, plateW+1.2, h+1.2, plateRim)
	f.rect(plateLeft, y, plateW, h, plateFace)
	f.rect(plateLeft, y, plateW, 0.8, color.RGBA{0, 0, 0, 200})
	f.rect(plateLeft, y+h-0.4, plateW, 0.4, mix(plateFace, plateRim, 0.6))
	for _, x := range []float32{plateLeft + 1.4, plateLeft + plateW - 1.4} {
		f.circle(x, y+1.6, 0.5, mix(plateRim, rgb(0xffffff), 0.3))
		f.line(x-0.35, y+1.6, x+0.35, y+1.6, 0.2, plateFace)
	}
}

// pane returns a field that draws only inside the pane from top, h tall:
// the same mapping, clipped to the pane.
func pane(f field, top, h float32) field {
	return clip(f, paneLeft, top, paneLeft+paneWidth, top+h)
}

// drawPaneSky fills a pane with the dusk sky and a strip of distant
// buildings with lit windows.
func drawPaneSky(f field, top, h float32) {
	const bands = 10
	for i := range bands {
		t := float32(i) / (bands - 1)
		f.rect(paneLeft, top+h*float32(i)/bands, paneWidth, h/bands+0.5, mix(skyTop, skyBottom, t))
	}
	roof := []float32{0, 0.62, 7, 0.5, 13, 0.7, 20, 0.45, 28, 0.6, 35, 0.52, 41, 0.66, 46}
	for i := 0; i+2 < len(roof); i += 2 {
		x0, ry, x1 := paneLeft+roof[i], top+h*roof[i+1], paneLeft+roof[i+2]
		f.rect(x0, ry, x1-x0+0.3, top+h-ry, skylineFar)
		for wy := ry + 2; wy < top+h-2; wy += 3.4 {
			for wx := x0 + 1.2; wx < x1-1; wx += 2.6 {
				if int(wx*7+wy*3)%5 == 0 {
					f.rect(wx, wy, 0.8, 1.2, mix(skylineFar, windowLight, 0.55))
				}
			}
		}
	}
}

// drawGlass lays a faint diagonal glint over a pane.
func drawGlass(f field, top, h float32) {
	glint := color.RGBA{14, 14, 18, 14} // premultiplied
	f.poly(glint, paneLeft+6, top, paneLeft+14, top, paneLeft+2, top+h, paneLeft-6, top+h)
	f.poly(glint, paneLeft+17, top, paneLeft+20, top, paneLeft+8, top+h, paneLeft+5, top+h)
}

// drawPortrait shows the thrower beating his chest on his girder with the
// captive beside him, at 70% of their play size.
func drawPortrait(f field, tick int64) {
	w := pane(f, portraitTop, portraitH)
	drawPaneSky(w, portraitTop, portraitH)
	const k = 0.7
	ox, oy := float32(paneLeft+5), float32(portraitTop+0.5)
	s := field{w.dst, f.left + ox*f.scale, f.top + oy*f.scale, f.scale * k}
	drawBeam(s, -10, 40, 90)
	pose := throwerTauntA
	if tick/24%2 == 1 {
		pose = throwerTauntB
	}
	drawThrower(s, thrower{x: 0, y: 8, pose: pose})
	drawCaptive(s, 46, 23, tick)
	drawGlass(w, portraitTop, portraitH)
}

// drawVignette shows the worker on a girder hopping each barrel that
// rolls by, as on the title.
func drawVignette(f field, tick int64) {
	w := pane(f, vignetteTop, vignetteH)
	drawPaneSky(w, vignetteTop, vignetteH)
	const beam, workerX = vignetteTop + 25, paneLeft + 16
	drawLadder(w, ladder{x: paneLeft + 33, top: vignetteTop - 4, height: 29}, false)
	drawBeam(w, paneLeft-4, beam, paneWidth+8)
	roll := int(tick) % 110
	bx := paneLeft + paneWidth + 4 - roll
	lift := hopLift(float32(bx+7-workerX), 14, 9)
	drawWorkerHop(field{w.dst, w.left + workerX*w.scale, w.top + (beam-16)*w.scale, w.scale}, lift)
	drawBarrel(w, barrel{x: bx, y: beam - 10, frame: roll / 3})
	drawGlass(w, vignetteTop, vignetteH)
}

// drawHats draws the lives as hard hats along a plate's value line: up to
// eight, or one and a count beyond that.
func drawHats(f field, y float32, n int) {
	shown := min(n, 8)
	if n > 8 {
		shown = 1
	}
	for i := range shown {
		drawHat(f, plateLeft+4+float32(i)*5.4, y)
	}
}

// drawHat draws a small hard hat with its top left near x, y: a shaded
// dome, a ridge, a shine and a brim.
func drawHat(f field, x, y float32) {
	cx := x + 2
	f.ellipse(cx+0.3, y+2.9, 2.4, 0.5, color.RGBA{0, 0, 0, 120})
	f.ellipse(cx, y+1.8, 1.8, 1.5, hatShade)
	f.ellipse(cx-0.2, y+1.6, 1.6, 1.3, hatYellow)
	f.rect(cx-0.25, y+0.3, 0.5, 1.8, mix(hatYellow, rgb(0xffffff), 0.3))
	f.ellipse(cx-0.9, y+1.1, 0.5, 0.3, mix(hatYellow, rgb(0xffffff), 0.6))
	f.roundRect(x-0.4, y+2.2, 4.8, 0.8, 0.4, hatYellow)
	f.rect(x-0.2, y+2.7, 4.4, 0.3, hatShade)
}

// drawSidebarText writes each plate's label and value; the lives plate
// gets a count beside a single hat when there are more than eight.
func (p *Play) drawSidebarText(ctx *scene.Context, f field, rows []sidebarRow) {
	font, sc := ctx.Fonts.Small, ctx.Scheme
	if font == nil {
		return
	}
	colour := func(name string, fallback int) int {
		if i, ok := config.Project.Palette.Lookup(name); ok {
			return i
		}
		return fallback
	}
	at := func(x, y float32) (int, int) {
		return int(f.left + x*f.scale), int(f.top + y*f.scale)
	}
	for i, r := range rows {
		top := plateTop(i)
		x, y := at(plateLeft+3, top+1.4)
		font.Draw(f.dst, r.label, x, y, colour(r.labelColour, sc.Dim))
		value := r.value
		if r.hats {
			if p.reserves <= 8 {
				continue
			}
			value = fmt.Sprint("x", p.reserves)
		}
		c := colour(r.valueColour, sc.Text)
		if r.valueColour == "Red" && ctx.Tick/16%2 == 1 {
			continue // the low bonus flashes
		}
		rx, vy := at(plateLeft+plateW-3, top+7.6)
		font.Draw(f.dst, value, rx-font.Width(value), vy, c)
	}
}
