package classic

import (
	"testing"

	"github.com/comprofix/barreljumper/internal/scene/scenetest"
)

func twoPlayers(t *testing.T) *Play {
	h := scenetest.New(t, 1)
	h.Ctx.Players = 2
	return NewPlay(h.Ctx, skillByID(t, "arcade"))
}

// dieNow runs out the bonus and plays the death sequence through.
func dieNow(t *testing.T, p *Play) {
	t.Helper()
	p.stage.bonus, p.updates = 100, 119
	p.tick(Controls{}, true)
	if p.phase != dying {
		t.Fatalf("phase %d after the bonus ran out", p.phase)
	}
	for range deathFreezeTicks + deathSpinTicks + deathHoldTicks {
		p.tick(Controls{}, true)
	}
}

func clearNow(p *Play) {
	p.stage.player.Y = barrelsGoal.y - supportFoot
	p.tick(Controls{}, false)
	for range clearTicks {
		p.tick(Controls{}, false)
	}
}

func TestTwoPlayersTakeTurnsAfterDeaths(t *testing.T) {
	p := twoPlayers(t)
	if p.current != 0 || p.players[0].reserves != 2 || p.players[1].reserves != 3 {
		t.Fatalf("start: current %d, reserves %d and %d", p.current, p.players[0].reserves, p.players[1].reserves)
	}
	clearNow(p) // player 1 clears Barrels and keeps the turn
	if p.current != 0 || p.stage.family != rivetsFamily || p.players[0].score == 0 {
		t.Fatalf("after a clear: current %d, family %d", p.current, p.stage.family)
	}
	dieNow(t, p)
	if p.current != 1 || p.players[1].reserves != 2 || p.stage.family != barrelsFamily || p.score != 0 {
		t.Fatalf("after player 1 died: current %d, reserves %d, family %d, score %d",
			p.current, p.players[1].reserves, p.stage.family, p.score)
	}
	dieNow(t, p)
	if p.current != 0 || p.stage.family != rivetsFamily || p.reserves != 1 {
		t.Errorf("back to player 1: current %d, family %d, reserves %d", p.current, p.stage.family, p.reserves)
	}
}

func TestTwoPlayersOneRunsOut(t *testing.T) {
	p := twoPlayers(t)
	p.players[0].reserves = 0
	dieNow(t, p)
	if p.phase != playerOut {
		t.Fatalf("phase %d when player 1 ran out, want the Game Over Player 1 message", p.phase)
	}
	for range playerOutTicks {
		p.tick(Controls{}, true)
	}
	if p.phase != playing || p.current != 1 || p.reserves != 2 {
		t.Fatalf("after the message: phase %d, current %d, reserves %d", p.phase, p.current, p.reserves)
	}
	// Player 1 has nothing left, so the turn skips back to player 2.
	dieNow(t, p)
	if p.phase != playing || p.current != 1 || p.reserves != 1 {
		t.Errorf("player 2 died: phase %d, current %d, reserves %d", p.phase, p.current, p.reserves)
	}
	p.reserves = 0
	dieNow(t, p)
	if p.phase != gameOver {
		t.Errorf("phase %d with neither player left", p.phase)
	}
}

// TestExtraLifeOnTheDeathUpdate checks a life earned on the update of a
// death is counted before the reserves are (DESIGN section 4, M4 session
// and turns).
func TestExtraLifeOnTheDeathUpdate(t *testing.T) {
	p := NewPlay(nil, skillByID(t, "arcade"))
	p.reserves, p.score = 0, 6999
	p.stage.earned = 1
	dieNow(t, p)
	if p.phase != playing || p.reserves != 0 {
		t.Errorf("phase %d, reserves %d: the 7000 life should give one more turn", p.phase, p.reserves)
	}
}
