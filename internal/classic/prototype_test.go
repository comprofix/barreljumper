package classic

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/barreljumper/internal/input"
	"github.com/comprofix/barreljumper/internal/scene"
	"github.com/comprofix/barreljumper/internal/scene/scenetest"
)

func TestPrototypeControlsPauseRestartAndMenu(t *testing.T) {
	h := scenetest.New(t, 42)
	p := NewPrototype()
	h.Start(p)
	h.Ticks(2*scene.FadeTicks + 3)
	h.Run(scenetest.Hold(12, ebiten.KeyArrowRight))
	if p.movement.X <= 40 {
		t.Fatal("Right did not move the player")
	}
	h.Run(scenetest.Hold(8, ebiten.KeySpace))
	if !p.movement.jumping {
		t.Fatal("Space did not start a jump")
	}
	h.Press(ebiten.KeyP)
	scenetest.Current[*scene.Pause](h)
	x, y, counter := p.movement.X, p.movement.Y, p.movement.counter
	h.Run(scenetest.Hold(60, ebiten.KeyArrowRight))
	if p.movement.X != x || p.movement.Y != y || p.movement.counter != counter {
		t.Fatal("pause advanced the movement or rule clock")
	}
	h.Press(ebiten.KeyEscape)
	if scenetest.Current[*Prototype](h) != p {
		t.Fatal("resume replaced the prototype")
	}
	previous := p.movement
	h.Run(input.Press(ebiten.KeyControlLeft, ebiten.KeyR), input.Idle(4))
	if p.movement == previous || p.movement.X != 40 || p.movement.Y != 178 || p.movement.Dead {
		t.Fatal("Ctrl+R did not reset and settle the movement field")
	}
	h.Press(ebiten.KeyControlLeft, ebiten.KeyM)
	scenetest.Current[*scene.MainMenu](h)
}
