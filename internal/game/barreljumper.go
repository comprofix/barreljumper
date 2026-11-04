package game

import (
	"github.com/comprofix/barreljumper/internal/classic"
	"github.com/comprofix/barreljumper/internal/config"
	"github.com/comprofix/barreljumper/internal/scene"
)

func init() {
	content.NewPlay = func(_ *scene.Context, _ config.Skill) scene.Scene {
		return classic.NewPrototype()
	}
}
