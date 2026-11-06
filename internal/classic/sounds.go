package classic

import (
	"math"

	"github.com/comprofix/barreljumper/internal/audio"
)

// Classic sounds, all original (DESIGN section 6, M5 audio decision):
// effects for walking, climbing, jumping, the hammer, smashes, awards and
// death, and short jingles for a screen's start and clear. There is no
// looping music.
//
// The effects are short square-wave blips and sweeps in the style of the
// owner's Alley Cat and Night Crossing, with noise kept for impacts.
// Footsteps and rungs sound on the same 4-pixel steps the worker's legs and
// arms are drawn on, so each sound lands with its frame, and the death
// sounds follow the death sequence.

// sound is one gameplay event that makes a sound; a stage collects them
// as a bit set while it updates.
type sound uint32

const (
	soundStep sound = 1 << iota
	soundStepB
	soundClimb
	soundClimbB
	soundJump
	soundLand
	soundHammer
	soundSwing
	soundSmash
	soundPoints
	soundHit
	soundSpin
	soundThud
	soundWinch  // the crane hoisting
	soundClank  // a girder set down
	soundRumble // the Rivets floors dropping
	soundWarn   // the bonus is running low
)

// blip is a square voice that dies away within a tick or two.
func blip(wave audio.Wave, volume float64) audio.Voice {
	return audio.Voice{Wave: wave, Decay: audio.Ticks(2), Release: audio.Ticks(1), Volume: volume}
}

// sweep is steps notes gliding from one pitch to another, each ticks long.
func sweep(from, to float64, steps, ticks int) []audio.Note {
	n := make([]audio.Note, steps)
	for i := range n {
		t := float64(i) / float64(max(steps-1, 1))
		n[i] = audio.Note{Hz: from * math.Pow(to/from, t), Dur: audio.Ticks(ticks)}
	}
	return n
}

// warble is a falling sweep that flips up a fifth every other tick, for
// the death spin.
func warble(from, to float64, ticks int) []audio.Note {
	n := make([]audio.Note, ticks)
	for i := range n {
		hz := from * math.Pow(to/from, float64(i)/float64(ticks-1))
		if i%2 == 1 {
			hz *= 1.5
		}
		n[i] = audio.Note{Hz: hz, Dur: audio.Ticks(1)}
	}
	return n
}

var (
	square = audio.Voice{Wave: audio.Pulse50, Sustain: 1, Release: audio.Ticks(1), Volume: 0.45}
	thin   = audio.Voice{Wave: audio.Pulse25, Sustain: 1, Release: audio.Ticks(1), Volume: 0.45}
	crack  = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(10), Release: audio.Ticks(2), Volume: 0.7}
)

// Effects. Footsteps and rungs alternate two pitches, one per foot or hand.
var (
	stepSound = &audio.Effect{Name: "step", Channel: 3, Part: audio.Part{
		Voice: blip(audio.Pulse50, 0.3), Notes: audio.Notes("G3:2"),
	}}
	stepSoundB = &audio.Effect{Name: "step 2", Channel: 3, Part: audio.Part{
		Voice: blip(audio.Pulse50, 0.3), Notes: audio.Notes("D4:2"),
	}}
	climbSound = &audio.Effect{Name: "rung", Channel: 3, Part: audio.Part{
		Voice: blip(audio.Pulse25, 0.3), Notes: audio.Notes("A4:2"),
	}}
	climbSoundB = &audio.Effect{Name: "rung 2", Channel: 3, Part: audio.Part{
		Voice: blip(audio.Pulse25, 0.3), Notes: audio.Notes("E5:2"),
	}}
	jumpSound = &audio.Effect{Name: "jump", Channel: 1, Priority: 2, Part: audio.Part{
		Voice: thin, Notes: sweep(330, 990, 8, 1),
	}}
	landSound = &audio.Effect{Name: "land", Channel: 3, Priority: 1, Part: audio.Part{
		Voice: blip(audio.Pulse50, 0.4), Notes: audio.Notes("C3:3"),
	}}
	hammerSound = &audio.Effect{Name: "hammer pickup", Channel: 1, Priority: 3, Part: audio.Part{
		Voice: thin, Notes: audio.Notes("C5:2 E5:2 G5:2 C6:2 E6:2 G6:6"),
	}}
	swingSound = &audio.Effect{Name: "hammer swing", Channel: 2, Priority: 1, Part: audio.Part{
		Voice: blip(audio.Pulse50, 0.4), Notes: audio.Notes("C4:1 G3:2"),
	}}
	smashSound = &audio.Effect{Name: "smash", Channel: 1, Priority: 3, Part: audio.Part{
		Voice: square, Notes: sweep(1400, 180, 14, 1),
	}}
	smashCrack = &audio.Effect{Name: "smash crack", Channel: 3, Priority: 3, Part: audio.Part{
		Voice: crack, Notes: []audio.Note{{Hz: 3000, Dur: audio.Ticks(3)}, {Hz: 1200, Dur: audio.Ticks(6)}},
	}}
	pointsSound = &audio.Effect{Name: "points", Channel: 0, Priority: 2, Part: audio.Part{
		Voice: thin, Notes: audio.Notes("C6:2 G6:5"),
	}}
	hitSound = &audio.Effect{Name: "hit", Channel: 3, Priority: 4, Part: audio.Part{
		Voice: crack, Notes: []audio.Note{{Hz: 2400, Dur: audio.Ticks(3)}, {Hz: 900, Dur: audio.Ticks(8)}},
	}}
	// spinSound lasts the death spin (140 counter ticks, about 1.9 s).
	spinSound = &audio.Effect{Name: "death spin", Channel: 0, Priority: 4, Part: audio.Part{
		Voice: square, Notes: warble(880, 150, 115),
	}}
	thudSound = &audio.Effect{Name: "fall thud", Channel: 3, Priority: 4, Part: audio.Part{
		Voice: crack, Notes: []audio.Note{{Hz: 300, Dur: audio.Ticks(4)}, {Hz: 150, Dur: audio.Ticks(12)}},
	}}
	// winchSound is the crane's ratchet, a rising run of short blips.
	winchSound = &audio.Effect{Name: "crane winch", Channel: 2, Priority: 2, Part: audio.Part{
		Voice: blip(audio.Pulse25, 0.3), Notes: sweep(150, 330, 16, 2),
	}}
	clankSound = &audio.Effect{Name: "girder clank", Channel: 3, Priority: 2, Part: audio.Part{
		Voice: crack, Notes: []audio.Note{{Hz: 2000, Dur: audio.Ticks(2)}, {Hz: 700, Dur: audio.Ticks(5)}},
	}}
	// warnSound repeats while the bonus is below 1000: two short falling
	// blips, quiet enough to sit under play.
	warnSound = &audio.Effect{Name: "low bonus", Channel: 0, Priority: 1, Part: audio.Part{
		Voice: blip(audio.Pulse25, 0.28), Notes: audio.Notes("B5:3 F5:4"),
	}}
	rumbleSound = &audio.Effect{Name: "floors rumble", Channel: 3, Priority: 4, Part: audio.Part{
		Voice: audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(40), Release: audio.Ticks(4), Volume: 0.7},
		Notes: []audio.Note{{Hz: 220, Dur: audio.Ticks(20)}, {Hz: 120, Dur: audio.Ticks(30)}},
	}}
)

