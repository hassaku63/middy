// Package fixtures produces the canonical JSON examples that the frontend
// tests check their mirrors of backend logic against (ADR-0007): chord
// symbols, roman numerals and one analysed timeline. The Go side is the
// source of truth; `make fixtures` writes the files and a test fails when
// the committed files drift.
package fixtures

import (
	"encoding/json"
	"fmt"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Files returns every fixture keyed by file name.
func Files() (map[string][]byte, error) {
	out := map[string]any{}

	n, err := notation()
	if err != nil {
		return nil, err
	}
	out["notation.json"] = n

	tl, err := timeline()
	if err != nil {
		return nil, err
	}
	out["timeline.json"] = tl

	files := make(map[string][]byte, len(out))
	for name, v := range out {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("fixtures: %s: %w", name, err)
		}
		files[name] = append(b, '\n')
	}
	return files, nil
}

type pitchClassFixture struct {
	PitchClass music.PitchClass `json:"pitchClass"`
	Name       string           `json:"name"`
}

type chordFixture struct {
	Chord  music.Chord `json:"chord"`
	Symbol string      `json:"symbol"`
}

type romanFixture struct {
	Key     music.Key   `json:"key"`
	Degree  int         `json:"degree"`
	Seventh bool        `json:"seventh"`
	Chord   music.Chord `json:"chord"`
	Numeral string      `json:"romanNumeral"`
}

type notationFixture struct {
	PitchClasses  []pitchClassFixture `json:"pitchClasses"`
	Chords        []chordFixture      `json:"chords"`
	RomanNumerals []romanFixture      `json:"romanNumerals"`
}

var allQualities = []music.ChordQuality{
	music.MajorTriad, music.MinorTriad, music.DiminishedTriad, music.AugmentedTriad,
	music.Major7, music.Minor7, music.Dominant7, music.HalfDiminished7,
}

func notation() (notationFixture, error) {
	var f notationFixture
	for pc := music.PitchClass(0); pc < 12; pc++ {
		f.PitchClasses = append(f.PitchClasses, pitchClassFixture{pc, pc.String()})
		for _, q := range allQualities {
			c := music.Chord{Root: pc, Quality: q}
			f.Chords = append(f.Chords, chordFixture{c, c.String()})
		}
	}
	for _, k := range []music.Key{{Tonic: music.C, Mode: music.Major}, {Tonic: music.A, Mode: music.Minor}} {
		for d := 1; d <= 7; d++ {
			for _, seventh := range []bool{false, true} {
				c, err := k.Diatonic(d, seventh)
				if err != nil {
					return f, err
				}
				rn, err := music.RomanNumeral(d, c.Quality)
				if err != nil {
					return f, err
				}
				f.RomanNumerals = append(f.RomanNumerals, romanFixture{k, d, seventh, c, rn})
			}
		}
	}
	return f, nil
}

// timeline is the analysed C major I-V-vi-IV at 90 BPM: the example used
// throughout the docs and the UI defaults.
func timeline() (music.Timeline, error) {
	prog := music.Progression{{Degree: 1, Beats: 4}, {Degree: 5, Beats: 4}, {Degree: 6, Beats: 4}, {Degree: 4, Beats: 4}}
	tl, err := music.FromProgression(music.Key{Tonic: music.C, Mode: music.Major}, prog, 90)
	if err != nil {
		return tl, err
	}
	return music.Analyze(tl)
}
