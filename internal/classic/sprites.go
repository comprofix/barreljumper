package classic

// The construction-site sprites. Each function notes the collision
// rectangle it keeps to.

// drawGirder draws a painted steel I-beam on the support's 7-pixel band,
// with a soft shadow on whatever is below it.
func drawGirder(f field, g girder) {
	if g.right <= g.left {
		return
	}
	x, y, w := float32(g.left), float32(g.y), float32(g.right-g.left)
	f.rect(x+1, y+7, w, 2.5, shadowColour)
	f.rect(x, y, w, 7, steelDark)
	f.rect(x, y, w, 2, steel)
	f.rect(x, y, w, 0.8, steelLight)
	f.rect(x, y+5, w, 2, steel)
	f.rect(x, y+6.4, w, 0.6, steelEdge)
	// The web's cross-bracing and the bolts at each joint.
	for cx := x; cx < x+w-0.5; cx += 10 {
		f.line(cx+1, y+2.2, cx+5, y+4.8, 0.7, steelEdge)
		f.line(cx+5, y+4.8, cx+9, y+2.2, 0.7, steelEdge)
		f.circle(cx+1, y+1, 0.45, boltColour)
	}
}

// drawLadder draws a steel ladder over its record's span.
func drawLadder(f field, l ladder, dim bool) {
	if l.height == 0 {
		return
	}
	rail, shade := ladderRail, ladderShade
	if dim {
		rail, shade = ladderDim, mix(ladderDim, skyTop, 0.4)
	}
	x, top, h := float32(l.x), float32(l.top), float32(l.height)
	for y := top + 2; y < top+h; y += 4 {
		f.rect(x+1, y, 6, 1, shade)
		f.rect(x+1, y, 6, 0.5, rail)
	}
	f.rect(x, top, 1.2, h, rail)
	f.rect(x+6.8, top, 1.2, h, rail)
	f.rect(x+0.8, top, 0.4, h, shade)
	f.rect(x+7.6, top, 0.4, h, shade)
}

// drawCables draws the Elevators shafts' hoist cables, behind the cells.
func drawCables(f field) {
	for _, x := range []float32{40, 120} {
		f.rect(x+3, 22, 0.8, 178, cable)
		f.rect(x+16, 22, 0.8, 178, cable)
		f.rect(x, 20, 20, 3, steelDark)
		f.circle(x+10, 21.5, 2, roller)
	}
}

// drawElevatorCell dresses an elevator support as a lift cage floor.
func drawElevatorCell(f field, g girder) {
	x, y := float32(g.left), float32(g.y)
	f.rect(x+3, y-1, 0.8, 1, cable)
	f.rect(x+16, y-1, 0.8, 1, cable)
	f.rect(x, y+2, 20, 1, hazardYellow)
	for i := float32(0); i < 20; i += 4 {
		f.rect(x+i, y+2, 2, 1, hazardBlack)
	}
}

// drawBelt draws a conveyor's rubber tread on its girder, with chevrons
// that run in the belt's direction and rollers at its ends.
func drawBelt(f field, b belt, tick int64) {
	x, y, w := float32(b.x), float32(b.y), float32(b.width)
	f.rect(x, y-0.5, w, 2.2, beltRubber)
	offset := float32((tick/6)%8) * float32(b.dir)
	for cx := x + 2 + offset; cx < x+w-2; cx += 8 {
		if cx < x+1 {
			continue
		}
		tip := cx + 1.5*float32(b.dir)
		f.line(cx, y-0.2, tip, y+0.6, 0.5, beltMark)
		f.line(tip, y+0.6, cx, y+1.4, 0.5, beltMark)
	}
	f.circle(x+2, y+3.5, 2.2, roller)
	f.circle(x+w-2, y+3.5, 2.2, roller)
	f.circle(x+2, y+3.5, 0.8, beltRubber)
	f.circle(x+w-2, y+3.5, 0.8, beltRubber)
}