// Jingles, in the same square voices.
var (
	jingleLead = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(8), Sustain: 0.5, Release: audio.Ticks(2), Volume: 0.5}
	jingleHarm = audio.Voice{Wave: audio.Pulse12, Decay: audio.Ticks(8), Sustain: 0.4, Release: audio.Ticks(2), Volume: 0.35}
	jingleBass = audio.Voice{Wave: audio.Triangle, Decay: audio.Ticks(10), Sustain: 0.6, Release: audio.Ticks(2), Volume: 0.8}

	// startJingle plays as a screen begins: a short call to climb.
	startJingle = &audio.Tune{Name: "screen start", Parts: []audio.Part{
		{Voice: jingleLead, Notes: audio.Notes("G4:4 C5:4 E5:4 G5:8 E5:4 G5:12")},
		{Voice: jingleHarm, Notes: audio.Notes("E4:4 G4:4 C5:4 E5:8 C5:4 E5:12")},
		{Voice: jingleBass, Notes: audio.Notes("C3:12 G2:8 C3:16")},
	}}
	// clearJingle plays when a screen is cleared.
	clearJingle = &audio.Tune{Name: "screen clear", Parts: []audio.Part{
		{Voice: jingleLead, Notes: audio.Notes("C5:4 E5:4 G5:4 C6:4 G5:4 C6:4 E6:16")},
		{Voice: jingleHarm, Notes: audio.Notes("G4:4 C5:4 E5:4 G5:4 E5:4 G5:4 C6:16")},
		{Voice: jingleBass, Notes: audio.Notes("C3:8 G2:8 C3:8 C2:16")},
	}}
)

// TitleTune is the title theme, for the title, menus and high scores.
var TitleTune = themeTune

// Sounds are the Classic sounds, for the sound test.
var Sounds = audio.Bank{
	Tunes: []*audio.Tune{themeTune, startJingle, clearJingle},
	Effects: []*audio.Effect{stepSound, stepSoundB, climbSound, climbSoundB, jumpSound, landSound,
		hammerSound, swingSound, smashSound, smashCrack, pointsSound, hitSound, spinSound, thudSound,
		winchSound, clankSound, rumbleSound, warnSound},
}

// effects maps each event to its effects, least important first, so the
// most important plays last on its channel.
var effects = []struct {
	event   sound
	effects []*audio.Effect
}{
	{soundWarn, []*audio.Effect{warnSound}},
	{soundStep, []*audio.Effect{stepSound}},
	{soundStepB, []*audio.Effect{stepSoundB}},
	{soundClimb, []*audio.Effect{climbSound}},
	{soundClimbB, []*audio.Effect{climbSoundB}},
	{soundLand, []*audio.Effect{landSound}},
	{soundSwing, []*audio.Effect{swingSound}},
	{soundJump, []*audio.Effect{jumpSound}},
	{soundPoints, []*audio.Effect{pointsSound}},
	{soundHammer, []*audio.Effect{hammerSound}},
	{soundSmash, []*audio.Effect{smashSound, smashCrack}},
	{soundHit, []*audio.Effect{hitSound}},
	{soundSpin, []*audio.Effect{spinSound}},
	{soundThud, []*audio.Effect{thudSound}},
	{soundWinch, []*audio.Effect{winchSound}},
	{soundClank, []*audio.Effect{clankSound}},
	{soundRumble, []*audio.Effect{rumbleSound}},
}

// play plays the effects for the events in set.
func play(a *audio.Synth, set sound) {
	for _, e := range effects {
		if set&e.event == 0 {
			continue
		}
		for _, fx := range e.effects {
			a.PlayEffect(fx)
		}
	}
}

// footfall is the step a coordinate is on: the worker's legs and arms
// change, and a step or rung sounds, every 4 pixels.
func footfall(v int) int { return v >> 2 }

// alternate picks a for one step and b for the next.
func alternate(v int, a, b sound) sound {
	if footfall(v)&1 == 0 {
		return a
	}
	return b
}
