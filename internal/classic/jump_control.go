package classic

// jumpControl tracks release-to-rearm for Classic takeoff. A blocked press
// leaves the control armed until the player can take off.
type jumpControl struct {
	armed bool
}

func (j *jumpControl) takeoff(held, eligible bool) bool {
	if !held {
		j.armed = true
		return false
	}
	if !j.armed || !eligible {
		return false
	}
	j.armed = false
	return true
}
