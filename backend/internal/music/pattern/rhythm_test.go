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
		{"arpeggio", "ballad", 4, 4, true, 2, false},
		{"strum", "bossa", 3, 2, true, 2, true},
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
			if !strings.Contains(d.Explanation, "キックとスネア") {
				t.Errorf("drum explanation should talk about kick and snare: %q", d.Explanation)
			}
		})
	}
}

func TestExplainPlacementWording(t *testing.T) {
	if got := explainPlacement("伴奏", placement{onBeat: 4}); !strings.Contains(got, "すべて拍の頭") {
		t.Errorf("all on beats: %q", got)
	}
	if got := explainPlacement("伴奏", placement{onBeat: 2, offBeat: 2}); !strings.Contains(got, "シンコペーション") {
		t.Errorf("half off beats should be syncopated: %q", got)
	}
	if got := explainPlacement("伴奏", placement{onBeat: 7, offBeat: 1}); !strings.Contains(got, "軽い推進力") {
		t.Errorf("mostly on beats: %q", got)
	}
	if got := explainPlacement("伴奏", placement{}); !strings.HasSuffix(got, "0 回。") {
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