// drawThrower draws the brute in the boss contact box X+2..X+37,
// Y..Y+31: a hulking demolition hand in a red hard hat.
func drawThrower(f field, t thrower) {
	x, y := float32(t.x), float32(t.y)
	f.ellipse(x+20, y+31, 16, 1.5, shadowColour)
	// Legs and boots.
	f.rect(x+9, y+24, 8, 7, bruteBib)
	f.rect(x+23, y+24, 8, 7, bruteBib)
	f.roundRect(x+7, y+29, 11, 3, 1, boot)
	f.roundRect(x+22, y+29, 11, 3, 1, boot)
	// The barrel chest in a grey work shirt, and bib overalls over it.
	f.roundRect(x+5, y+8, 30, 18, 6, bruteShirt)
	f.roundRect(x+10, y+15, 20, 11, 2, bruteBib)
	f.rect(x+12, y+9, 2.5, 7, bruteBib)
	f.rect(x+25.5, y+9, 2.5, 7, bruteBib)
	f.circle(x+13.2, y+16, 1, hammerSteel)
	f.circle(x+26.8, y+16, 1, hammerSteel)
	f.rect(x+5, y+8, 30, 2, mix(bruteShirt, rgb(0xffffff), 0.15))
	// The head, a little small for the shoulders.
	f.circle(x+20, y+6, 5.5, bruteSkin)
	f.rect(x+16, y+7, 8, 3.5, mix(bruteSkin, eyeColour, 0.45)) // stubble
	f.rect(x+16.5, y+4.5, 2, 1.4, eyeColour)
	f.rect(x+21.5, y+4.5, 2, 1.4, eyeColour)
	f.rect(x+15.5, y+3.4, 3.5, 0.8, eyeColour) // a scowl
	f.rect(x+21, y+3.4, 3.5, 0.8, eyeColour)
	f.ellipse(x+20, y+1.5, 6.5, 3, bruteHat)
	f.rect(x+12.5, y+1.8, 15, 1.4, mix(bruteHat, eyeColour, 0.3))
	arm := func(ax, ay, w, h float32) {
		f.roundRect(ax, ay, w, h, 2.5, bruteShirt)
		f.circle(ax+w/2, ay+h, 3, bruteSkin)
	}
	switch t.pose {
	case throwerPickUp:
		arm(x+1, y+10, 6, 14)
		arm(x+33, y+10, 6, 9)
	case throwerThrow:
		arm(x+1, y+10, 6, 9)
		f.roundRect(x+32, y+9, 9, 6, 2.5, bruteShirt)
		f.circle(x+40, y+12, 3, bruteSkin)
	case throwerTauntA:
		arm(x+1, y+1, 6, 9)
		arm(x+33, y+10, 6, 12)
	case throwerTauntB:
		arm(x+1, y+10, 6, 12)
		arm(x+33, y+1, 6, 9)
	default:
		arm(x+1, y+10, 6, 11)
		arm(x+33, y+10, 6, 11)
	}
	if t.holding {
		drawSpecialBarrel(f, x+11, y+16)
	}
}

// drawCaptive draws the site architect waiting to be rescued, in the
// 10 x 17 box at (x, y), waving a rolled blueprint.
func drawCaptive(f field, x, y float32, tick int64) {
	f.ellipse(x+5, y+17, 5, 1, shadowColour)
	f.rect(x+2, y+11, 2.5, 6, denimDark)
	f.rect(x+5.5, y+11, 2.5, 6, denimDark)
	f.roundRect(x+1, y+5, 8, 7, 2, captiveShirt)
	f.circle(x+5, y+3, 2.8, skin)
	f.ellipse(x+5, y+1.4, 3, 1.8, captiveHair)
	f.rect(x+2, y+1.4, 1, 3, captiveHair)
	wave := float32((tick/20)%2) * 2
	f.line(x+8.5, y+6, x+10, y+1+wave, 1.2, captiveShirt)
	f.roundRect(x+9, y-2+wave, 2.2, 6, 1, blueprint)
	f.circle(x+5.8, y+3, 0.4, eyeColour)
}

