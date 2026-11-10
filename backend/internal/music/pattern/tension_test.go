package pattern

import (
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// TestStyleTensions (BL-125, ADR-0012 decision 4): a jazz example adds
// the 9th to its seventh chords and the 13th to its dominants, only when
// the key's scale has them and never a semitone above a chord tone; the
// numerals stay at the seventh, and the rhythm notes say what was added.
func TestStyleTensions(t *testing.T) {
	render := func(key music.Key, style string, steps ...music.Step) music.Timeline {
		t.Helper()
		tl, err := music.FromProgression(key, music.Progression(steps), 120)
		if err != nil {
			t.Fatal(err)
		}
		out, err := Render(tl, Choice{Accompaniment: "block", Style: style})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	classes := func(tl music.Timeline, at music.Tick) []music.PitchClass {
		var pcs []music.PitchClass
		for _, n := range tl.Tracks[0].Notes {
			if n.Start == at && !slices.Contains(pcs, n.Pitch.Class()) {
				pcs = append(pcs, n.Pitch.Class())
			}
		}
		slices.Sort(pcs)
		return pcs
	}
	c := music.Key{Tonic: music.C, Mode: music.Major}
	bar := music.Tick(4 * music.DefaultPPQ)
	jazz := render(c, "jazz",
		music.Step{Degree: 2, Seventh: true, Beats: 4},
		music.Step{Degree: 5, Seventh: true, Beats: 4},
		music.Step{Degree: 1, Seventh: true, Beats: 4},
	)
	want := [][]music.PitchClass{
		{music.C, music.D, music.E, music.F, music.A},          // Dm7 + 9 (E)
		{music.D, music.E, music.F, music.G, music.A, music.B}, // G7 + 9 (A) + 13 (E)
		{music.C, music.D, music.E, music.G, music.B},          // Cmaj7 + 9 (D)
	}
	for i, w := range want {
		if got := classes(jazz, music.Tick(i)*bar); !slices.Equal(got, w) {
			t.Errorf("chord %d: %v, want %v", i, got, w)
		}
	}
	if jazz.Harmony[1].RomanNumeral != "V7" {
		t.Errorf("the numeral stays at the seventh: %s", jazz.Harmony[1].RomanNumeral)
	}
	if s := jazz.Rhythm.Style; s == nil || !strings.Contains(s.Tensions, "9 度") {
		t.Errorf("the style says what it added: %+v", s)
	}

	// In A minor the harmonic-minor V7 (E7) has no diatonic 9th or 13th.
	minor := render(music.Key{Tonic: music.A, Mode: music.Minor}, "jazz",
		music.Step{Degree: 5, Seventh: true, Beats: 4, Quality: ptrQuality(music.Dominant7)},
	)
	if got := classes(minor, 0); len(got) != 4 {
		t.Errorf("E7 in A minor keeps its four tones: %v", got)
	}
	// Rock adds nothing.
	rock := render(c, "rock", music.Step{Degree: 1, Seventh: true, Beats: 4})
	if got := classes(rock, 0); len(got) != 4 || rock.Rhythm.Style.Tensions != "" {
		t.Errorf("rock keeps the plain chord: %v %q", got, rock.Rhythm.Style.Tensions)
	}
}

func ptrQuality(q music.ChordQuality) *music.ChordQuality { return &q }

// TestTensionsReview (review of BL-125): jazz on plain block chords still
// colours them; a tension a semitone above a slash bass is not added; and
// the catalogue cannot be changed through Styles().
func TestTensionsReview(t *testing.T) {
	c := music.Key{Tonic: music.C, Mode: music.Major}
	tl, err := music.FromProgression(c, music.Progression{{Degree: 2, Seventh: true, Beats: 4}}, 120)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Render(tl, Choice{Style: "jazz"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plain.Tracks[0].Notes) != 5 || plain.Rhythm == nil || plain.Rhythm.Style.Tensions == "" {
		t.Errorf("Dm7 in jazz takes its 9th even without a pattern: %d notes, %+v", len(plain.Tracks[0].Notes), plain.Rhythm)
	}

	// Cmaj7 over a C# bass: its 9th, D, would sit a semitone above the
	// bass, so it is not added.
	bass := music.PitchClass(1)
	h := music.HarmonyEvent{Chord: music.Chord{Root: music.C, Quality: music.Major7}, Bass: &bass}
	jazz, _ := FindStyle("jazz")
	if got := tensionsOf(jazz.Tensions, music.Key{Tonic: music.C, Mode: music.Major}, h); len(got) != 0 {
		t.Errorf("no 9th a semitone above the slash bass: %v", got)
	}

	for _, st := range Styles() {
		if st.Tensions != nil {
			st.Tensions.Ninths = false
		}
	}
	if again, _ := FindStyle("jazz"); !again.Tensions.Ninths {
		t.Error("Styles() must not hand out the catalogue's own rule")
	}
}

// TestTensionsSecondReview: the rhythm note names only the tensions that
// sounded (ii7–Imaj7 has no dominant, so no 13th), swing on block chords
// in 6/8 is still refused, and a style from FindStyle cannot change the
// catalogue's rule.
func TestTensionsSecondReview(t *testing.T) {
	c := music.Key{Tonic: music.C, Mode: music.Major}
	tl, err := music.FromProgression(c, music.Progression{{Degree: 2, Seventh: true, Beats: 4}, {Degree: 1, Seventh: true, Beats: 4}}, 120)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(tl, Choice{Accompaniment: "block", Style: "jazz"})
	if err != nil {
		t.Fatal(err)
	}
	if text := out.Rhythm.Style.Tensions; !strings.Contains(text, "9 度") || strings.Contains(text, "13 度") {
		t.Errorf("ii7–Imaj7 adds 9ths only: %q", text)
	}

	six, err := music.FromProgressionIn(music.TimeSignature{Numerator: 6, Denominator: 8}, c, music.Progression{{Degree: 1, Beats: 2}}, 60)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Render(six, Choice{Accompaniment: "block", Feel: FeelSwing}); err == nil {
		t.Error("swing on 6/8 block chords was accepted")
	}

	s, err := FindStyle("jazz")
	if err != nil {
		t.Fatal(err)
	}
	s.Tensions.Ninths = false
	again, _ := FindStyle("jazz")
	if !again.Tensions.Ninths {
		t.Error("changing a found style changed the catalogue")
	}
}

// TestArpeggioPlaysTheTension (#307): an arpeggio strikes no whole chord,
// so it plays the style's first tension where its octave was, and the
// rhythm notes say so; a style without tensions keeps the octave.
func TestArpeggioPlaysTheTension(t *testing.T) {
	key := music.Key{Tonic: music.F, Mode: music.Major}
	var prog music.Progression
	for _, l := range []string{"Imaj7", "ii7"} {
		s, err := music.StepFromLabel(key, l, 4)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	tl, err := music.FromProgression(key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	classes := func(style string) map[music.PitchClass]bool {
		out, err := Render(tl, Choice{Accompaniment: "arpeggio", Style: style})
		if err != nil {
			t.Fatal(err)
		}
		got := map[music.PitchClass]bool{}
		for _, n := range out.Tracks[0].Notes {
			if n.Start < 1920 { // the Imaj7 bar
				got[n.Pitch.Class()] = true
			}
		}
		if style == "bossa" && (out.Rhythm.Style == nil || !strings.Contains(out.Rhythm.Style.Tensions, "9 度")) {
			t.Errorf("bossa's notes do not name the ninth: %+v", out.Rhythm.Style)
		}
		return got
	}
	if g := classes("bossa"); !g[music.G] {
		t.Errorf("bossa arpeggio on Fmaj7 lacks the ninth G: %v", g)
	}
	if g := classes("pop"); g[music.G] {
		t.Errorf("pop has no tensions, yet G sounds on Fmaj7: %v", g)
	}
}

// TestRequestedTensions (ADR-0037, BL-201): a candidate's colour replaces
// its style's rule: pop takes a ninth on Imaj7 and a thirteenth on V7 with
// "9-13", jazz takes none with "none"; the numerals stay at the seventh,
// and the notes say what was added even without a style.
func TestRequestedTensions(t *testing.T) {
	key := music.Key{Tonic: music.C, Mode: music.Major}
	var prog music.Progression
	for _, l := range []string{"Imaj7", "V7"} {
		s, err := music.StepFromLabel(key, l, 4)
		if err != nil {
			t.Fatal(err)
		}
		prog = append(prog, s)
	}
	tl, err := music.FromProgression(key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	sounds := func(out music.Timeline, from, to music.Tick, pc music.PitchClass) bool {
		for _, n := range out.Tracks[0].Notes {
			if n.Start >= from && n.Start < to && n.Pitch.Class() == pc {
				return true
			}
		}
		return false
	}
	out, err := Render(tl, Choice{Accompaniment: "pulse", Style: "pop", Tensions: TensionsNinths13})
	if err != nil {
		t.Fatal(err)
	}
	if !sounds(out, 0, 1920, music.D) || !sounds(out, 1920, 3840, music.E) {
		t.Errorf("9-13 on pop: want D over Cmaj7 and E (13th) over G7")
	}
	if rt := out.Rhythm.Tensions; rt == nil || rt.ID != "9-13" || !strings.Contains(rt.Explanation, "13 度") {
		t.Errorf("rhythm tensions = %+v", rt)
	}
	if out.Rhythm.Style == nil || out.Rhythm.Style.Tensions != "" {
		t.Errorf("the style does not claim the requested colour: %+v", out.Rhythm.Style)
	}
	if out.Harmony[0].RomanNumeral != "Imaj7" {
		t.Errorf("numeral = %s, want Imaj7", out.Harmony[0].RomanNumeral)
	}

	jazz, err := Render(tl, Choice{Accompaniment: "comp", Style: "jazz", Tensions: TensionsNone})
	if err != nil {
		t.Fatal(err)
	}
	if sounds(jazz, 0, 1920, music.D) {
		t.Error(`"none" on jazz still adds the ninth`)
	}

	bare, err := Render(tl, Choice{Tensions: TensionsNinths})
	if err != nil {
		t.Fatal(err)
	}
	if !sounds(bare, 0, 1920, music.D) || bare.Rhythm == nil || bare.Rhythm.Tensions == nil || bare.Rhythm.Tensions.Explanation == "" {
		t.Errorf("block chords without a style still take and explain the ninth: %+v", bare.Rhythm)
	}

	if _, err := Render(tl, Choice{Tensions: "11"}); err == nil || !strings.Contains(err.Error(), `tensions "11"`) {
		t.Errorf("unknown colour: %v", err)
	}
}
