package classic

import "testing"

// TestDemoRun checks the attract demo's fixed run: from its seed the pilot
// clears the first Barrels screen, jumping barrels on the way, and the
// demo ends after the carry-off, well inside its time limit.
func TestDemoRun(t *testing.T) {
	d := newDemo(demoSeed)
	jumps := 0
	for !d.over {
		was := d.p.stage.player.jumping
		d.step()
		if d.p.stage.player.jumping && !was {
			jumps++
		}
	}
	if d.died || !d.ending {
		t.Fatalf("the demo did not clear the screen: died %v", d.died)
	}
	if jumps < 3 || d.ticks >= demoTicks {
		t.Errorf("%d jumps in %d ticks", jumps, d.ticks)
	}
}
