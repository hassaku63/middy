package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestGridNamesItsNearestPattern (BL-124): a grid is compared with the
// catalogue drums of its meter, and the explanation names the nearest and
// the strokes it adds and drops.
func TestGridNamesItsNearestPattern(t *testing.T) {
	explain := func(ts music.TimeSignature, g music.DrumGrid) string {
		t.Helper()
		d, err := GridDrums(g, ts, 100)
		if err != nil {
			t.Fatal(err)
		}
		return describeDrums(d, ts).Explanation
	}
	// The eight-beat with the kick on 3 moved to the "and" of 3.
	moved := explain(music.FourFour, grid(2, map[string]string{
		"kick": "x... .x..", "snare": "..x. ..x.", "hat": "xxxx xxxx",
	}))
	for _, want := range []string{"いちばん近いカタログの型は「8 ビート」", "キックを 3 拍目の裏に足し", "キックを 3 拍目から抜いた"} {
		if !strings.Contains(moved, want) {
			t.Errorf("want %q in %s", want, moved)
		}
	}
	same := explain(music.FourFour, grid(2, map[string]string{
		"kick": "x... x...", "snare": "..x. ..x.", "hat": "xxxx xxxx",
	}))
	if !strings.Contains(same, "「8 ビート」と同じ打ち方") {
		t.Errorf("an unchanged eight-beat says so: %s", same)
	}
	// In 6/8 a jig with a crash on the first beat.
	jig := explain(music.SixEight, grid(3, map[string]string{
		"kick": "x..x..", "snare": "..x..x", "hat": "xxxxxx", "crash": "x.....",
	}))
	if !strings.Contains(jig, "「ジグ」") || !strings.Contains(jig, "クラッシュを 1 拍目に足した") {
		t.Errorf("jig: %s", jig)
	}
	// Sixteenths are named within the beat.
	sixteenth := explain(music.FourFour, grid(4, map[string]string{
		"kick": "x... ...x x... ....", "snare": ".... x... .... x...", "hat": "x.x. x.x. x.x. x.x.",
	}))
	if !strings.Contains(sixteenth, "キックを 2 拍目の 4 つめの 16 分に足した") {
		t.Errorf("sixteenth: %s", sixteenth)
	}
	// Far from everything: say how far rather than list every stroke.
	far := explain(music.FourFour, grid(4, map[string]string{
		"low-tom": "x.x. x.x. x.x. x.x.", "high-tom": ".x.x .x.x .x.x .x.x",
	}))
	if !strings.Contains(far, "打ち方が違う") {
		t.Errorf("far: %s", far)
	}
}
