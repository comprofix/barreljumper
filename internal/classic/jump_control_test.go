package classic

import "testing"

func TestJumpControlRearmsOnRelease(t *testing.T) {
	jump := jumpControl{armed: true}
	for i, tc := range []struct {
		held, eligible, want bool
	}{
		{true, true, true},
		{true, false, false},
		{true, true, false},
		{false, true, false},
		{true, true, true},
	} {
		if got := jump.takeoff(tc.held, tc.eligible); got != tc.want {
			t.Fatalf("input %d: takeoff = %v, want %v", i, got, tc.want)
		}
	}
}
