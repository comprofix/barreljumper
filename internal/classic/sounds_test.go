package classic

import (
	"testing"

	"github.com/comprofix/barreljumper/internal/audio"
	"github.com/comprofix/barreljumper/internal/scene"
)

func TestSoundTables(t *testing.T) {
	if err := audio.Join(scene.Sounds, Sounds).Check(); err != nil {
		t.Fatal(err)
	}
	for _, e := range effects {
		for _, fx := range e.effects {
			found := false
			for _, b := range Sounds.Effects {
				found = found || b == fx
			}
			if !found {
				t.Errorf("%s is not in the sound test bank", fx.Name)
			}
		}
	}
}

// TestFootfallsFollowTheStride checks a step sounds each time the drawn
// stride changes, alternating feet.
func TestFootfallsFollowTheStride(t *testing.T) {
	p := toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	p.jingle = nil
	m := p.stage.player
	var steps []sound
	lastX := m.X
	for range 400 {
		p.tick(Controls{Horizontal: 1}, true)
		if got := p.sounds & (soundStep | soundStepB); got != 0 {
			steps = append(steps, got)
			if footfall(m.X) == footfall(lastX) {
				t.Fatalf("a step sounded at X %d without a new stride", m.X)
			}
		}
		lastX = m.X
		p.sounds = 0
	}
	if len(steps) < 4 {
		t.Fatalf("%d steps while walking", len(steps))
	}
	for i := 1; i < len(steps); i++ {
		if steps[i] == steps[i-1] {
			t.Errorf("steps %d and %d used the same foot", i-1, i)
		}
	}
}

func TestGameplaySounds(t *testing.T) {
	p := toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	if p.jingle != startJingle {
		t.Error("no start jingle on entering a screen")
	}
	p.jingle = nil
	m := p.stage.player
	for i := 0; i < 40 && !m.jumping; i++ {
		p.tick(Controls{Jump: i >= 20}, true) // settle and rearm, then jump
	}
	if p.sounds&soundJump == 0 || !m.jumping {
		t.Errorf("sounds %b after a jump", p.sounds)
	}
	p.sounds = 0
	for i := 0; i < 200 && m.jumping; i++ {
		p.tick(Controls{}, true)
	}
	if p.sounds&soundLand == 0 {
		t.Errorf("sounds %b after landing", p.sounds)
	}
	p.sounds = 0
	p.stage.award(100, 0, 0)
	p.stage.startSmash(0, 0, smashedBarrel)
	p.tick(Controls{}, false)
	if p.sounds&(soundPoints|soundSmash) != soundPoints|soundSmash {
		t.Errorf("sounds %b after an award and a smash", p.sounds)
	}
	p = toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	p.stage.player.Y = barrelsGoal.y - supportFoot
	p.tick(Controls{}, false)
	if p.jingle != clearJingle {
		t.Error("no clear jingle")
	}
	p = toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	p.stage.bonus, p.updates = 100, 119
	p.tick(Controls{}, true)
	if p.sounds&soundHit == 0 {
		t.Error("no hit sound at the death")
	}
	p.sounds = 0
	for range deathFreezeTicks + deathSpinTicks {
		p.tick(Controls{}, true)
	}
	if p.sounds&(soundSpin|soundThud) != soundSpin|soundThud {
		t.Errorf("sounds %b through the death sequence", p.sounds)
	}
}

// TestLowBonusWarning checks the warning sounds every 34 eligible updates
// once the bonus is below 1000, and not above it.
func TestLowBonusWarning(t *testing.T) {
	p := toPlay(t, NewPlay(nil, skillByID(t, "arcade")))
	count := func(bonus int) int {
		p.stage.bonus, p.updates, p.stage.sounds = bonus, 1, 0
		n := 0
		for range 68 {
			p.stage.update(Controls{}, true)
			if p.stage.sounds&soundWarn != 0 {
				n++
			}
			p.stage.sounds = 0
		}
		return n
	}
	if n := count(900); n != 2 {
		t.Errorf("%d warnings in 68 updates at bonus 900, want 2", n)
	}
	if n := count(1000); n != 0 {
		t.Errorf("%d warnings at bonus 1000", n)
	}
}
