package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestSyncopationNeedsASilentNextBeat (BL-083): a hit off the beat is a
// syncopation only when the next beat does not sound (Longuet-Higgins &
// Lee). The pattern repeats, so the beat after the last one is the next
// bar's first. The explanation names the beat that is skipped.
func TestSyncopationNeedsASilentNextBeat(t *testing.T) {
	q, e := quarter, eighth
	tests := []struct {
		name   string
		starts []music.Tick
		syn    bool
		beats  string // the skipped beats, as the explanation names them
	}{
		{"even eighths", []music.Tick{0, e, q, q + e, 2 * q, 2*q + e, 3 * q, 3*q + e}, false, ""},
		{"and-of-4 into a silent 1", []music.Tick{q, 2 * q, 3*q + e}, true, "1 拍目"},
		{"and-of-2 into a silent 3", []music.Tick{0, q + e, 3 * q}, true, "3 拍目"},
		{"offbeats answered by the beat", []music.Tick{0, e, q, 2 * q, 2*q + e, 3 * q}, false, ""},
		{"and-of-1 leaves 2 silent", []music.Tick{0, e, 2 * q, 2*q + e}, true, "2 拍目と 4 拍目"},
		{"only on the beats", []music.Tick{0, q, 2 * q, 3 * q}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := classify(tt.starts, music.FourFour)
			if p.syncopated() != tt.syn {
				t.Fatalf("syncopated = %v, want %v (skipped %v)", p.syncopated(), tt.syn, p.skipped)
			}
			e := explainPlacement("伴奏", p, 4)
			if tt.syn && !strings.Contains(e, tt.beats+"の頭が鳴らない") {
				t.Errorf("explanation should name %s: %s", tt.beats, e)
			}
			if !tt.syn && strings.Contains(e, "シンコペーション") {
				t.Errorf("no syncopation to speak of: %s", e)
			}
		})
	}
}

// TestCatalogueSyncopation pins the catalogue under the new definition:
// the reel's offbeat chops leave beats 2 and 4 silent (an off-beat
// backing), while the arpeggio's and the strum's offbeats are answered by
// the next beat, and comping's and-of-2 by the bass on 3.
func TestCatalogueSyncopation(t *testing.T) {
	for id, want := range map[string]bool{"reel": true, "comp": false, "arpeggio": false, "pulse": false, "strum": false} {
		a, err := FindAccompaniment(id)
		if err != nil {
			t.Fatal(err)
		}
		if got := describeAccompaniment(a, music.FourFour).Syncopated; got != want {
			t.Errorf("%s syncopated = %v, want %v", id, got, want)
		}
	}
}
