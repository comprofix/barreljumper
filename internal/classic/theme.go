package classic

import "github.com/comprofix/barreljumper/internal/audio"

// The title theme: an original construction-site loop for the title, the
// menus and the high scores (DESIGN section 6, Title, attract and music).
// Play itself has no music. Eight bars of D minor, Bb, C, F, G minor and a
// turnaround on A, with a hammer-on-steel clank in the drums.
var (
	themeLead = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(10), Sustain: 0.55, Release: audio.Ticks(3), Volume: 0.7}
	themePad  = audio.Voice{Wave: audio.Pulse12, Attack: audio.Ticks(3), Release: audio.Ticks(6), Volume: 0.3}
	themeBass = audio.Voice{Wave: audio.Triangle, Decay: audio.Ticks(12), Sustain: 0.7, Release: audio.Ticks(2)}
	themeKit  = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(5), Release: audio.Ticks(1), Volume: 0.4}

	themeTune = &audio.Tune{Name: "title theme", Loop: true, Parts: []audio.Part{
		{Voice: themeLead, Notes: audio.Notes(
			"D5:12 F5:12 A5:24 G5:12 F5:12 E5:12 D5:12",
			"A4:12 D5:12 F5:12 A5:12 C6:24 A5:24",
			"Bb5:24 A5:12 G5:12 F5:24 D5:24",
			"E5:12 G5:12 C6:24 Bb5:12 G5:12 E5:24",
			"D6:24 C6:12 A5:12 F5:24 A5:24",
			"C6:12 A5:12 F5:12 A5:12 C6:24 F6:24",
			"D6:12 Bb5:12 G5:12 Bb5:12 D6:24 C6:12 Bb5:12",
			"A5:24 C#6:24 E6:24 -:24",
		)},
		{Voice: themePad, Notes: audio.Notes(
			"F4:48 A4:48", "F4:48 A4:48", "D4:48 F4:48", "E4:48 G4:48",
			"F4:48 A4:48", "A4:48 C5:48", "Bb4:48 D5:48", "C#5:48 E5:48",
		)},
		{Voice: themeBass, Notes: audio.Notes(
			"D3:24 A2:24 D3:24 A2:24",
			"D3:24 F3:24 A3:24 F3:24",
			"Bb2:24 F3:24 Bb2:24 F3:24",
			"C3:24 G3:24 C3:24 E3:24",
			"D3:24 A2:24 D3:24 A2:24",
			"F2:24 C3:24 F3:24 C3:24",
			"G2:24 D3:24 G3:24 D3:24",
			"A2:24 E3:24 A2:24 C#3:24",
		)},
		{Voice: themeKit, Notes: themeDrums(8)},
	}}
)

// themeDrums is n bars of kick, hats, snare and, on the last eighth, a
// high metallic clank, as noise at different rates.
func themeDrums(n int) []audio.Note {
	const (
		kick  = 280
		snare = 3200
		hat   = 12000
		clank = 20000
	)
	var out []audio.Note
	for range n {
		for _, hz := range []float64{kick, hat, snare, hat, kick, kick, snare, clank} {
			out = append(out, audio.Note{Hz: hz, Dur: audio.Ticks(12)})
		}
	}
	return out
}
