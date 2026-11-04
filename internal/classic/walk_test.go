package classic

import "testing"

func TestGroundX(t *testing.T) {
	for _, tc := range []struct {
		name         string
		x, direction int
		skip, hammer bool
		want         int
	}{
		{"right", 40, 1, false, false, 41},
		{"left", 40, -1, false, false, 39},
		{"idle", 40, 0, false, false, 40},
		{"skipped", 40, 1, true, false, 40},
		{"left edge", 0, -1, false, false, 0},
		{"right edge", 244, 1, false, false, 244},
		{"hammer left edge", 9, -1, false, true, 9},
		{"hammer right edge", 236, 1, false, true, 236},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := groundX(tc.x, tc.direction, tc.skip, tc.hammer); got != tc.want {
				t.Errorf("ground X = %d, want %d", got, tc.want)
			}
		})
	}
}