// drawDrum draws the oil drum: contact X+2..X+15 at its lip, Y+8..Y+11.
func drawDrum(f field, o oilDrum, tick int64) {
	x, y := float32(o.x), float32(o.y)
	f.ellipse(x+9, y+30, 9, 1.5, shadowColour)
	f.rect(x+2, y+9, 14, 21, blueDrum)
	f.rect(x+2, y+9, 3, 21, blueLight)
	f.rect(x+13, y+9, 3, 21, blueDark)
	f.rect(x+2, y+15, 14, 5, hazardYellow)
	for i := float32(0); i < 14; i += 4 {
		f.poly(hazardBlack, x+2+i, y+20, x+4+i, y+20, x+6+i, y+15, x+4+i, y+15)
	}
	f.rect(x+1.5, y+8, 15, 1.5, hoopLight)
	f.rect(x+1.5, y+28.5, 15, 1.5, hoop)
	if o.burning {
		flick := float32(tick/4%3) - 1
		f.poly(flameOuter, x+3, y+8, x+5+flick, y+1, x+8, y+4, x+10, y-2-flick, x+12, y+3, x+14-flick, y+0, x+15, y+8)
		f.poly(flameMid, x+5, y+8, x+7, y+3+flick, x+9, y+5, x+11, y+1+flick, x+13, y+8)
		f.poly(flameCore, x+7, y+8, x+9, y+4-flick, x+11, y+8)
	}
}

// drawBarrel draws a rolling barrel: X+2..X+12, Y+1..Y+10, or the blue
// special barrel X+2..X+16.
func drawBarrel(f field, b barrel) {
	x, y := float32(b.x), float32(b.y-b.bounce)
	if b.kind != regularBarrel {
		drawSpecialBarrel(f, x, y)
		return
	}
	f.circle(x+7, y+6, 5, woodDark)
	f.circle(x+7, y+6, 4.3, wood)
	f.circle(x+6.2, y+5, 2.5, woodLight)
	f.circle(x+7, y+6, 2.2, wood)
	// Two hoops seen edge-on, turning as it rolls.
	a := float32(b.frame%4) * 0.785
	for _, k := range []float32{0, 3.1416} {
		dx, dy := 4.3*cos(a+k), 4.3*sin(a+k)
		f.line(x+7-dy*0.35+dx*0.2, y+6+dx*0.35+dy*0.2, x+7+dx, y+6+dy, 1, hoop)
	}
	f.circle(x+7, y+6, 1, woodDark)
}

// drawSpecialBarrel draws the blue steel drum lying on its side.
func drawSpecialBarrel(f field, x, y float32) {
	f.roundRect(x+2, y+1, 14, 10, 2.5, blueDrum)
	f.rect(x+3, y+2, 12, 2.5, blueLight)
	f.rect(x+3, y+8, 12, 2, blueDark)
	f.rect(x+5, y+1, 1.2, 10, hoop)
	f.rect(x+12, y+1, 1.2, 10, hoop)
}

// drawFireball draws a living flame in X+3..X+12, Y+2..Y+14; it turns
// blue and fearful while the player holds the hammer.
func drawFireball(f field, fb fireball, scared bool, tick int64) {
	x, y := float32(fb.x), float32(fb.y)
	outer, mid, core := flameOuter, flameMid, flameCore
	if scared {
		outer, mid, core = scaredOut, mix(scaredOut, scaredCore, 0.5), scaredCore
	}
	flick := float32(tick/3%2) * 1.2
	f.poly(outer, x+3, y+14, x+3, y+8, x+5, y+5, x+6, y+2-flick, x+8, y+5, x+10, y+1+flick, x+12, y+7, x+12, y+14)
	f.ellipse(x+7.5, y+10.5, 4.5, 3.8, mid)
	f.ellipse(x+7.5, y+11.5, 2.6, 2.4, core)
	eyeY := y + 8
	if scared {
		eyeY += 0.8
	}
	f.rect(x+5.2, eyeY, 1.4, 2, eyeColour)
	f.rect(x+8.6, eyeY, 1.4, 2, eyeColour)
}

// drawHammer draws a sledgehammer head over the rectangle the rules test,
// with its handle back toward hx, hy.
func drawHammer(f field, x, y, w, h, hx, hy float32, warning bool) {
	head, dark := hammerSteel, hammerDark
	if warning {
		head, dark = warnColour, mix(warnColour, eyeColour, 0.4)
	}
	f.line(hx, hy, x+w/2, y+h/2, 1.6, handle)
	f.roundRect(x, y, w, h, 1, dark)
	f.roundRect(x+0.6, y+0.6, w-1.2, h-1.6, 0.8, head)
}

