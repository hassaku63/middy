package music

import (
	"fmt"
	"strings"
)

// Bass is the note a slash chord puts under its chord (ADR-0020): a scale
// degree of the key, raised or lowered by Accidental (+1 sharp, -1 flat).
// The chord keeps its identity; only the lowest note changes, as in an
// inversion (I/3), a pedal (IV/1) or a line cliché (i, i/#7, i/7, i/#6).
type Bass struct {
	Degree     int `json:"degree"`               // 1..7
	Accidental int `json:"accidental,omitempty"` // -1, 0 or +1
}

// suffix writes the bass as a label writes it: "/3", "/b7", "/#7".
func (b Bass) suffix() string {
	switch b.Accidental {
	case -1:
		return fmt.Sprintf("/b%d", b.Degree)
	case 1:
		return fmt.Sprintf("/#%d", b.Degree)
	}
	return fmt.Sprintf("/%d", b.Degree)
}

// pitchClass is the bass note in the key.
func (b Bass) pitchClass(key Key) PitchClass {
	return key.Scale()[b.Degree-1].Add(b.Accidental)
}

// splitBass separates a slash-bass suffix ("/3", "/b7", "/#7") from the
// chord label. A "/" followed by a roman numeral is a secondary dominant
// (V/V) and is left in the label; a "/" followed by anything else that is
// not a degree 1..7 is an error.
func splitBass(label string) (string, *Bass, error) {
	trimmed := strings.TrimSpace(label)
	i := strings.LastIndex(trimmed, "/")
	if i < 0 {
		return trimmed, nil, nil
	}
	head, tail := trimmed[:i], trimmed[i+1:]
	acc := 0
	switch {
	case strings.HasPrefix(tail, "b"):
		acc, tail = -1, tail[1:]
	case strings.HasPrefix(tail, "#"):
		acc, tail = 1, tail[1:]
	}
	if len(tail) == 1 && tail[0] >= '0' && tail[0] <= '9' {
		d := int(tail[0] - '0')
		if d < 1 || d > 7 {
			return "", nil, fmt.Errorf("music: %q: the bass after \"/\" must be a scale degree 1..7, got %d", trimmed, d)
		}
		return head, &Bass{Degree: d, Accidental: acc}, nil
	}
	if acc != 0 || tail == "" || (tail[0] >= '0' && tail[0] <= '9') {
		return "", nil, fmt.Errorf("music: %q: the bass after \"/\" must be a scale degree 1..7 (optionally b or #)", trimmed)
	}
	return trimmed, nil, nil // a secondary dominant such as V/V
}
