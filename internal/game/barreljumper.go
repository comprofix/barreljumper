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
}
