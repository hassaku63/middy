package music

import (
	"strings"
	"testing"
)

// TestFlatAliasFromMajor (#130): a chord whose root sits a semitone below
// the major scale's degree also gets the name a major-scale reading gives
// it, so a learner can match textbooks that write ♭VII or ♭VI.
func TestFlatAliasFromMajor(t *testing.T) {
	tests := []struct {
		key    Key
		labels []string
		at     int
		alias  string // "" means no alias
	}{
		{Key{G, Mixolydian}, []string{"I", "VII", "IV", "I"}, 1, "♭VII"},
		{Key{D, Dorian}, []string{"i", "IV", "i", "VII"}, 3, "♭VII"},
		{Key{D, Dorian}, []string{"i", "IV"}, 1, ""}, // the dorian IV is the major scale's IV
		{Key{A, Minor}, []string{"i", "VI", "III", "VII"}, 1, "♭VI"},
		{Key{A, Minor}, []string{"i", "VI", "III", "VII"}, 2, "♭III"},
		{Key{C, Major}, []string{"I", "IV", "V", "I"}, 1, ""},
	}
	for _, tt := range tests {
		e := analysed(t, tt.key, 4, tt.labels...).Harmony[tt.at].Analysis.Explanation
		has := strings.Contains(e, "長調の音階から見ると")
		if tt.alias == "" && has {
			t.Errorf("%s %s: no alias expected: %s", tt.key, tt.labels[tt.at], e)
		}
		if tt.alias != "" && !strings.Contains(e, "長調の音階から見ると "+tt.alias) {
			t.Errorf("%s %s: alias %s expected: %s", tt.key, tt.labels[tt.at], tt.alias, e)
		}
	}
}
