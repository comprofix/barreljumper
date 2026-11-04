package classic

import "testing"

func TestAutoUpdateEligibility(t *testing.T) {
	for _, tc := range []struct {
		counter uint16
		want    bool
	}{
		{0, false},
		{1, true},
		{11, true},
		{12, false},
		{13, true},
		{24, false},
		{65535, true},
	} {
		if got := autoUpdateEligible(tc.counter); got != tc.want {
			t.Errorf("counter %d: eligible = %v, want %v", tc.counter, got, tc.want)
		}
	}
}
