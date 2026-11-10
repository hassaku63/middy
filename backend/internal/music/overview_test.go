package music

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func analysed(t *testing.T, key Key, beats int, labels ...string) Timeline {
	t.Helper()
	var prog Progression
	for _, l := range labels {
		s, err := StepFromLabel(key, l, beats)
		if err != nil {
			t.Fatalf("%s %q: %v", key, l, err)
		}
		prog = append(prog, s)
	}
	tl, err := FromProgression(key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	tl, err = Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

func TestOverviewExplainsTheModeAndTheFlow(t *testing.T) {
	tests := []struct {
		name   string
		key    Key
		beats  int
		labels []string
		whole  []string
		flow   []string
	}{
		{"dorian names the chord with the characteristic tone and its minor twin",
			Key{D, Dorian}, 4, []string{"i", "IV", "i", "VII"},
			[]string{"D dorian（ドリアン）", "この例では IV（G）が特徴音 B を含む", "短調なら、ここは iv（Gm）になる"},
			[]string{"4 個の和音で 4 小節", "最後は VII（C）で、主和音に戻らずに終わる", "繰り返すと、最後の VII から頭の i へ、導音を使わずに戻る"}},
		{"dorian without the characteristic tone sounds like minor",
			Key{D, Dorian}, 4, []string{"i", "v", "i"},
			[]string{"特徴音（B）を含まないので、響きは短調とほとんど変わらない"},
			// v has no leading tone, so v - i is not called an authentic cadence (BL-103).
			[]string{"最後は主和音に戻って落ち着く"}},
		{"mixolydian names VII and the major key's vii°",
			Key{G, Mixolydian}, 4, []string{"I", "VII", "IV", "I"},
			[]string{"この例では VII（F）が特徴音 F を含む", "長調なら、ここは vii°（F#dim）になる"},
			[]string{"最後は IV から I への変格終止で、柔らかく閉じる"}},
		{"several characteristic chords are listed in order",
			Key{D, Dorian}, 4, []string{"i", "ii", "IV", "i"},
			[]string{"ii（Em）と IV（G）が特徴音 B を含む", "それぞれ ii°（Edim）と iv（Gm）になる"}, nil},
		{"major names borrowed chords and where they come from",
			Key{G, Major}, 4, []string{"I", "IV", "iv", "I"},
			[]string{"G major（長調）", "この例では iv（Cm）が同主短調（G minor）から借りた和音"},
			[]string{"最後は iv から I への変格終止"}},
		{"minor names the harmonic-minor dominant",
			Key{E, Minor}, 2, []string{"i", "iv", "VII", "III", "VI", "iiø7", "V7", "i"},
			[]string{"E minor（短調）", "V7（B7）は和声的短音階の形"},
			[]string{"8 個の和音で 4 小節", "最後は V7 から i への正格終止で閉じる"}},
		{"a plain major loop says how it turns around",
			Key{C, Major}, 4, []string{"I", "V", "vi", "IV"},
			nil, []string{"主和音から始まり", "繰り返すと、最後の IV から頭の I へ柔らかく戻り、ひと回りが閉じる"}},
		// The seam names no second cadence on the V (#378, ADR-0052).
		{"a loop that stops on V and turns around to I",
			Key{C, Major}, 4, []string{"I", "vi", "IV", "V"},
			nil, []string{"最後は IV から V への半終止で止まり", "繰り返すと、最後の V は頭の I へ解決し、ひと回りが閉じる"}},
		{"a loop that starts elsewhere and stops on V",
			Key{C, Major}, 4, []string{"vi", "IV", "I", "V"},
			nil, []string{"vi（Am）から始まり", "最後は I から V への半終止で止まり"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := analysed(t, tt.key, tt.beats, tt.labels...).Overview
			if o == nil {
				t.Fatal("no overview")
			}
			for _, w := range tt.whole {
				if !strings.Contains(o.Whole, w) {
					t.Errorf("whole lacks %q:\n%s", w, o.Whole)
				}
			}
			for _, w := range tt.flow {
				if !strings.Contains(o.Flow, w) {
					t.Errorf("flow lacks %q:\n%s", w, o.Flow)
				}
			}
		})
	}
}

func TestOverviewOfASingleChordAndOfNothing(t *testing.T) {
	o := analysed(t, Key{C, Major}, 4, "I").Overview
	if o.Flow != "1 個の和音で 1 小節。" {
		t.Errorf("flow = %q", o.Flow)
	}
	empty, err := Analyze(NewTimeline(Key{C, Major}, 100))
	if err != nil {
		t.Fatal(err)
	}
	if empty.Overview != nil {
		t.Errorf("an empty timeline has no overview: %+v", empty.Overview)
	}
}

func TestOverviewNamesTheMeterAndCountsItsBars(t *testing.T) {
	key := Key{C, Major}
	var prog Progression
	for _, l := range []string{"I", "IV", "V7", "I"} {
		s, err := StepFromLabel(key, l, 3)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	tl, err := FromProgressionIn(ThreeFour, key, prog, 150)
	if err != nil {
		t.Fatal(err)
	}
	tl, err = Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"3 拍子（3/4）で、1 拍目が強く 2・3 拍目が弱い", "4 個の和音で 4 小節"} {
		if !strings.Contains(tl.Overview.Flow, want) {
			t.Errorf("flow lacks %q: %s", want, tl.Overview.Flow)
		}
	}
	if _, err := FromProgressionIn(TimeSignature{9, 8}, key, prog, 150); err == nil {
		t.Error("9/8 is not supported yet")
	}
}

// TestSixEightCountsTwoDottedQuarterBeats checks the rule of ADR-0017 in a
// compound meter: a chord of 2 beats lasts one bar of 6/8 (two dotted
// quarters, 1440 ticks), and 1 beat lasts half a bar.
func TestSixEightCountsTwoDottedQuarterBeats(t *testing.T) {
	key := Key{C, Major}
	prog := Progression{{Degree: 1, Beats: 2}, {Degree: 4, Beats: 1}, {Degree: 5, Beats: 1}, {Degree: 1, Beats: 2}}
	tl, err := FromProgressionIn(SixEight, key, prog, 60)
	if err != nil {
		t.Fatal(err)
	}
	var starts, durations []Tick
	for _, h := range tl.Harmony {
		starts = append(starts, h.Start)
		durations = append(durations, h.Duration)
	}
	if fmt.Sprint(starts) != "[0 1440 2160 2880]" || fmt.Sprint(durations) != "[1440 720 720 1440]" {
		t.Errorf("starts %v, durations %v", starts, durations)
	}
	tl, err = Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"付点 4 分の 2 拍", "4 個の和音で 3 小節"} {
		if !strings.Contains(tl.Overview.Flow, want) {
			t.Errorf("flow lacks %q: %s", want, tl.Overview.Flow)
		}
	}
}

// TestTwelveEightCountsFourDottedQuarterBeats (#308): 12/8 is counted in
// four dotted quarters, so a chord of 4 beats lasts a bar (2880 ticks) and
// 2 beats half a bar; the overview says how the meter is counted.
func TestTwelveEightCountsFourDottedQuarterBeats(t *testing.T) {
	if err := SupportedMeter(TwelveEight); err != nil {
		t.Fatal(err)
	}
	key := Key{C, Major}
	prog := Progression{{Degree: 1, Beats: 4}, {Degree: 4, Beats: 2}, {Degree: 5, Beats: 2}, {Degree: 1, Beats: 4}}
	tl, err := FromProgressionIn(TwelveEight, key, prog, 50)
	if err != nil {
		t.Fatal(err)
	}
	var starts, durations []Tick
	for _, h := range tl.Harmony {
		starts = append(starts, h.Start)
		durations = append(durations, h.Duration)
	}
	if fmt.Sprint(starts) != "[0 2880 4320 5760]" || fmt.Sprint(durations) != "[2880 1440 1440 2880]" {
		t.Errorf("starts %v, durations %v", starts, durations)
	}
	tl, err = Analyze(tl)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"付点 4 分の 4 拍", "4 個の和音で 3 小節"} {
		if !strings.Contains(tl.Overview.Flow, want) {
			t.Errorf("flow lacks %q: %s", want, tl.Overview.Flow)
		}
	}
	if _, err := FromProgressionIn(TimeSignature{Numerator: 9, Denominator: 8}, key, prog, 50); err == nil {
		t.Error("9/8 is not rendered yet")
	}
}

