package classic

// Stage families, numbered as the reference stage IDs (DESIGN section 4,
// Screen progression).
const (
	barrelsFamily = iota
	conveyorFamily
	elevatorsFamily
	rivetsFamily
)

// classicSchedule is the Classic screen order for each level modulo five.
var classicSchedule = [5][]int{
	{barrelsFamily, rivetsFamily},
	{barrelsFamily, elevatorsFamily, rivetsFamily},
	{barrelsFamily, conveyorFamily, elevatorsFamily, rivetsFamily},
	{barrelsFamily, conveyorFamily, barrelsFamily, elevatorsFamily, rivetsFamily},
	{barrelsFamily, conveyorFamily, barrelsFamily, elevatorsFamily, barrelsFamily, rivetsFamily},
}

// point is a reference origin.
type point struct{ x, y int }

// layout is one stage's setup: its live support and ladder records, in
// reference record order, and where things start (DESIGN section 4, Stage
// layout anchors and the M3 setup notes). Each call returns fresh slices,
// since stages change some records as they play.
type layout struct {
	family   int
	girders  []girder
	ladders  []ladder // the records scanned by players and fireballs
	reserved []ladder // the final four records, drawn only
	player   point
	thrower  point
	oil      oilDrum
	hammers  []point
	items    []point
}

func newLayout(family int) layout {
	switch family {
	case rivetsFamily:
		return rivetsLayout()
	case elevatorsFamily:
		return elevatorsLayout()
	}
	return barrelsLayout()
}

func barrelsLayout() layout {
	return layout{
		family:   barrelsFamily,
		girders:  barrelsGirders(),
		ladders:  append([]ladder(nil), barrelsLadders[:]...),
		reserved: append([]ladder(nil), barrelsReservedLadders[:]...),
		player:   point{40, 176},
		thrower:  point{30, 6},
		oil:      oilDrum{x: 18, y: 164, present: true},
		hammers:  []point{{230, 152}, {18, 47}},
	}
}

// rivetsLayout follows the Rivets setup. The last eight girders are the
// rivet supports, in rivet order; its four reserved ladder records are empty.
func rivetsLayout() layout {
	var g []girder
	for x := 0; x <= 240; x += 20 {
		g = append(g, girder{x, x + 20, 194})
	}
	rows := []struct {
		y     int
		cells []int // starts of the fixed cells, 20 wide unless listed in short
		short map[int]bool
	}{
		{162, []int{10, 20, 40, 70, 90, 110, 130, 150, 170, 200, 210, 230}, map[int]bool{10: true, 200: true}},
		{130, []int{20, 40, 70, 90, 110, 130, 150, 170, 200, 220}, nil},
		{98, []int{30, 40, 70, 90, 110, 130, 150, 170, 200, 220}, map[int]bool{30: true, 220: true}},
		{66, []int{40, 70, 90, 110, 130, 150, 170, 200}, nil},
		{24, []int{60, 80, 100, 120, 140, 160, 180}, nil},
	}
	for _, r := range rows {
		for _, x := range r.cells {
			w := 20
			if r.short[x] {
				w = 10
			}
			g = append(g, girder{x, x + w, r.y})
		}
	}
	for _, x := range []int{60, 190} {
		for _, y := range rivetRows {
			g = append(g, girder{x, x + 10, y})
		}
	}
	return layout{
		family:  rivetsFamily,
		girders: g,
		ladders: []ladder{
			{10, 168, 26, fullLadder}, {125, 168, 26, fullLadder}, {240, 168, 26, fullLadder},
			{20, 136, 26, fullLadder}, {80, 136, 26, fullLadder}, {170, 136, 26, fullLadder}, {230, 136, 26, fullLadder},
			{30, 104, 26, fullLadder}, {125, 104, 26, fullLadder}, {220, 104, 26, fullLadder},
			{40, 72, 26, fullLadder}, {80, 72, 26, fullLadder}, {170, 72, 26, fullLadder}, {210, 72, 26, fullLadder},
		},
		player:  point{30, 176},
		thrower: point{110, 32},
		hammers: []point{{10, 102}, {124, 74}},
		items:   []point{{144, 184}, {239, 152}, {39, 56}},
	}
}

// rivetRows are the heights of the four removable rows, bottom first.
var rivetRows = [4]int{162, 130, 98, 66}
