package game

import (
	"github.com/comprofix/barreljumper/internal/classic"
	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/scene"
)

func init() {
	content.NewPlay = func(ctx *scene.Context, skill config.Skill) scene.Scene {
		return classic.NewPlay(ctx, skill)
	}
	content.MaxPlayers = 2
	content.Sounds = classic.Sounds
	content.TitleTune = classic.TitleTune
	content.DrawTitle = classic.DrawTitle
	// The reference's STARTING LEVEL setting offers levels 1 to 8.
	content.StartLevels = 8
}
