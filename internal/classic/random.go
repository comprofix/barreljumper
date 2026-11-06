package classic

// random is the Classic gameplay random stream in DESIGN.md: a 32-bit
// linear congruential state whose draws are bits 16-30 of the new state.
// Stages call it in the reference decision order (DESIGN section 2, random
// parity); cosmetic choices do not draw from it.
type random struct {
	state uint32
}

func (r *random) next() int {
	r.state = r.state*22695477 + 1
	return int(r.state>>16) & 0x7fff
}

// bucket draws one of n equally likely values, 0 to n-1.
func (r *random) bucket(n int) int {
	return r.next() * n / 32768
}
