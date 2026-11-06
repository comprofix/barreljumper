package classic

// girder describes a collision surface in Classic reference coordinates.
type girder struct {
	left, right, y int
}

// barrelsGirders returns the Barrels collision cells specified in DESIGN.md.
func barrelsGirders() []girder {
	rows := [...]struct {
		start, count, height, step int
	}{
		{0, 13, 194, 0},
		{0, 12, 161, 1},
		{20, 12, 140, -1},
		{0, 12, 97, 1},
		{20, 12, 76, -1},
		{0, 12, 40, 0},
	}
	surfaces := make([]girder, 0, 73)
	for _, row := range rows {
		for cell := 0; cell < row.count; cell++ {
			left := row.start + cell*20
			y := row.height + cell*row.step
			if row.height == 40 && left > 140 {
				y += (left - 140) / 20
			}
			surfaces = append(surfaces, girder{left, left + 20, y})
		}
	}
	return surfaces
}

// supports reports whether the active foot is within this girder's contact
// band. A zero-width support, such as a removed rivet, supports nothing.
func (g girder) supports(playerX, footY int) bool {
	return g.right > g.left && g.left < playerX+13 && playerX+5 <= g.right &&
		g.y <= footY && footY <= g.y+2
}

// landingY returns the snapped origin when the active foot contacts this girder.
func (g girder) landingY(playerX, playerY, footOffset int) (int, bool) {
	if !g.supports(playerX, playerY+footOffset) {
		return playerY, false
	}
	return g.y - footOffset - 1, true
}
