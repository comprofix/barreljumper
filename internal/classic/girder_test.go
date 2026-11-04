package classic

import "testing"

func TestBarrelsGirderRows(t *testing.T) {
	cells := barrelsGirders()
	if len(cells) != 73 {
		t.Fatalf("got %d cells, want 73", len(cells))
	}
	for _, check := range []struct {
		index int
		want  girder
	}{
		{0, girder{0, 20, 194}},
		{12, girder{240, 260, 194}},
		{24, girder{220, 240, 172}},
		{36, girder{240, 260, 129}},
		{48, girder{220, 240, 108}},
		{60, girder{240, 260, 65}},
		{68, girder{140, 160, 40}},
		{69, girder{160, 180, 41}},
		{72, girder{220, 240, 44}},
	} {
		if got := cells[check.index]; got != check.want {
			t.Errorf("cell %d = %+v, want %+v", check.index, got, check.want)
		}
	}
}

func TestGirderContactBoundaries(t *testing.T) {
	cell := girder{20, 40, 161}
	for _, check := range []struct {
		x, foot int
		want    bool
	}{
		{7, 161, false},
		{8, 161, true},
		{35, 163, true},
		{36, 163, false},
		{20, 160, false},
		{20, 164, false},
	} {
		if got := cell.supports(check.x, check.foot); got != check.want {
			t.Errorf("support at (%d, %d) = %v, want %v", check.x, check.foot, got, check.want)
		}
	}
}

func TestGirderLandingY(t *testing.T) {
	cell := girder{20, 40, 161}
	for _, tc := range []struct {
		x, y, offset, wantY int
		wantContact         bool
	}{
		{20, 145, 16, 144, true},
		{20, 147, 16, 144, true},
		{20, 149, 12, 148, true},
		{20, 148, 16, 148, false},
		{36, 145, 16, 145, false},
	} {
		y, contact := cell.landingY(tc.x, tc.y, tc.offset)
		if y != tc.wantY || contact != tc.wantContact {
			t.Errorf("landingY(%d, %d, %d) = (%d, %v), want (%d, %v)", tc.x, tc.y, tc.offset, y, contact, tc.wantY, tc.wantContact)
		}
	}
}
