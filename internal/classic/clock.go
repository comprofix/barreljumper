package classic

// autoUpdateEligible reports whether Classic AUTO runs a simulation update
// for the sampled PIT counter. Counter sampling and pacing belong to the caller.
func autoUpdateEligible(counter uint16) bool {
	return counter%12 != 0
}
