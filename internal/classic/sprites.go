package classic

import "image/color"

// The construction-site sprites. Each function notes the collision
// rectangle it keeps to.

// drawGirder draws a painted steel I-beam on the support's 7-pixel band,
// with a soft shadow on whatever is below it.
func drawGirder(f field, g girder) {
	if g.right <= g.left {
		return
	}
	drawBeam(f, float32(g.left), float32(g.y), float32(g.right-g.left))
}

// drawBeam draws a girder w wide with its top left at x, y.
func drawBeam(f field, x, y, w float32) {
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
	white, black := rgb(0xffffff), rgb(0x101018)
	shirtLight, shirtDark := mix(bruteShirt, white, 0.22), mix(bruteShirt, black, 0.3)
	bibLight, bibDark := mix(bruteBib, white, 0.2), mix(bruteBib, black, 0.35)
	skinLight, skinDark := mix(bruteSkin, white, 0.25), mix(bruteSkin, black, 0.3)
	f.ellipse(x+20, y+31.2, 17, 1.6, shadowColour)

	// Legs in the overalls, creased, and steel-toed boots.
	for _, lx := range []float32{x + 9, x + 23} {
		f.rect(lx, y+23, 8, 7, bruteBib)
		f.rect(lx, y+23, 2, 7, bibLight)
		f.rect(lx+6, y+23, 2, 7, bibDark)
		f.rect(lx+3.8, y+25, 0.4, 4, bibDark)
		f.roundRect(lx-2, y+28.6, 12, 3.4, 1.2, boot)
		f.roundRect(lx-2, y+28.6, 4.5, 2.4, 1.1, mix(boot, white, 0.25)) // the toe cap
		f.rect(lx-2, y+31.2, 12, 0.8, black)                             // the sole
		f.line(lx+3, y+29.2, lx+7, y+29.2, 0.35, mix(boot, white, 0.4))  // laces
	}

	// The barrel chest in a work shirt, lit from the left.
	f.roundRect(x+5, y+8, 30, 17, 6, bruteShirt)
	f.roundRect(x+5, y+8, 10, 15, 5, shirtLight)
	f.roundRect(x+25, y+9, 10, 15, 5, shirtDark)
	f.roundRect(x+9, y+8, 22, 15, 5, bruteShirt)
	f.ellipse(x+14, y+11, 4, 2, shirtLight) // the chest
	f.ellipse(x+26, y+11, 4, 2, mix(bruteShirt, white, 0.1))
	// The bib of the overalls: straps, buckles, a pocket with a pencil.
	f.roundRect(x+10, y+14, 20, 11, 2, bruteBib)
	f.rect(x+10, y+14, 3, 11, bibLight)
	f.rect(x+27, y+14, 3, 11, bibDark)
	f.rect(x+12, y+8.5, 2.6, 6.5, bruteBib)
	f.rect(x+25.4, y+8.5, 2.6, 6.5, bibDark)
	for _, bx := range []float32{x + 13.3, x + 26.7} {
		f.roundRect(bx-1.4, y+14.2, 2.8, 2, 0.5, hammerDark)
		f.roundRect(bx-1, y+14.5, 2, 1.3, 0.4, hammerSteel)
	}
	f.roundRect(x+16, y+17, 8, 5, 1, bibDark)
	f.rect(x+16, y+17, 8, 1, bibLight)
	f.rect(x+21, y+15, 1, 3, hazardYellow)
	f.rect(x+21, y+14.6, 1, 0.6, rgb(0xe08080))
	f.line(x+10.6, y+24.4, x+29.4, y+24.4, 0.3, bibLight) // stitching

	// The head, small for the shoulders: ears, a broad jaw with stubble,
	// heavy brows and a scowl.
	f.rect(x+17.5, y+9, 5, 2, skinDark) // the neck
	f.circle(x+14.6, y+6, 1.3, skinDark)
	f.circle(x+25.4, y+6, 1.3, skinDark)
	f.circle(x+20, y+5.8, 5.4, bruteSkin)
	f.roundRect(x+15.4, y+6, 9.2, 5.2, 2.4, bruteSkin)
	f.ellipse(x+17.6, y+5, 2.2, 2.6, skinLight)
	f.ellipse(x+20, y+9.6, 4.2, 1.7, mix(bruteSkin, eyeColour, 0.16)) // stubble along the jaw
	for i := float32(0); i < 4; i++ {
		f.circle(x+17+i*2, y+10.2, 0.2, mix(bruteSkin, eyeColour, 0.4))
	}
	f.rect(x+17.5, y+9, 5, 0.7, rgb(0x5a2a20))                               // a hard mouth
	f.poly(skinDark, x+19.4, y+5, x+20.6, y+5, x+21.2, y+7.6, x+18.8, y+7.6) // the nose
	f.rect(x+19.4, y+5, 0.6, 2.4, skinLight)
	f.rect(x+16.4, y+4.6, 2, 1.3, white)
	f.rect(x+21.6, y+4.6, 2, 1.3, white)
	f.rect(x+17.2, y+4.7, 1, 1.1, eyeColour)
	f.rect(x+21.8, y+4.7, 1, 1.1, eyeColour)
	f.poly(captiveHair, x+15.4, y+3.2, x+19.2, y+4, x+19.2, y+4.8, x+15.6, y+4.2) // the brows
	f.poly(captiveHair, x+24.6, y+3.2, x+20.8, y+4, x+20.8, y+4.8, x+24.4, y+4.2)

	// The red hard hat, shaded round, with its ridge, brim and a dent.
	hatLight, hatDark := mix(bruteHat, white, 0.35), mix(bruteHat, black, 0.35)
	f.ellipse(x+20, y+1.6, 6.6, 3.2, hatDark)
	f.ellipse(x+19.6, y+1.3, 6, 2.8, bruteHat)
	f.ellipse(x+17.6, y+0.4, 2.6, 1.2, hatLight)
	f.rect(x+19.5, y-1.5, 1, 4, hatDark)
	f.roundRect(x+12.4, y+2.4, 15.2, 1.6, 0.8, hatDark)
	f.rect(x+12.4, y+2.4, 15.2, 0.5, bruteHat)
	f.ellipse(x+23.4, y+0.6, 1.1, 0.6, hatDark)

	// The arms: rolled sleeves, shaded forearms and fists, by pose.
	type limb struct{ sx, sy, ex, ey, hx, hy float32 }
	left := limb{6, 11, 2.5, 17.5, 3.5, 23.5}
	right := limb{34, 11, 37.5, 17.5, 36.5, 23.5}
	switch t.pose {
	case throwerPickUp:
		left = limb{6, 11, 2, 18, 3, 26}
		right = limb{34, 11, 37.5, 16, 37, 21}
	case throwerThrow:
		right = limb{34, 11, 38, 12, 41, 12.5}
	case throwerTauntA:
		left = limb{6, 11, 1.5, 7, 3, 1.5}
	case throwerTauntB:
		right = limb{34, 11, 38.5, 7, 37, 1.5}
	}
	if t.holding {
		left = limb{6, 11, 4, 17, 11.5, 20.5}
		right = limb{34, 11, 36, 17, 28.5, 20.5}
	}
	arm := func(l limb, lit bool) {
		sleeve, fore := bruteShirt, bruteSkin
		if lit {
			sleeve, fore = shirtLight, skinLight
		}
		sx, sy, ex, ey, hx, hy := x+l.sx, y+l.sy, x+l.ex, y+l.ey, x+l.hx, y+l.hy
		f.line(sx, sy, ex, ey, 6.2, shirtDark)
		f.line(sx, sy, ex, ey, 5, sleeve)
		f.circle(sx, sy, 3, sleeve)
		f.circle(ex, ey, 2.7, shirtDark) // the rolled cuff
		f.line(ex, ey, hx, hy, 4.4, skinDark)
		f.line(ex, ey, hx, hy, 3.4, fore)
		f.circle(hx, hy, 2.8, skinDark)
		f.circle(hx, hy, 2.3, bruteSkin)
		f.circle(hx-0.6, hy-0.7, 0.9, skinLight)
		for i := float32(-1); i <= 1; i++ { // knuckles
			f.circle(hx+i*0.9, hy+1.4, 0.35, skinDark)
		}
	}
	if t.holding {
		drawSpecialBarrel(f, x+11, y+16)
	}
	arm(left, true)
	arm(right, false)
}

