package classic

import "testing"

func TestWalkingFallLanding(t *testing.T) {
	var fall walkingFall
	y := 100
	for range 16 {
		y = fall.advance(y)
	}
	if y != 116 || fall.land() {
		t.Fatalf("16-update fall ended at Y=%d, want 116 and a safe landing", y)
	}

	for range 17 {
		y = fall.advance(y)
	}
	if y != 133 || !fall.land() {
		t.Fatalf("17-update fall ended at Y=%d, want 133 and a fatal landing", y)
	}
	if fall.land() {
		t.Fatal("landing retained the previous fall count")
	}
}
