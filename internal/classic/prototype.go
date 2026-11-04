package classic

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/barreljumper/internal/input"
	"github.com/comprofix/barreljumper/internal/scene"
)

// Prototype exposes the movement test field through the skeleton's Play menu.
type Prototype struct {
	movement *Movement
}

func NewPrototype() *Prototype { return &Prototype{movement: NewMovement()} }

func (p *Prototype) Update(ctx *scene.Context) error {
	if ctx.Input.JustPressed(input.Pause) || ctx.Input.JustPressed(input.Back) {
		ctx.Cut(scene.NewPause(p))
		return nil
	}
	if ctx.Input.JustPressed(input.Menu) {
		ctx.Switch(scene.NewMainMenu())
		return nil
	}
	if ctx.Input.JustPressed(input.Restart) {
		p.movement = NewMovement()
		return nil
	}
	axis := func(negative, positive input.Action) int {
		if ctx.Input.Held(negative) == ctx.Input.Held(positive) {
			return 0
		}
		if ctx.Input.Held(negative) {
			return -1
		}
		return 1
	}
	p.movement.Advance(Controls{
		Horizontal: axis(input.Left, input.Right),
		Vertical:   axis(input.Up, input.Down),
		Jump:       ctx.Input.Held(input.Confirm),
	}, true)
	return nil
}

func (p *Prototype) Draw(_ *scene.Context, dst *ebiten.Image) {
	dst.Fill(color.RGBA{18, 22, 32, 255})
	size := dst.Bounds().Size()
	scale := min(float32(size.X)/260, float32(size.Y)/200)
	left := (float32(size.X) - 260*scale) / 2
	top := (float32(size.Y) - 200*scale) / 2
	rect := func(r image.Rectangle, c color.Color) {
		vector.FillRect(dst, left+float32(r.Min.X)*scale, top+float32(r.Min.Y)*scale,
			float32(r.Dx())*scale, float32(r.Dy())*scale, c, false)
	}
	for _, g := range p.movement.surfaces {
		rect(image.Rect(g.left, g.y, g.right, g.y+2), color.RGBA{224, 109, 91, 255})
	}
	for _, l := range barrelsLadders {
		rect(image.Rect(l.x, l.top, l.x+7, l.top+l.height), color.RGBA{85, 166, 185, 255})
	}
	x, y := p.movement.X, p.movement.Y
	player := color.RGBA{251, 217, 124, 255}
	if p.movement.Dead {
		player = color.RGBA{179, 62, 62, 255}
	}
	rect(image.Rect(x+2, y, x+14, y+16), player)
}
