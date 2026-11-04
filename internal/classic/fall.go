package classic

// walkingFall tracks a fall after walking off a girder. Jump descent uses
// separate rules. Each unsupported player update advances one reference pixel.
type walkingFall struct {
	distance int
}

func (f *walkingFall) advance(y int) int {
	f.distance++
	return y + 1
}

// land applies the Classic fatal-fall threshold and clears the fall count.
func (f *walkingFall) land() (fatal bool) {
	fatal = f.distance > 16
	f.distance = 0
	return fatal
}

// landOn checks a candidate girder at the current position. A miss preserves
// the fall count so the next contact can still apply the fatal-fall rule.
func (f *walkingFall) landOn(g girder, x, y, footOffset int) (landingY int, contact, fatal bool) {
	landingY, contact = g.landingY(x, y, footOffset)
	if contact {
		fatal = f.land()
	}
	return
}
