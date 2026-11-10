package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func TestRhythmDescriptions(t *testing.T) {
	tests := []struct {
		acc        string
		drums      string
		wantAccOn  int
		wantAccOff int
		wantAccSyn bool
		wantDrumOn int
		wantDrumSy bool
	}{
		{"pulse", "eight-beat", 4, 0, false, 4, false},
		{"arpeggio", "ballad", 4, 4, false, 2, false}, // even eighths: every offbeat is answered (BL-083)
		{"strum", "bossa", 3, 2, false, 2, false},     // the next beat sounds after each offbeat (BL-083)
		{"block", "", 1, 0, false, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.acc+"+"+tt.drums, func(t *testing.T) {
			tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1, 4)
			var dr *Drums
			if tt.drums != "" {
				d, err := FindDrums(tt.drums)
				if err != nil {
					t.Fatal(err)
				}
				dr = &d
			}
			got, err := Apply(tl, mustAcc(t, tt.acc), dr)
			if err != nil {
				t.Fatal(err)
			}
			if got.Rhythm == nil || got.Rhythm.Accompaniment == nil {
				t.Fatal("rhythm explanation missing")
			}
			a := got.Rhythm.Accompaniment
			if a.ID != tt.acc || a.OnBeat != tt.wantAccOn || a.OffBeat != tt.wantAccOff || a.Syncopated != tt.wantAccSyn {
				t.Errorf("accompaniment = %+v", a)
			}
			if a.Explanation == "" || a.Description == "" {
				t.Errorf("accompaniment needs text: %+v", a)
			}
			if tt.drums == "" {
				if got.Rhythm.Drums != nil {
					t.Errorf("no drums requested but described: %+v", got.Rhythm.Drums)
				}
				return
			}
			d := got.Rhythm.Drums
			if d == nil || d.ID != tt.drums || d.OnBeat != tt.wantDrumOn || d.Syncopated != tt.wantDrumSy {
				t.Errorf("drums = %+v", d)
			}
			// Every kit here has a kick; bossa has no snare, so only the kick is named.
			if !strings.Contains(d.Explanation, "1 小節にキック") {
				t.Errorf("drum explanation should name the kick: %q", d.Explanation)
			}
		})
	}
}

func TestExplainPlacementWording(t *testing.T) {
	if got := explainPlacement("伴奏", placement{onBeat: 4}, 4); !strings.Contains(got, "すべて拍の頭（1・2・3・4）") {
		t.Errorf("all on beats: %q", got)
	}
	// Syncopation is a skipped beat after an offbeat hit (BL-083), not a
	// share of offbeats: half off the beat is not enough on its own.
	if got := explainPlacement("伴奏", placement{onBeat: 2, offBeat: 2}, 4); strings.Contains(got, "シンコペーション") {
		t.Errorf("offbeats answered by the beats are not syncopated: %q", got)
	}
	if got := explainPlacement("伴奏", placement{onBeat: 2, offBeat: 2, skipped: []int{3}}, 4); !strings.Contains(got, "3 拍目の頭が鳴らない") {
		t.Errorf("a skipped beat 3 is a syncopation over beat 3: %q", got)
	}
	if got := explainPlacement("伴奏", placement{onBeat: 7, offBeat: 1}, 4); !strings.Contains(got, "軽い推進力") {
		t.Errorf("mostly on beats: %q", got)
	}
	if got := explainPlacement("伴奏", placement{}, 4); !strings.HasSuffix(got, "0 回。") {
		t.Errorf("empty: %q", got)
	}
}

func TestRhythmSurvivesJSON(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1)
	d, _ := FindDrums("eight-beat")
	got, err := Apply(tl, mustAcc(t, "pulse"), &d)
	if err != nil {
		t.Fatal(err)
	}
	if got.Rhythm.Drums.Hits != 4 { // 2 kicks + 2 snares; hats are not counted
		t.Fatalf("drum hits = %d", got.Rhythm.Drums.Hits)
	}
}

// TestSixteenBeat (#308): the sixteen-beat sits in pop and rock after their
// eight-beat, and its explanation counts the kick on the last sixteenth of
// beat 2 as finer than the offbeat, not as an offbeat.
func TestSixteenBeat(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1, 4, 5, 1)
	out, err := Render(tl, Choice{Accompaniment: "strum", Drums: "sixteen-beat"})
	if err != nil {
		t.Fatal(err)
	}
	ex := out.Rhythm.Drums.Explanation
	for _, want := range []string{"拍の頭が 4 回", "細かい位置（16 分など）が 1 回", "2・4 拍目のスネアがバックビート"} {
		if !strings.Contains(ex, want) {
			t.Errorf("explanation lacks %q: %s", want, ex)
		}
	}
	for _, id := range []string{"pop", "rock"} {
		st, err := FindStyle(id)
		if err != nil {
			t.Fatal(err)
		}
		if st.Drums[len(st.Drums)-1] != "sixteen-beat" || st.Groove(music.FourFour).Drums == "sixteen-beat" {
			t.Errorf("%s: drums %v, groove %+v", id, st.Drums, st.Groove(music.FourFour))
		}
	}
}

// TestSixteenthStrum (#308): the sixteenth-note cutting joins pop and rock
// after their other accompaniments, and its notes count the hits finer
// than the offbeat and the beats it skips over.
func TestSixteenthStrum(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1, 4)
	out, err := Render(tl, Choice{Accompaniment: "strum-16"})
	if err != nil {
		t.Fatal(err)
	}
	ex := out.Rhythm.Accompaniment.Explanation
	for _, want := range []string{"1 小節に伴奏が 7 回", "裏拍（8 分の裏）が 3 回、それより細かい位置（16 分など）が 2 回", "シンコペーション"} {
		if !strings.Contains(ex, want) {
			t.Errorf("explanation lacks %q: %s", want, ex)
		}
	}
	for _, id := range []string{"pop", "rock"} {
		st, err := FindStyle(id)
		if err != nil {
			t.Fatal(err)
		}
		if st.Accompaniments[len(st.Accompaniments)-1] != "strum-16" || st.Groove(music.FourFour).Accompaniment == "strum-16" {
			t.Errorf("%s: accompaniments %v, groove %+v", id, st.Accompaniments, st.Groove(music.FourFour))
		}
	}
}

// TestOffBeatMeansTheEighthOffBeat (#390 T3): 裏拍 is the eighth off-beat,
// so a figure with no hit there does not credit 裏拍 with its push.
func TestOffBeatMeansTheEighthOffBeat(t *testing.T) {
	tl := blockTimeline(t, music.FourFour)
	explain := func(c Choice) string {
		out, err := Render(tl, c)
		if err != nil {
			t.Fatal(err)
		}
		return out.Rhythm.Drums.Explanation
	}
	if e := explain(Choice{Drums: "sixteen-beat"}); strings.Contains(e, "裏拍が軽い推進力") || !strings.Contains(e, "拍の頭以外の細かい音が軽い推進力") {
		t.Errorf("sixteen-beat:\n%s", e)
	}
	out, err := Render(tl, Choice{Accompaniment: "reel"})
	if err != nil {
		t.Fatal(err)
	}
	if e := out.Rhythm.Accompaniment.Explanation; !strings.Contains(e, "裏拍（8 分の裏）が") {
		t.Errorf("the reel glosses 裏拍:\n%s", e)
	}
}
