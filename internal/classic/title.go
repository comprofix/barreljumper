package classic

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/barreljumper/internal/scene"
)

// The title art (DESIGN section 6, Title, attract and music): a logo of
// riveted steel blocks over a construction site at dusk, with the thrower
// and the captive on a high girder and the worker below.

// titleGlyphs are the logo's letters on a 5 × 7 grid, a block per #.
var titleGlyphs = map[rune][7]string{
	'B': {"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
	'A': {".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'R': {"####.", "#...#", "#...#", "####.", "#.#..", "#..#.", "#...#"},
	'E': {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
	'L': {"#....", "#....", "#....", "#....", "#....", "#....", "#####"},
	'J': {"..###", "...#.", "...#.", "...#.", "#..#.", "#..#.", ".##.."},
	'U': {"#...#", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'M': {"#...#", "##.##", "#.#.#", "#.#.#", "#...#", "#...#", "#...#"},
	'P': {"####.", "#...#", "#...#", "####.", "#....", "#....", "#...."},
}

// The title is laid out on a 320 × 180 grid scaled to the screen.
const (
	titleWidth  = 320
	titleHeight = 180
	titleBlock  = 4 // a logo block's size
)

// DrawTitle draws the title art; the shared title scene adds the prompt.
func DrawTitle(ctx *scene.Context, dst *ebiten.Image) {
	size := dst.Bounds().Size()
	scale := min(float32(size.X)/titleWidth, float32(size.Y)/titleHeight)
	f := field{dst, (float32(size.X) - titleWidth*scale) / 2, (float32(size.Y) - titleHeight*scale) / 2, scale}
	dst.Fill(skyTop)
	drawTitleSite(f, ctx.Tick)
	drawLogo(f, "BARREL", 14, ctx.Tick)
	drawLogo(f, "JUMPER", 14+8*titleBlock+2, ctx.Tick-12)
}

// drawTitleSite draws the sky, the skyline and crane, the girders, and
// the characters.
func drawTitleSite(f field, tick int64) {
	const bands = 36
	for i := range bands {
		t := float32(i) / (bands - 1)
		f.rect(-40, titleHeight*float32(i)/bands, titleWidth+80, titleHeight/bands+0.5, mix(skyTop, skyBottom, t))
	}
	far := []float32{-40, 140, -10, 124, 14, 132, 36, 112, 58, 130, 80, 118, 104, 136, 126, 110, 150, 128,
		176, 106, 200, 126, 224, 116, 246, 132, 270, 112, 296, 128, 318, 118, 360, 140}
	for i := 0; i+3 < len(far); i += 2 {
		f.rect(far[i], far[i+1], far[i+2]-far[i]+0.5, titleHeight-far[i+1], skylineFar)
		for wy := far[i+1] + 6; wy < titleHeight-4; wy += 9 {
			for wx := far[i] + 3; wx < far[i+2]-3; wx += 6 {
				if int(wx*7+wy*3)%5 == 0 {
					f.rect(wx, wy, 2, 3, mix(skylineFar, windowLight, 0.55))
				}
			}
		}
	}
	// A tower crane over the site, its hook on a cable.
	f.rect(26, 24, 5, 150, craneColour)
	for y := float32(28); y < 170; y += 8 {
		f.line(26, y, 31, y+8, 0.6, skyBottom)
	}
	f.rect(4, 22, 150, 3, craneColour)
	f.rect(26, 12, 5, 10, craneColour)
	f.line(28, 12, 140, 22, 0.6, craneColour)
	f.line(28, 12, 6, 22, 0.6, craneColour)
	f.rect(6, 25, 12, 9, craneColour)
	hookY := 58 + 4*sin(float32(tick)/90)
	f.line(54, 25, 54, hookY, 0.5, cable)
	drawHook(f, 54, hookY)

	// The girders: the thrower's, a lower one with a ladder, and the ground.
	drawBeam(f, 214, 108, 120)
	drawBeam(f, 40, 142, 200)
	drawBeam(f, -40, 172, titleWidth+80)
	drawLadder(f, ladder{x: 222, top: 108, height: 34}, false)
	drawLadder(f, ladder{x: 70, top: 142, height: 30}, false)

	// The thrower beats his chest; the captive calls for help beside him.
	pose := throwerTauntA
	if tick/24%2 == 1 {
		pose = throwerTauntB
	}
	drawThrower(f, thrower{x: 240, y: 76, pose: pose})
	drawCaptive(f, 290, 91, tick)
	// The worker on the lower girder hops each barrel that rolls by.
	roll := int(tick) % 190
	bx := 226 - roll
	drawWorkerHop(field{f.dst, f.left + titleWorkerX*f.scale, f.top + 126*f.scale, f.scale}, titleHop(bx))
	drawBarrel(f, barrel{x: bx, y: 132, frame: roll / 3})
}

// titleWorkerX is the centre of the title's worker.
const titleWorkerX = 112

// titleHop is how high the title's worker is off the girder with the
// barrel at bx: he takes off as it closes in and lands once it is past.
func titleHop(bx int) float32 {
	return hopLift(float32(bx+7-titleWorkerX), 26, 16)
}

// hopLift is a hop's height with the barrel's centre d from the worker's:
// he takes off when it is reach away, peaks at height over it and lands
// once it is reach past.
func hopLift(d, reach, height float32) float32 {
	if d > reach || d < -reach {
		return 0
	}
	return height * sin((reach-d)/(2*reach)*3.1416)
}

// drawLogo draws word in steel blocks centred across the title, with its
// top at y. A glint runs across the blocks every few seconds.
func drawLogo(f field, word string, y float32, tick int64) {
	drawLogoAt(f, word, titleWidth/2, y, titleBlock, tick)
}

// drawLogoAt draws word in steel blocks of size b, centred on cx with its
// top at y. The glint keeps the title's timing at any size.
func drawLogoAt(f field, word string, cx, y, b float32, tick int64) {
	const gap = 1.5 // between letters, in blocks
	n := float32(len(word))
	width := (n*5 + (n-1)*gap) * b
	left := cx - width/2
	k := b / titleBlock
	glint := (float32(tick%300)*1.2 - 40) * k
	for i, r := range word {
		g := titleGlyphs[r]
		lx := left + float32(i)*(5+gap)*b
		for row := range 7 {
			for col := range 5 {
				if g[row][col] != '#' {
					continue
				}
				x, by := lx+float32(col)*b, y+float32(row)*b
				drawLogoBlock(f, x, by, b, clamp01(1-abs32(x-left-glint)/(14*k)))
			}
		}
	}
}

// drawLogoBlock draws one bevelled steel block of size b with a bolt, lit
// by shine from 0 to 1.
func drawLogoBlock(f field, x, y, b, shine float32) {
	f.rect(x+0.2*b, y+0.2*b, b, b, rgb(0x0c0c14))
	f.rect(x, y, b, b, steelEdge)
	f.rect(x, y, b-0.1*b, b-0.1*b, mix(steel, rgb(0xfff0d0), shine*0.6))
	f.rect(x, y, b-0.1*b, 0.15*b, mix(steelLight, rgb(0xffffff), shine))
	f.rect(x, y, 0.15*b, b-0.1*b, mix(steelLight, rgb(0xffffff), shine*0.6))
	f.rect(x+0.1*b, y+b-0.25*b, b-0.2*b, 0.15*b, steelDark)
	f.circle(x+b/2, y+b/2, 0.1125*b, boltColour)
}
