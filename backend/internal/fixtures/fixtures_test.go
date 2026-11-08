package fixtures

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Where the frontend reads the fixtures from, relative to this package.
const frontendDir = "../../../frontend/src/shared/api/fixtures"

// TestCommittedFixturesAreCurrent fails when the files under the frontend
// differ from what the engine produces now. Fix with `make fixtures`.
func TestCommittedFixturesAreCurrent(t *testing.T) {
	files, err := Files()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no fixtures generated")
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(frontendDir, name))
		if err != nil {
			t.Fatalf("%s: %v (run `make fixtures`)", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is stale: run `make fixtures` and commit the result", name)
		}
	}
}

func TestNotationCoversEveryQualityAndKey(t *testing.T) {
	f, err := notation()
	if err != nil {
		t.Fatal(err)
	}
	if len(f.PitchClasses) != 12 || len(f.RomanNumerals) != 2*7*2 {
		t.Fatalf("unexpected sizes: %d %d", len(f.PitchClasses), len(f.RomanNumerals))
	}
	if want := 12 * len(music.Qualities()); len(f.Chords) != want || want < 12*9 {
		t.Fatalf("chords = %d, want one per root and quality (%d)", len(f.Chords), want)
	}
	if want := 12 * len(music.Modes()); len(f.KeyNames) != want || want < 12*4 {
		t.Fatalf("key names = %d, want one per tonic and mode (%d)", len(f.KeyNames), want)
	}
}

// TestNoOrphanFixtures fails when a committed fixture is no longer
// generated, so the frontend cannot keep asserting against a stale file.
func TestNoOrphanFixtures(t *testing.T) {
	files, err := Files()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(frontendDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if _, ok := files[e.Name()]; !ok {
			t.Errorf("%s is committed but no longer generated: delete it", e.Name())
		}
	}
}
