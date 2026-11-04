package classic

const maxPlayerX = 244

// groundX advances an eligible walking update using the Classic movement rule
// in DESIGN.md. The caller supplies the current slot of the six-update cycle
// as a skip flag and the horizontal input as -1, 0 or 1.
func groundX(x, direction int, skipStep, hammer bool) int {
	left, right := 0, maxPlayerX
	if hammer {
		left, right = 9, 236
	}
	if skipStep {
		direction = 0
	}
	return min(max(x+direction, left), right)
}
