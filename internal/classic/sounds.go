package classic

import "github.com/comprofix/barreljumper/internal/audio"

// Classic sounds, all original (DESIGN section 6, M5 audio decision):
// effects for walking, climbing, jumping, the hammer, smashes, awards and
// death, and short jingles for a screen's start and clear. There is no
// looping music.

// sound is one gameplay event that makes a sound; a stage collects them
// as a bit set while it updates.
type sound uint16

const (
	soundStep sound = 1 << iota
	soundClimb
	soundJump
	soundHammer
	soundSwing
	soundSmash
	soundPoints
	soundDeath
)

// Voices.
var (
	tick   = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(2), Release: audio.Ticks(1), Volume: 0.25}
	chirp  = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(8), Sustain: 0.4, Release: audio.Ticks(2), Volume: 0.6}
	thud   = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(6), Release: audio.Ticks(2), Volume: 0.5}
	bright = audio.Voice{Wave: audio.Pulse12, Decay: audio.Ticks(6), Sustain: 0.5, Release: audio.Ticks(2), Volume: 0.6}
	horn   = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(14), Sustain: 0.6, Release: audio.Ticks(4), Volume: 0.7}
	low    = audio.Voice{Wave: audio.Triangle, Decay: audio.Ticks(16), Sustain: 0.7, Release: audio.Ticks(3)}
)

// Effects. Channel 3 is noise, so footsteps and impacts share it; the
// others keep to the pulse channels a jingle leaves free.
var (
	stepSound = &audio.Effect{Name: "step", Channel: 3, Priority: 0, Part: audio.Part{
		Voice: tick, Notes: []audio.Note{{Hz: 9000, Dur: audio.Ticks(2)}},
	}}
	climbSound = &audio.Effect{Name: "climb", Channel: 3, Priority: 0, Part: audio.Part{
		Voice: tick, Notes: []audio.Note{{Hz: 5000, Dur: audio.Ticks(2)}},
	}}
	jumpSound = &audio.Effect{Name: "jump", Channel: 1, Priority: 2, Part: audio.Part{
		Voice: chirp, Notes: audio.Notes("C5:3 E5:3 G5:3 C6:6"),
	}}
	hammerSound = &audio.Effect{Name: "hammer pickup", Channel: 1, Priority: 3, Part: audio.Part{
		Voice: bright, Notes: audio.Notes("G5:4 C6:4 E6:4 G6:8"),
	}}
	swingSound = &audio.Effect{Name: "hammer swing", Channel: 3, Priority: 1, Part: audio.Part{
		Voice: thud, Notes: []audio.Note{{Hz: 600, Dur: audio.Ticks(4)}},
	}}
	smashSound = &audio.Effect{Name: "smash", Channel: 3, Priority: 3, Part: audio.Part{
		Voice: thud, Notes: []audio.Note{{Hz: 2500, Dur: audio.Ticks(6)}, {Hz: 900, Dur: audio.Ticks(10)}, {Hz: 300, Dur: audio.Ticks(14)}},
	}}
	pointsSound = &audio.Effect{Name: "points", Channel: 0, Priority: 2, Part: audio.Part{
		Voice: bright, Notes: audio.Notes("E6:3 A6:6"),
	}}
	deathSound = &audio.Effect{Name: "death", Channel: 0, Priority: 4, Part: audio.Part{
		Voice: horn, Notes: audio.Notes("G5:8 F#5:8 F5:8 E5:8 D#5:8 D5:8 C#5:8 C5:24"),
	}}
)

// Jingles.
var (
	// startJingle plays as a screen begins: a rising call to climb.
	startJingle = &audio.Tune{Name: "screen start", Parts: []audio.Part{
		{Voice: horn, Notes: audio.Notes("C5:6 E5:6 G5:6 C6:12 G5:6 C6:18")},
		{Voice: bright, Notes: audio.Notes("-:6 C5:6 E5:6 G5:12 E5:6 G5:18")},
		{Voice: low, Notes: audio.Notes("C3:18 G2:12 C3:24")},
	}}
	// clearJingle plays when a screen is cleared.
	clearJingle = &audio.Tune{Name: "screen clear", Parts: []audio.Part{
		{Voice: horn, Notes: audio.Notes("G5:6 C6:6 E6:6 D6:6 C6:6 D6:6 E6:24")},
		{Voice: bright, Notes: audio.Notes("E5:6 G5:6 C6:6 B5:6 A5:6 B5:6 C6:24")},
		{Voice: low, Notes: audio.Notes("C3:12 E3:12 G3:12 C3:24")},
	}}
)

// Sounds are the Classic sounds, for the sound test.
var Sounds = audio.Bank{
	Tunes: []*audio.Tune{startJingle, clearJingle},
	Effects: []*audio.Effect{stepSound, climbSound, jumpSound, hammerSound, swingSound,
		smashSound, pointsSound, deathSound},
}

// effects maps each event to its effect, in priority order for a frame
// that has several: the most important plays last on its channel.
var effects = []struct {
	event  sound
	effect *audio.Effect
}{
	{soundStep, stepSound}, {soundClimb, climbSound}, {soundSwing, swingSound},
	{soundJump, jumpSound}, {soundPoints, pointsSound}, {soundHammer, hammerSound},
	{soundSmash, smashSound}, {soundDeath, deathSound},
}

// play plays the effects for the events in set.
func play(a *audio.Synth, set sound) {
	for _, e := range effects {
		if set&e.event != 0 {
			a.PlayEffect(e.effect)
		}
	}
}
