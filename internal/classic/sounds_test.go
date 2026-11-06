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
		found := false
		for _, b := range Sounds.Effects {
			found = found || b == e.effect
		}
		if !found {
			t.Errorf("%s is not in the sound test bank", e.effect.Name)
		}
	}
}

func TestGameplaySounds(t *testing.T) {
	p := NewPlay(nil, skillByID(t, "arcade"))
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
	p.stage.award(100, 0, 0)
	p.stage.startSmash(0, 0)
	p.tick(Controls{}, false)
	if p.sounds&(soundPoints|soundSmash) != soundPoints|soundSmash {
		t.Errorf("sounds %b after an award and a smash", p.sounds)
	}
	p = NewPlay(nil, skillByID(t, "arcade"))
	p.stage.player.Y = barrelsGoal.y - supportFoot
	p.tick(Controls{}, false)
	if p.jingle != clearJingle {
		t.Error("no clear jingle")
	}
	p = NewPlay(nil, skillByID(t, "arcade"))
	p.stage.bonus, p.updates = 100, 119
	p.tick(Controls{}, true)
	if p.sounds&soundDeath == 0 {
		t.Error("no death sound")
	}
}
