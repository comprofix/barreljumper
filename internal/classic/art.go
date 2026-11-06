package classic

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The construction-site artwork (DESIGN section 6, M5 presentation
// decisions), drawn in code as shaded vector shapes. Every sprite keeps to
// its collision rectangle, so what the player sees is what the rules test,
// and cosmetic animation never moves collision geometry.

// Palette: a dusk sky behind painted steel.
var (
	skyTop       = rgb(0x1a2036)
	skyBottom    = rgb(0x4a3a58)
	skylineFar   = rgb(0x2a2c44)
	skylineNear  = rgb(0x22243a)
	windowLight  = rgb(0xe8c070)
	craneColour  = rgb(0x2e3048)
	shadowColour = color.RGBA{0, 0, 0, 70}

	steel      = rgb(0xd0532c)
	steelLight = rgb(0xf28a52)
	steelDark  = rgb(0x93361c)
	steelEdge  = rgb(0x5e1e0e)
	boltColour = rgb(0xffc8a0)

	ladderRail  = rgb(0xe6b42e)
	ladderShade = rgb(0x9c7418)
	ladderDim   = rgb(0x6a5a3a)

	wood      = rgb(0xb06a32)
	woodLight = rgb(0xd8925a)
	woodDark  = rgb(0x6e3c18)
	hoop      = rgb(0x50505a)
	hoopLight = rgb(0x9898a4)
	blueDrum  = rgb(0x3a6cc8)
	blueLight = rgb(0x78a4f0)
	blueDark  = rgb(0x22407a)

	flameOuter = rgb(0xff6a20)
	flameMid   = rgb(0xffa830)
	flameCore  = rgb(0xfff0a0)
	scaredOut  = rgb(0x4a78ff)
	scaredCore = rgb(0xc8dcff)
	eyeColour  = rgb(0x201010)

	hatYellow  = rgb(0xffd23a)
	hatShade   = rgb(0xc89a10)
	vestOrange = rgb(0xff8a1f)
	vestStripe = rgb(0xf0f0dc)
	denim      = rgb(0x3b5a8c)
	denimDark  = rgb(0x26395a)
	skin       = rgb(0xe0a878)
	skinShade  = rgb(0xb07850)
	boot       = rgb(0x4a2c18)
	hurt       = rgb(0xd84848)

	bruteShirt = rgb(0x5a5f6e)
	bruteDark  = rgb(0x3a3e4a)
	bruteSkin  = rgb(0xc08a60)
	bruteHat   = rgb(0xd83a2a)
	bruteBib   = rgb(0x6a4a2e)

	captiveShirt = rgb(0x4ab0a0)
	captiveHair  = rgb(0x3a2418)
	blueprint    = rgb(0x5a8ad8)

	hammerSteel = rgb(0xb8bcc8)
	hammerDark  = rgb(0x6a6e7a)
	handle      = rgb(0x9a6838)
	warnColour  = rgb(0xff5040)

	boltHead   = rgb(0xe8d070)
	boltShade  = rgb(0x9a8430)
	springCoil = rgb(0xc8ccd8)
	springDark = rgb(0x6a6e80)
	pieTin     = rgb(0xa8acb8)
	pieCrust   = rgb(0xe0b070)
	pieFill    = rgb(0xc04060)
	beltRubber = rgb(0x2a2a30)
	beltMark   = rgb(0x8a8a94)
	roller     = rgb(0x70747e)
	cable      = rgb(0x606478)
	lunchbox   = rgb(0xd03838)
	thermos    = rgb(0x38a060)
	toolbox    = rgb(0x3878d0)
	itemLight  = color.RGBA{255, 255, 255, 110}

	sidebarColour = rgb(0x12141f)
	sidebarEdge   = rgb(0x2e3248)
	hazardYellow  = rgb(0xf0c020)
	hazardBlack   = rgb(0x1a1a1a)
)

func rgb(v uint32) color.RGBA {
	return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
}

// mix is a blended a fraction t of the way to b.
func mix(a, b color.RGBA, t float32) color.RGBA {
	l := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*t) }
	return color.RGBA{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), 255}
}

