package classic

import "testing"

func TestSidebarRows(t *testing.T) {
	p := NewPlay(nil, skillByID(t, "arcade"))
	p.score = 4200
	rows := p.sidebarRows(12300)
	want := []string{"HI SCORE 12300", "SCORE 4200", "BONUS 5000", "LEVEL 1 "}
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(rows), len(want))
	}
	for i, r := range rows {
		if got := r.label + " " + r.value; got != want[i] {
			t.Errorf("row %d is %q, want %q", i, got, want[i])
		}
	}
	if !rows[3].hats || rows[2].valueColour != "Amber" {
		t.Errorf("lives row hats %v, bonus colour %q", rows[3].hats, rows[2].valueColour)
	}

	// A score past the table's best is the high score; a low bonus is red.
	p.score, p.stage.bonus = 15000, lowBonus-100
	rows = p.sidebarRows(12300)
	if rows[0].value != "15000" || rows[2].valueColour != "Red" {
		t.Errorf("hi score %s, bonus colour %q", rows[0].value, rows[2].valueColour)
	}

	// With two players, each has a plate and the current one is lit.
	p.players = append(p.players, p.players[0])
	p.player = &p.players[0]
	p.players[1].score = 700
	rows = p.sidebarRows(0)
	if len(rows) != 5 || rows[1].label != "PLAYER 1" || rows[2].label != "PLAYER 2" || rows[2].value != "700" {
		t.Fatalf("two-player rows %+v", rows)
	}
	if rows[1].labelColour != "Gold" || rows[2].labelColour != "" {
		t.Errorf("label colours %q, %q", rows[1].labelColour, rows[2].labelColour)
	}
}
