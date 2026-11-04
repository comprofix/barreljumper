package classic

import "testing"

func TestLadderAttachment(t *testing.T) {
	l := barrelsLadders[0]
	for _, tc := range []struct {
		x, foot  int
		up, want bool
	}{
		{211, 180, true, false},
		{212, 180, true, true},
		{219, 180, true, true},
		{220, 180, true, false},
		{216, 172, true, false},
		{216, 172, false, true},
		{216, 194, true, true},
		{216, 194, false, false},
	} {
		x, ok := l.attach(tc.x, tc.foot, tc.up)
		if ok != tc.want || (ok && x != 216) {
			t.Errorf("attach(%d, %d, %v) = (%d, %v), want eligibility %v and snap to 216 on success", tc.x, tc.foot, tc.up, x, ok, tc.want)
		}
	}
}

func TestLadderEndpoints(t *testing.T) {
	for _, tc := range []struct {
		index, foot, wantFoot int
		wantAttached          bool
	}{
		{0, 172, 172, false},
		{0, 193, 193, true},
		{0, 194, 194, false},
		{9, 68, 68, false},
		{9, 85, 84, true},
		{10, 97, 98, true},
		{10, 106, 106, false},
	} {
		foot, attached := barrelsLadders[tc.index].climbLimit(tc.foot)
		if foot != tc.wantFoot || attached != tc.wantAttached {
			t.Errorf("ladder %d at foot %d: got (%d, %v), want (%d, %v)", tc.index, tc.foot, foot, attached, tc.wantFoot, tc.wantAttached)
		}
	}
}