// poly fills the polygon through the reference points xy (x0, y0, x1, ...).
func (f field) poly(c color.Color, xy ...float32) {
	var p vector.Path
	for i := 0; i+1 < len(xy); i += 2 {
		x, y := f.left+xy[i]*f.scale, f.top+xy[i+1]*f.scale
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	f.fill(&p, c)
}

func (f field) fill(p *vector.Path, c color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(f.dst, p, nil, op)
}

// ellipse fills an ellipse centred on (x, y).
func (f field) ellipse(x, y, rx, ry float32, c color.Color) {
	var p vector.Path
	const k = 0.5523 // cubic Bézier circle constant
	cx, cy := f.left+x*f.scale, f.top+y*f.scale
	ax, ay := rx*f.scale, ry*f.scale
	p.MoveTo(cx+ax, cy)
	p.CubicTo(cx+ax, cy+ay*k, cx+ax*k, cy+ay, cx, cy+ay)
	p.CubicTo(cx-ax*k, cy+ay, cx-ax, cy+ay*k, cx-ax, cy)
	p.CubicTo(cx-ax, cy-ay*k, cx-ax*k, cy-ay, cx, cy-ay)
	p.CubicTo(cx+ax*k, cy-ay, cx+ax, cy-ay*k, cx+ax, cy)
	p.Close()
	f.fill(&p, c)
}

// roundRect fills a rectangle with corners of radius r.
func (f field) roundRect(x, y, w, h, r float32, c color.Color) {
	r = min(r, w/2, h/2)
	f.rect(x+r, y, w-2*r, h, c)
	f.rect(x, y+r, w, h-2*r, c)
	f.circle(x+r, y+r, r, c)
	f.circle(x+w-r, y+r, r, c)
	f.circle(x+r, y+h-r, r, c)
	f.circle(x+w-r, y+h-r, r, c)
}

// backdrop caches the static sky and skyline for one screen size.
type backdrop struct {
	img  *ebiten.Image
	w, h int
}

var theBackdrop backdrop

// drawBackdrop draws the dusk sky, the distant city and a tower crane
// behind the field, cached between frames.
func drawBackdrop(f field) {
	size := f.dst.Bounds().Size()
	b := &theBackdrop
	if b.img == nil || b.w != size.X || b.h != size.Y {
		if b.img != nil {
			b.img.Deallocate()
		}
		b.img, b.w, b.h = ebiten.NewImage(size.X, size.Y), size.X, size.Y
		paintBackdrop(field{b.img, f.left, f.top, f.scale})
	}
	f.dst.DrawImage(b.img, nil)
}

func paintBackdrop(f field) {
	const bands = 40
	for i := range bands {
		t := float32(i) / (bands - 1)
		f.rect(0, 200*float32(i)/bands, 260, 200/bands+0.5, mix(skyTop, skyBottom, t))
	}
	// The far skyline: blocks of offices with a few lit windows.
	far := []float32{0, 150, 18, 132, 30, 140, 44, 118, 60, 138, 74, 126, 92, 144, 104, 120, 122, 136,
		140, 112, 156, 134, 170, 124, 186, 140, 200, 116, 216, 132, 232, 122, 248, 138, 260, 150}
	for i := 0; i+3 < len(far); i += 2 {
		f.rect(far[i], far[i+1], far[i+2]-far[i]+0.5, 200-far[i+1], skylineFar)
	}
	for i := 0; i+3 < len(far); i += 2 {
		for wy := far[i+1] + 6; wy < 196; wy += 9 {
			for wx := far[i] + 3; wx < far[i+2]-3; wx += 6 {
				if int(wx*7+wy*3)%5 == 0 {
					f.rect(wx, wy, 2, 3, mix(skylineFar, windowLight, 0.55))
				}
			}
		}
	}
	near := []float32{0, 176, 26, 168, 52, 180, 80, 164, 108, 178, 150, 170, 182, 182, 214, 166, 240, 176, 260, 176}
	for i := 0; i+3 < len(near); i += 2 {
		f.rect(near[i], near[i+1], near[i+2]-near[i]+0.5, 200-near[i+1], skylineNear)
	}
	// A tower crane on the skyline, its jib reaching over the site.
	f.rect(214, 30, 5, 150, craneColour)
	for y := float32(34); y < 176; y += 8 {
		f.line(214, y, 219, y+8, 0.6, skyBottom)
	}
	f.rect(150, 28, 100, 3, craneColour)
	f.rect(214, 20, 5, 10, craneColour)
	f.line(216, 20, 160, 28, 0.6, craneColour)
	f.line(216, 20, 246, 28, 0.6, craneColour)
	f.rect(238, 31, 10, 8, craneColour)
	f.line(176, 31, 176, 60, 0.4, craneColour)
}
