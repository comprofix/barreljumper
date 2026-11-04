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

func TestWalkingFallGirderContact(t *testing.T) {
	cell := girder{20, 40, 161}
	var fall walkingFall
	y := 128
	for range 17 {
		y = fall.advance(y)
	}
	if got, contact, fatal := fall.landOn(cell, 36, y, 16); got != 145 || contact || fatal {
		t.Fatalf("miss = (%d, %v, %v), want (145, false, false)", got, contact, fatal)
	}
	if got, contact, fatal := fall.landOn(cell, 20, y, 16); got != 144 || !contact || !fatal {
		t.Fatalf("landing = (%d, %v, %v), want (144, true, true)", got, contact, fatal)
	}
}
