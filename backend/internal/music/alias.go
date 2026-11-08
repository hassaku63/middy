package music

import (
	"fmt"
	"strings"
)

// majorAlias names a chord the way a reading against the major scale
// would (#130): mixolydian's VII is ♭VII there, minor's VI is ♭VI. middy
// writes numerals against the key's own scale (ADR-0015), and textbooks
// and lead sheets often use the other reading, so the explanation gives
// both. Empty for chords whose root is on the major scale's degree, and in
// major keys.
func majorAlias(key Key, degree int, c Chord) string {
	if key.Mode == Major || degree < 1 || degree > 7 {
		return ""
	}
	major := Key{Tonic: key.Tonic, Mode: Major}
	if c.Root != key.Scale()[degree-1] || c.Root != major.Scale()[degree-1].Add(-1) {
		return ""
	}
	numeral, err := RomanNumeralAlt(degree, -1, c.Quality)
	if err != nil {
		return ""
	}
	return fmt.Sprintf(" 長調の音階から見ると %s（主音から数えた位置が長調より半音低い。教本や譜面ではこう書くことも多い）。", strings.Replace(numeral, "b", "♭", 1))
}