// drawPickup draws a hammer waiting to be collected: X..X+9, Y..Y+7.
func drawPickup(f field, x, y float32) {
	f.rect(x+4, y+3, 2, 5, handle)
	f.roundRect(x, y, 10, 4, 1, hammerDark)
	f.roundRect(x+0.5, y+0.4, 9, 2.6, 0.8, hammerSteel)
}

// drawRivet draws a removable rivet as a bolt head on its support.
func drawRivet(f field, x, y float32) {
	f.rect(x+2, y-2, 6, 9, boltShade)
	f.rect(x+2.5, y-2, 5, 8, boltHead)
	f.rect(x+3.5, y-1, 3, 1, mix(boltHead, rgb(0xffffff), 0.5))
	f.rect(x+2, y+2.5, 6, 0.8, boltShade)
}

// drawItem draws a lost item, X+1..X+11, Y..Y+9: a lunchbox, a thermos or
// a toolbox, by its index.
func drawItem(f field, i int, x, y float32) {
	switch i % 3 {
	case 0:
		f.roundRect(x+1, y+3, 11, 7, 1.5, lunchbox)
		f.rect(x+1, y+5.5, 11, 0.8, mix(lunchbox, eyeColour, 0.4))
		f.line(x+4, y+3, x+5, y+1, 0.8, hoop)
		f.line(x+5, y+1, x+8, y+1, 0.8, hoop)
		f.line(x+8, y+1, x+9, y+3, 0.8, hoop)
	case 1:
		f.roundRect(x+4, y+1, 5, 9, 1.5, thermos)
		f.rect(x+4, y, 5, 2, hoopLight)
		f.rect(x+4, y+6, 5, 0.8, mix(thermos, eyeColour, 0.4))
	default:
		f.roundRect(x+1, y+4, 11, 6, 1, toolbox)
		f.rect(x+1, y+6, 11, 0.8, mix(toolbox, eyeColour, 0.4))
		f.rect(x+4, y+2, 5, 1, hoop)
		f.rect(x+4, y+2, 1, 2, hoop)
		f.rect(x+8, y+2, 1, 2, hoop)
	}
	f.rect(x+2, y+4, 2, 1, itemLight)
}

// drawSpring draws a bouncing spring: X+2..X+13, Y+2..Y+13, squashed
// while it compresses.
func drawSpring(f field, p spring) {
	x, y := float32(p.x), float32(p.y)
	h := float32(12)
	if p.compressing {
		h = 8
	}
	top := y + 14 - h
	f.roundRect(x+2, top, 12, 2, 0.8, springDark)
	f.roundRect(x+2, y+12, 12, 2, 0.8, springDark)
	coils := 4
	step := (h - 4) / float32(coils)
	for i := range coils {
		cy := top + 2 + float32(i)*step
		f.line(x+3, cy, x+13, cy+step/2, 1.1, springCoil)
		f.line(x+13, cy+step/2, x+3, cy+step, 1.1, springCoil)
	}
}

// drawPie draws a pie on its tin: X+2..X+11, Y+2..Y+5.
func drawPie(f field, x, y float32) {
	f.poly(pieTin, x+1.5, y+3.5, x+12.5, y+3.5, x+11.5, y+6, x+2.5, y+6)
	f.ellipse(x+7, y+3.2, 5, 1.8, pieCrust)
	f.ellipse(x+7, y+2.8, 3.2, 0.9, pieFill)
}

// drawSmash draws the burst of a smashed hazard at its frame.
func drawSmash(f field, x, y float32, frame int) {
	r := float32(3 + frame)
	cx, cy := x+5, y+5
	var pts []float32
	for i := range 16 {
		a := float32(i) * 0.3927
		k := r
		if i%2 == 1 {
			k = r * 0.45
		}
		pts = append(pts, cx+k*cos(a), cy+k*sin(a))
	}
	f.poly(flameMid, pts...)
	f.circle(cx, cy, r*0.35, flameCore)
}
