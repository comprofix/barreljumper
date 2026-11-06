package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/gfx"
	"github.com/comprofix/barreljumper/internal/ui"
)

// BlinkTicks is the blink period of "press any key" prompts.
const BlinkTicks = config.TPS

// Title is the title screen: the game's name and a blinking prompt, over
// the game playing itself if it has a demo (Context.NewDemo). Any key or
// button goes to the main menu.
type Title struct {
	shown int64 // ticks since the scene started, for the blink
	demo  Scene // nil until the first update, or if the game has none
}

// TitleDim is how far the demo is dimmed under the title, 0–1.
const TitleDim = 0.45

// Demo is the demo running behind the title, or nil.
func (t *Title) Demo() Scene { return t.demo }

// NewTitle returns the title scene.
func NewTitle() *Title { return &Title{} }

// AttractTicks is how long a title with its own art (Context.DrawTitle)
// shows before the demo plays: 20 seconds.
const AttractTicks = 20 * config.TPS

// DemoOver is a demo that can finish on its own, which ends the attract
// demo and returns to the title.
type DemoOver interface {
	Over() bool
}

// Update implements Scene.
func (t *Title) Update(ctx *Context) error {
	t.shown++
	if ctx.DrawTitle != nil {
		return t.attract(ctx)
	}
	playTitleTune(ctx)
	if t.demo == nil && ctx.NewDemo != nil {
		t.demo = ctx.makeGame(ctx.NewDemo)
	}
	if t.demo != nil {
		if err := updateScene(ctx, t.demo); err != nil {
			return err
		}
	}
	if ctx.Input.AnyJustPressed() {
		ctx.Audio.PlayEffect(MenuConfirm)
		ctx.Switch(NewMainMenu())
	}
	return nil
}

// attract runs the attract cycle of a title with its own art: the art
// with the title tune, which a key leaves for the main menu, then the
// demo, which a key or its end returns from.
func (t *Title) attract(ctx *Context) error {
	if t.demo == nil {
		playTitleTune(ctx)
		if ctx.Input.AnyJustPressed() {
			ctx.Audio.PlayEffect(MenuConfirm)
			ctx.Switch(NewMainMenu())
			return nil
		}
		if ctx.NewDemo != nil && t.shown >= AttractTicks {
			ctx.Audio.Stop()
			t.demo = ctx.makeGame(ctx.NewDemo)
		}
		return nil
	}
	if err := updateScene(ctx, t.demo); err != nil {
		return err
	}
	over, _ := t.demo.(DemoOver)
	if ctx.Input.AnyJustPressed() || over != nil && over.Over() {
		ctx.Audio.Stop()
		t.demo, t.shown = nil, 0
	}
	return nil
}

// Draw implements Scene.
func (t *Title) Draw(ctx *Context, dst *ebiten.Image) {
	if ctx.DrawTitle != nil {
		t.drawAttract(ctx, dst)
		return
	}
	sc, f := ctx.Scheme, ctx.Fonts
	w, h := ctx.Preset.ScreenWidth, ctx.Preset.ScreenHeight
	// Without a demo, the name a third of the way down and the prompt two
	// thirds; over a demo, both together on a panel, so they read over it.
	y, py := h/3-f.Large.LineHeight()/2, h*2/3
	if t.demo != nil {
		drawScene(ctx, t.demo, dst)
		Dim(dst, sc.Shadow, TitleDim)
		pad := 2 * f.Small.LineHeight()
		body := f.Large.LineHeight() + pad + f.Small.LineHeight()
		y = (h - body) / 2
		py = y + f.Large.LineHeight() + pad
		ui.Panel(dst, ui.PanelAround(w, f.Large.Width(config.Title)+2*pad, body, y, pad), sc.Panel, sc.Border)
	} else {
		dst.Fill(gfx.RGBA(sc.Backdrop))
	}

	ui.TextCentred(dst, f.Large, config.Title, y, sc.Title, sc.Shadow)
	if ui.Blink(t.shown, BlinkTicks) {
		ui.TextCentred(dst, f.Small, "PRESS ANY KEY", py, sc.Text, sc.Shadow)
	}
	ui.TextCentred(dst, f.Small, "F11 fullscreen   F10 quit", h-2*f.Small.LineHeight(), sc.Dim, sc.Shadow)
}

// drawAttract draws the game's title art with the prompt, or the demo.
func (t *Title) drawAttract(ctx *Context, dst *ebiten.Image) {
	sc, f := ctx.Scheme, ctx.Fonts
	h := ctx.Preset.ScreenHeight
	if t.demo != nil {
		drawScene(ctx, t.demo, dst)
		return
	}
	ctx.DrawTitle(ctx, dst)
	if ui.Blink(t.shown, BlinkTicks) {
		ui.TextCentred(dst, f.Small, "PRESS ANY KEY", h-4*f.Small.LineHeight(), sc.Text, sc.Shadow)
	}
	ui.TextCentred(dst, f.Small, "F11 fullscreen   F10 quit", h-2*f.Small.LineHeight(), sc.Dim, sc.Shadow)
}