// drawCaptive draws the site architect waiting to be rescued, in the
// 10 x 17 box at (x, y), waving a rolled blueprint.
func drawCaptive(f field, x, y float32, tick int64) {
	jacket, jacketDark, jacketLight := captiveShirt, mix(captiveShirt, eyeColour, 0.35), mix(captiveShirt, rgb(0xffffff), 0.3)
	skirt, skirtDark := rgb(0x2e3450), rgb(0x1e2238)
	hairLight := mix(captiveHair, rgb(0xd08850), 0.45)
	wave := float32((tick/20)%2) * 1.6
	f.ellipse(x+5, y+17, 5, 1, shadowColour)
	// Legs, shoes and a pencil skirt.
	f.rect(x+3, y+12.5, 1.6, 3.6, skin)
	f.rect(x+5.6, y+12.5, 1.6, 3.6, skinShade)
	f.roundRect(x+2.4, y+15.6, 2.6, 1.4, 0.6, rgb(0x6a2030))
	f.roundRect(x+5.2, y+15.6, 2.6, 1.4, 0.6, rgb(0x6a2030))
	f.poly(skirt, x+2.2, y+10.4, x+8, y+10.4, x+7.6, y+13.4, x+2.6, y+13.4)
	f.rect(x+5.6, y+10.4, 2.4, 3, skirtDark)
	// The jacket: lit on one side, lapels open over a white blouse.
	f.roundRect(x+1.6, y+5.4, 7, 5.6, 1.6, jacket)
	f.rect(x+6.4, y+5.8, 2, 5, jacketDark)
	f.rect(x+1.8, y+5.8, 1.2, 4.6, jacketLight)
	f.poly(vestStripe, x+4.2, y+5.4, x+5.8, y+5.4, x+5, y+7.6)
	f.line(x+4.2, y+5.4, x+4.8, y+8, 0.5, jacketDark)
	f.line(x+5.8, y+5.4, x+5.2, y+8, 0.5, jacketDark)
	f.circle(x+5, y+8.8, 0.35, jacketLight)
	f.rect(x+1.8, y+10.2, 6.6, 0.7, rgb(0x3a2418)) // belt
	// The left arm at her side; the right waves the rolled blueprint.
	f.line(x+2, y+6.4, x+1.4, y+9.8, 1.3, jacket)
	f.circle(x+1.4, y+10.2, 0.7, skin)
	f.line(x+8, y+6.2, x+9.4, y+3.2+wave, 1.3, jacket)
	f.circle(x+9.6, y+2.8+wave, 0.8, skin)
	f.roundRect(x+9, y-2.6+wave, 2.4, 6.4, 1, blueprint)
	f.rect(x+9.3, y-1.4+wave, 1.8, 0.4, vestStripe)
	f.rect(x+9.3, y+0.4+wave, 1.8, 0.4, vestStripe)
	f.ellipse(x+10.2, y-2.6+wave, 1.2, 0.5, mix(blueprint, rgb(0xffffff), 0.5))
	// Long hair falling behind the shoulders, the neck, the face, then the
	// fringe over it.
	f.roundRect(x+2, y+0.8, 1.6, 6.2, 0.8, captiveHair)
	f.roundRect(x+6.4, y+0.8, 1.6, 6.2, 0.8, captiveHair)
	f.rect(x+4.3, y+5, 1.4, 1, skinShade)
	f.circle(x+5, y+3.1, 2.4, skin)
	f.ellipse(x+5, y+1.2, 2.9, 1.6, captiveHair)
	f.rect(x+2.1, y+1.2, 1, 3.2, captiveHair)
	f.rect(x+6.9, y+1.2, 1, 3.2, captiveHair)
	f.ellipse(x+4.2, y+0.9, 1.4, 0.6, hairLight)
	f.rect(x+3.9, y+2.9, 0.8, 0.9, eyeColour)
	f.rect(x+5.5, y+2.9, 0.8, 0.9, eyeColour)
	f.circle(x+3.6, y+4.1, 0.45, rgb(0xe88080)) // blush
	f.circle(x+6.6, y+4.1, 0.45, rgb(0xe88080))
	f.ellipse(x+5, y+4.5, 0.55, 0.4, rgb(0x9a3040)) // calling for help
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

// drawHammer draws a sledgehammer over the rectangle the rules test, with
// its handle back to the hands at hx, hy. Overhead, the head lies across
// the top of the box; swung forward, it stands upright at its front.
func drawHammer(f field, x, y, w, h, hx, hy float32, forward, warning bool) {
	if !forward {
		top := y + h*0.1
		drawHandle(f, x+w/2, top+h*0.75, hx, hy)
		drawSledge(f, x, top, w, h*0.8, true, warning)
		return
	}
	left := x + w*0.2
	drawHandle(f, left+w*0.3, y+h/2, hx, hy)
	drawSledge(f, left, y, w*0.6, h, false, warning)
}

// drawHandle draws a hickory handle from the head's collar to the grip,
// lit along one edge, with a taped grip at the hands' end.
func drawHandle(f field, x0, y0, x1, y1 float32) {
	f.line(x0, y0, x1, y1, 2.0, woodDark)
	f.line(x0, y0, x1, y1, 1.3, handle)
	f.line(x0-0.25, y0-0.25, x1-0.25, y1-0.25, 0.4, woodLight)
	gx, gy := x0+(x1-x0)*0.68, y0+(y1-y0)*0.68
	f.line(gx, gy, x1, y1, 2.2, rgb(0x24242a))
	for t := float32(0.72); t < 1; t += 0.09 {
		cx, cy := x0+(x1-x0)*t, y0+(y1-y0)*t
		f.circle(cx, cy, 0.55, rgb(0x46464e))
	}
}

// drawSledge draws a sledgehammer head in the box x, y, w, h: a forged
// steel block shaded round its length like a bar lit from above, with
// bevelled striking faces at both ends and a collar where the handle
// enters. across says the head lies along x; otherwise it stands along y.
func drawSledge(f field, x, y, w, h float32, across, warning bool) {
	base := hammerSteel
	if warning {
		base = warnColour
	}
	white, black := rgb(0xffffff), rgb(0x101018)
	outline := mix(base, black, 0.6)
	bands := []color.RGBA{mix(base, white, 0.7), mix(base, white, 0.35), base, mix(base, black, 0.2), mix(base, black, 0.38)}
	face, faceLight := mix(base, black, 0.15), mix(base, white, 0.5)
	// Local coordinates: u along the head, v across it.
	length, thick := w, h
	if !across {
		length, thick = h, w
	}
	at := func(u, v float32) (float32, float32) {
		if across {
			return x + u, y + v
		}
		return x + v, y + u
	}
	quad := func(c color.Color, u0, v0, u1, v1 float32) {
		ax, ay := at(u0, v0)
		bx, by := at(u1, v0)
		cx, cy := at(u1, v1)
		dx, dy := at(u0, v1)
		f.poly(c, ax, ay, bx, by, cx, cy, dx, dy)
	}
	e := min(1.8, length*0.17) // the bevel's depth
	quad(outline, 0, 0, length, thick)
	n := float32(len(bands))
	for i, c := range bands {
		v0 := 0.4 + (thick-0.8)*float32(i)/n
		quad(c, e, v0, length-e, v0+(thick-0.8)/n+0.05)
	}
	// The striking faces, narrowed by the bevel, lit at their upper edge.
	for _, end := range [][2]float32{{0.35, e}, {length - 0.35, length - e}} {
		o, i := end[0], end[1]
		ax, ay := at(i, 0.4)
		bx, by := at(i, thick-0.4)
		cx, cy := at(o, thick-1)
		dx, dy := at(o, 1)
		f.poly(face, ax, ay, bx, by, cx, cy, dx, dy)
		lx, ly := at(o, 1)
		mx, my := at(i, 0.4)
		f.line(lx, ly, mx, my, 0.45, faceLight)
	}
	// A glint on the top band, and the collar's wedge on the handle side.
	gx, gy := at(length*0.3, thick*0.18)
	f.ellipse(gx, gy, 0.9, 0.35, white)
	quad(outline, length/2-1.6, thick-1.1, length/2+1.6, thick)
	quad(mix(base, black, 0.3), length/2-1.1, thick-0.9, length/2+1.1, thick-0.2)
}

// drawPickup draws a sledgehammer waiting to be collected, standing on
// its head's end: X..X+9, Y..Y+7.
func drawPickup(f field, x, y float32) {
	f.ellipse(x+5, y+8, 5, 0.8, shadowColour)
	drawHandle(f, x+5, y+3.5, x+5, y+8)
	drawSledge(f, x, y, 10, 4.2, true, false)
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

// drawSmash draws a smashed hazard's explosion through its 13 frames: a
// flash and shock ring, a rolling fireball that darkens as it burns out,
// debris of whatever was hit thrown out under gravity, sparks, and smoke
// rising and thinning at the end.
func drawSmash(f field, k smash) {
	age := float32(k.frame) + float32(min(k.timer, 6))/7
	cx, cy := float32(k.x)+5, float32(k.y)+4
	white := rgb(0xffffff)
	outer, mid, core := flameOuter, flameMid, flameCore
	if k.kind == smashedFlame {
		outer, mid, core = scaredOut, blueLight, scaredCore
	}
	// Smoke, behind the rest: puffs that swell, rise and thin out.
	if age > 2.5 {
		t := age - 2.5
		for i := range 5 {
			fi := float32(i)
			px := cx + (fi-2)*3.2 + 0.8*sin(fi*1.7+t*0.6)
			py := cy - 1 - t*1.3 - 1.5*float32(i%2)
			r := 2.2 + t*0.55 + 0.4*fi*float32(i%3)
			fade := clamp01(t / 9)
			smoke := mix(rgb(0x5a5a68), skyTop, fade)
			f.circle(px+0.6, py+0.6, r, mix(smoke, eyeColour, 0.25))
			f.circle(px, py, r, smoke)
			f.circle(px-r*0.3, py-r*0.3, r*0.45, mix(smoke, white, 0.12*(1-fade)))
		}
	}
	// The shock ring.
	if age < 4 {
		r := 3 + age*4.5
		ring := mix(core, skyTop, age/4)
		for i := range 20 {
			a := float32(i) * 0.31416
			f.circle(cx+r*cos(a), cy+r*0.8*sin(a), 0.55*(1-age/4)+0.2, ring)
		}
	}
	// The fireball: lumpy layers that grow, roll and burn down.
	if age < 9 {
		size := 9 * smooth(age/2) * (1 - smooth((age-4.5)/4.5))
		burnt := clamp01((age - 3) / 6)
		layers := []struct {
			scale float32
			c     color.RGBA
		}{
			{1, mix(outer, rgb(0x6a2410), burnt)},
			{0.72, mix(mid, outer, burnt)},
			{0.42, mix(core, mid, burnt)},
		}
		if k.kind == smashedFlame {
			layers[0].c = mix(outer, blueDark, burnt)
		}
		for _, l := range layers {
			r := size * l.scale
			if r <= 0.2 {
				continue
			}
			for i := range 7 {
				fi := float32(i)
				a := fi*0.8976 + age*0.25
				lump := r * (0.5 + 0.12*sin(fi*2.3+age))
				f.circle(cx+r*0.45*cos(a), cy+r*0.4*sin(a)-age*0.3, lump, l.c)
			}
			f.circle(cx, cy-age*0.3, r*0.6, l.c)
		}
	}
	// Debris of what was hit, flung out and falling.
	if age > 0.6 && age < 10 {
		t := age - 0.6
		shrink := 1 - clamp01((age-7)/3)
		for i := range 9 {
			fi := float32(i)
			a := fi*0.698 + 0.35*sin(fi*3.1) - 1.57
			v := 3 + float32(i%3)*1.1
			px := cx + cos(a)*v*t
			py := cy + sin(a)*v*t*0.8 + 0.32*t*t
			spin := fi + t*1.6
			dx, dy := cos(spin)*1.8*shrink, sin(spin)*1.8*shrink
			switch k.kind {
			case smashedBarrel:
				if i%3 == 2 {
					f.line(px-dx, py-dy, px+dx, py+dy, 0.7*shrink+0.1, hoopLight) // a bent hoop piece
					f.line(px+dx, py+dy, px+dx*1.4-dy*0.6, py+dy*1.4+dx*0.6, 0.7*shrink+0.1, hoop)
				} else {
					f.line(px-dx, py-dy, px+dx, py+dy, 1.4*shrink+0.1, woodDark) // a stave
					f.line(px-dx*0.9, py-dy*0.9, px+dx*0.9, py+dy*0.9, 0.6*shrink+0.1, woodLight)
				}
			case smashedPie:
				if i%2 == 0 {
					f.line(px-dx, py-dy, px+dx, py+dy, 1.3*shrink+0.1, pieCrust)
				} else {
					f.circle(px, py, 0.9*shrink+0.2, pieFill)
					f.circle(px-0.3, py-0.3, 0.35*shrink+0.1, mix(pieFill, white, 0.4))
				}
			default:
				f.line(px, py, px-cos(a)*2.2, py-sin(a)*1.8, 0.5*shrink+0.1, mix(mid, skyTop, 0.3))
				f.circle(px, py, 0.8*shrink+0.2, core)
			}
		}
	}
	// Sparks, quick and bright, with short trails.
	if age > 0.2 && age < 5 {
		t := age - 0.2
		for i := range 12 {
			fi := float32(i)
			a := fi*0.5236 + 0.2
			v := 5 + float32(i%4)*1.4
			px, py := cx+cos(a)*v*t, cy+sin(a)*v*t*0.85+0.25*t*t
			c := mix(core, mid, t/5)
			f.line(px, py, px-cos(a)*1.6, py-sin(a)*1.4, 0.45, mix(c, skyTop, 0.3))
			f.circle(px, py, 0.5*(1-t/5)+0.15, c)
		}
	}
	// The flash at the hit: a white core and rays.
	if age < 2.2 {
		fl := 1 - age/2.2
		for i := range 8 {
			a := float32(i)*0.7854 + 0.39
			r := 4 + age*5
			f.line(cx+cos(a)*1.5, cy+sin(a)*1.5, cx+cos(a)*r, cy+sin(a)*r, 1.1*fl+0.1, mix(core, white, fl))
		}
		f.circle(cx, cy, 2+age*1.8, mix(core, white, fl))
		f.circle(cx, cy, (1+age)*fl, white)
	}
}