func TestBeatsOfEachMeter(t *testing.T) {
	tests := []struct {
		ts          TimeSignature
		compound    bool
		beats       int
		beat, whole Tick
	}{
		{FourFour, false, 4, 480, 1920},
		{ThreeFour, false, 3, 480, 1440},
		{SixEight, true, 2, 720, 1440},
		{TimeSignature{12, 8}, true, 4, 720, 2880},
		{TimeSignature{3, 8}, false, 3, 240, 720},
		{TimeSignature{2, 2}, false, 2, 960, 1920},
	}
	for _, tt := range tests {
		if tt.ts.Compound() != tt.compound || tt.ts.BeatsPerBar() != tt.beats ||
			BeatTicks(tt.ts, DefaultPPQ) != tt.beat || BarTicks(tt.ts, DefaultPPQ) != tt.whole {
			t.Errorf("%d/%d: compound %v, %d beats of %d in a bar of %d", tt.ts.Numerator, tt.ts.Denominator,
				tt.ts.Compound(), tt.ts.BeatsPerBar(), BeatTicks(tt.ts, DefaultPPQ), BarTicks(tt.ts, DefaultPPQ))
		}
	}
}

// TestTempoCountsTheBeat: 60 BPM in 6/8 is 60 dotted quarters a minute,
// so 90 quarters; in 4/4 and 3/4 the beat is the quarter (ADR-0019).
func TestTempoCountsTheBeat(t *testing.T) {
	for _, tt := range []struct {
		ts   TimeSignature
		want float64
	}{{FourFour, 60}, {ThreeFour, 60}, {SixEight, 90}, {TwelveEight, 90}} {
		tl := NewTimeline(Key{C, Major}, 60)
		tl.TimeSignature = tt.ts
		if got := tl.QuarterBPM(); got != tt.want {
			t.Errorf("%d/%d at 60 BPM = %v quarters a minute, want %v", tt.ts.Numerator, tt.ts.Denominator, got, tt.want)
		}
	}
}

