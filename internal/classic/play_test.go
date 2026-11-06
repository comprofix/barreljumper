package classic

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/barreljumper/internal/scene"
	"github.com/comprofix/barreljumper/internal/scene/scenetest"
)

// TestPlayControls drives the scene: walking moves the player, P pauses,
// Ctrl+R starts a new game and Ctrl+M returns to the menu. Enter skips
// the opening.
func TestPlayControls(t *testing.T) {
	h := scenetest.New(t, 1)
	sk := skillByID(t, "arcade")
	h.Start(NewPlay(h.Ctx, sk))
	h.Ticks(2 * scene.FadeTicks)
	p := scenetest.Current[*Play](h)
	if p.phase != opening {
		t.Fatalf("phase %d at the start, want the opening", p.phase)
	}
	h.Press(ebiten.KeyEnter) // skips the opening to the floor card
	h.Ticks(cardTicks)
	if p.phase != playing {
		t.Fatalf("phase %d after the floor card", p.phase)
	}
	h.Run(scenetest.Hold(60, ebiten.KeyArrowRight))
	if p.stage.player.X <= 40 {
		t.Errorf("holding Right left the player at X %d", p.stage.player.X)
	}
	h.Press(ebiten.KeyP)
	scenetest.Current[*scene.Pause](h)
	h.Press(ebiten.KeyP)
	if scenetest.Current[*Play](h) != p {
		t.Fatal("unpausing did not return to the same game")
	}
	h.Press(ebiten.KeyControlLeft, ebiten.KeyR)
	if q := scenetest.Current[*Play](h); q == p || q.stage.player.X != 40 {
		t.Error("Ctrl+R did not start a new game")
	}
	h.Press(ebiten.KeyControlLeft, ebiten.KeyM)
	scenetest.Current[*scene.MainMenu](h)
}
