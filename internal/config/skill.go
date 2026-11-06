package config

// Skill is a difficulty level the player picks before a game.
// Games replace these with their own skills and tuning.
type Skill struct {
	ID   string // stable key, used in saved files
	Name string // shown in menus
	Desc string // one line shown under the menu
}

// Skills are the Classic reference skills, easiest first. Arcade is the
// verified comparison baseline.
var Skills = []Skill{
	{ID: "easy", Name: "Easy", Desc: "Five lives; a held hammer guards you."},
	{ID: "arcade", Name: "Arcade", Desc: "Three lives, as the arcade plays."},
	{ID: "hard", Name: "Hard", Desc: "Three lives and faster throws."},
}

// SkillNamed returns the skill with ID id, or nil.
func SkillNamed(id string) *Skill {
	for i := range Skills {
		if Skills[i].ID == id {
			return &Skills[i]
		}
	}
	return nil
}

// DefaultSkill is the skill a new game setup starts with: Arcade.
func DefaultSkill() Skill { return *SkillNamed("arcade") }
