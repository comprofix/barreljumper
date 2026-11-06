package classic

type ladderKind uint8

const (
	fullLadder ladderKind = iota
	lowerSegment
	upperSegment
)

type ladder struct {
	x, top, height int
	kind           ladderKind
}

// barrelsLadders lists the 17 attachable records in the documented Barrels
// layout, in the reference record order. Player, barrel and fireball scans
// take the first match, so the order is part of the specification.
var barrelsLadders = [...]ladder{
	{220, 178, 16, fullLadder}, {30, 146, 16, fullLadder},
	{220, 114, 16, fullLadder}, {30, 82, 16, fullLadder},
	{130, 141, 26, fullLadder}, {90, 79, 22, fullLadder},
	{150, 110, 24, fullLadder}, {220, 50, 16, fullLadder},
	{190, 96, 10, lowerSegment}, {190, 74, 5, upperSegment},
	{110, 62, 10, lowerSegment}, {110, 46, 5, upperSegment},
	{80, 123, 14, lowerSegment}, {80, 107, 5, upperSegment},
	{100, 184, 10, lowerSegment}, {100, 172, 5, upperSegment},
	{150, 30, 10, fullLadder},
}

// barrelsReservedLadders are the four upper records beside the thrower. No
// player, barrel or fireball scan uses them; they are drawn only.
var barrelsReservedLadders = [...]ladder{
	{80, 0, 20, fullLadder}, {100, 0, 20, fullLadder},
	{80, 20, 20, fullLadder}, {100, 20, 20, fullLadder},
}

// attach returns the snapped player X when the player can enter the ladder.
func (l ladder) attach(playerX, footY int, up bool) (int, bool) {
	center := playerX + 8
	bottom := l.top + l.height
	if center <= l.x-1 || center >= l.x+8 || footY < l.top-6 || footY > bottom {
		return playerX, false
	}
	if (up && footY == l.top-6) || (!up && footY == bottom) {
		return playerX, false
	}
	return l.x - 4, true
}

// climbLimit applies the endpoint rules to the resulting foot coordinate.
func (l ladder) climbLimit(footY int) (int, bool) {
	bottom := l.top + l.height
	switch l.kind {
	case lowerSegment:
		return max(footY, l.top+2), footY < bottom
	case upperSegment:
		return min(footY, bottom+5), footY > l.top-6
	default:
		return footY, footY > l.top-6 && footY < bottom
	}
}