// TestDegreeMapListsTheModesSevenChords checks the degree map (BL-099):
// the seven triads of the mode with their function, and which the example
// plays as part of the key (a borrowed iv or an applied V/vi does not mark
// iv or V; a seventh chord marks its degree).
func TestDegreeMapListsTheModesSevenChords(t *testing.T) {
	tests := []struct {
		key      Key
		labels   []string
		numerals string
		symbols  string
		used     string
	}{
		{Key{C, Major}, []string{"I", "V7/vi", "vi", "iv", "V7", "I"},
			"I ii iii IV V vi vii°", "C Dm Em F G Am Bdim", "1 5 6"},
		{Key{D, Dorian}, []string{"i", "IV", "i", "VII"},
			"i ii III IV v vi° VII", "Dm Em F G Am Bdim C", "1 4 7"},
		{Key{A, Minor}, []string{"i", "VI", "III", "VII", "iv", "V7", "i"},
			"i ii° III iv v VI VII", "Am Bdim C Dm Em F G", "1 3 4 5 6 7"},
	}
	for _, tt := range tests {
		tl := analysed(t, tt.key, 4, tt.labels...)
		var numerals, symbols, used []string
		for _, d := range tl.Overview.Degrees {
			numerals = append(numerals, d.RomanNumeral)
			symbols = append(symbols, d.Symbol)
			if d.Used {
				used = append(used, strconv.Itoa(d.Degree))
			}
		}
		if got := strings.Join(numerals, " "); got != tt.numerals {
			t.Errorf("%s numerals = %s, want %s", tt.key, got, tt.numerals)
		}
		if got := strings.Join(symbols, " "); got != tt.symbols {
			t.Errorf("%s symbols = %s, want %s", tt.key, got, tt.symbols)
		}
		if got := strings.Join(used, " "); got != tt.used {
			t.Errorf("%s used = %s, want %s", tt.key, got, tt.used)
		}
		if f := tl.Overview.Degrees[4].Function; f != Dominant {
			t.Errorf("%s: degree 5 should be dominant, got %s", tt.key, f)
		}
	}
}

// TestDegreeMapNamesAlteredChords (#296): a chord played on a degree whose
// tones leave the scale (minor's harmonic V7, a blues I7) is named on the
// map; a seventh or a suspension inside the scale is not.
func TestDegreeMapNamesAlteredChords(t *testing.T) {
	tests := []struct {
		key     Key
		labels  []string
		altered map[int]string
	}{
		{Key{E, Minor}, []string{"i", "iv", "V7", "i"}, map[int]string{5: "B7"}},
		{Key{E, Minor}, []string{"i", "v", "V", "V7", "i"}, map[int]string{5: "B B7"}},
		{Key{C, Major}, []string{"I7", "IV7", "I7", "V7"}, map[int]string{1: "C7", 4: "F7"}},
		{Key{C, Major}, []string{"Imaj7", "ii7", "Vsus4", "V7", "I"}, map[int]string{}},
	}
	for _, tt := range tests {
		tl := analysed(t, tt.key, 4, tt.labels...)
		got := map[int]string{}
		for _, d := range tl.Overview.Degrees {
			if len(d.Altered) > 0 {
				got[d.Degree] = strings.Join(d.Altered, " ")
			}
		}
		if fmt.Sprint(got) != fmt.Sprint(tt.altered) {
			t.Errorf("%s %v: altered = %v, want %v", tt.key, tt.labels, got, tt.altered)
		}
	}
}

// TestLoopSeamNamesNoSecondCadence (#378 B1, ADR-0052): the V that ends a
// loop is a half cadence; the seam back to the head's I is not called a
// second, authentic one.
func TestLoopSeamNamesNoSecondCadence(t *testing.T) {
	for _, labels := range [][]string{{"I", "vi", "IV", "V"}, {"I", "V", "vi", "IV"}} {
		flow := analysed(t, Key{C, Major}, 4, labels...).Overview.Flow
		if strings.Contains(flow, "頭の I への") || strings.Count(flow, "終止") > 1 {
			t.Errorf("%v:\n%s", labels, flow)
		}
	}
}
