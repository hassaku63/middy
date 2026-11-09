// Package golden holds the golden-file tests of the pattern catalogue
// (BL-176, #207). They live outside internal/music, which must not touch
// the OS (depguard), and see the catalogue only through its exported API.
package golden

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

// The catalogue as it renders today (BL-176, #207): every accompaniment
// and drum pattern in each meter it fits, and every style's groove in each
// meter, on one fixed progression. A change to the catalogue's data or to
// how it renders shows here line by line; refresh on purpose with
//
//	go test ./internal/golden -update
var update = flag.Bool("update", false, "rewrite the golden files in testdata")

// goldenBase is I–IV–V7–I in C major, a bar per chord, at 120 BPM.
func goldenBase(t *testing.T, ts music.TimeSignature) music.Timeline {
	t.Helper()
	bar := ts.BeatsPerBar()
	prog := music.Progression{{Degree: 1, Beats: bar}, {Degree: 4, Beats: bar}, {Degree: 5, Seventh: true, Beats: bar}, {Degree: 1, Beats: bar}}
	tl, err := music.FromProgressionIn(ts, music.Key{Tonic: music.C, Mode: music.Major}, prog, 120)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

func meterName(ts music.TimeSignature) string {
	return fmt.Sprintf("%d/%d", ts.Numerator, ts.Denominator)
}

// dump writes a timeline's tracks one note per line.
func dump(b *strings.Builder, title string, tl music.Timeline) {
	fmt.Fprintf(b, "== %s\n", title)
	for _, tr := range tl.Tracks {
		fmt.Fprintf(b, "track %s ch%d program %d\n", tr.Name, tr.Channel, tr.Program)
		for _, n := range tr.Notes {
			fmt.Fprintf(b, "  %d %d %d %d\n", n.Start, n.Duration, n.Pitch, n.Velocity)
		}
	}
}

func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "pattern", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path) //nolint:gosec // the repository's own golden files
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if got != string(want) {
		gl, wl := strings.Split(got, "\n"), strings.Split(string(want), "\n")
		for i := range min(len(gl), len(wl)) {
			if gl[i] != wl[i] {
				t.Fatalf("%s differs from line %d: got %q, want %q (run with -update if intended)", path, i+1, gl[i], wl[i])
			}
		}
		t.Fatalf("%s differs in length: got %d lines, want %d (run with -update if intended)", path, len(gl), len(wl))
	}
}

func TestGoldenAccompaniments(t *testing.T) {
	var b strings.Builder
	for _, ts := range pattern.GrooveMeters {
		for _, acc := range pattern.AccompanimentsFor(ts) {
			tl, err := pattern.Render(goldenBase(t, ts), pattern.Choice{Accompaniment: acc.ID})
			if err != nil {
				t.Fatalf("%s in %s: %v", acc.ID, meterName(ts), err)
			}
			dump(&b, acc.ID+" "+meterName(ts), tl)
		}
	}
	checkGolden(t, "accompaniments.txt", b.String())
}

func TestGoldenDrums(t *testing.T) {
	var b strings.Builder
	for _, ts := range pattern.GrooveMeters {
		for _, dr := range pattern.DrumPatternsFor(ts) {
			tl, err := pattern.Render(goldenBase(t, ts), pattern.Choice{Drums: dr.ID})
			if err != nil {
				t.Fatalf("%s in %s: %v", dr.ID, meterName(ts), err)
			}
			dump(&b, dr.ID+" "+meterName(ts), tl)
		}
	}
	checkGolden(t, "drums.txt", b.String())
}

func TestGoldenStyleGrooves(t *testing.T) {
	var b strings.Builder
	for _, st := range pattern.Styles() {
		for _, ts := range pattern.GrooveMeters {
			g := st.Groove(ts)
			tl, err := pattern.Render(goldenBase(t, ts), pattern.Choice{Accompaniment: g.Accompaniment, Drums: g.Drums, Feel: g.Feel, Style: st.ID})
			if err != nil {
				t.Fatalf("%s in %s: %v", st.ID, meterName(ts), err)
			}
			dump(&b, fmt.Sprintf("%s %s (%s, %s, %s)", st.ID, meterName(ts), g.Accompaniment, g.Drums, g.Feel), tl)
		}
	}
	checkGolden(t, "grooves.txt", b.String())
}
