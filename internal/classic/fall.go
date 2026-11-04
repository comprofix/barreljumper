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
