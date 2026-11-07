package fixtures

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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
	if len(f.PitchClasses) != 12 || len(f.Chords) != 12*8 || len(f.RomanNumerals) != 2*7*2 {
		t.Fatalf("unexpected sizes: %d %d %d", len(f.PitchClasses), len(f.Chords), len(f.RomanNumerals))
	}
}
