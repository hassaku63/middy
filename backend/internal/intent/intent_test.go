package intent

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func valid() Intent {
	return Intent{
		Summary: "雨の日曜の午後。静かで少し切ない。",
		Mood:    []string{"静か", "切ない"},
		Candidates: []Candidate{
			{
				Title: "窓辺の雨", Key: music.Key{Tonic: music.A, Mode: music.Minor}, TempoBPM: 72,
				Progression: []string{"i", "VI", "III", "VII"}, BeatsPerChord: 4,
				Reason: "短調と遅めのテンポで落ち着いた切なさを出す。",
			},
			{
				Title: "ゆるやかな午後", Key: music.Key{Tonic: music.F, Mode: music.Major}, TempoBPM: 80,
				Progression: []string{"Imaj7", "vi7", "ii7", "V7"}, BeatsPerChord: 4,
				Reason: "七の和音で柔らかく、少し気だるい雰囲気に。",
			},
		},
	}
}

func TestValidIntentPasses(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateReportsEveryProblem(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Intent)
		want   []string
	}{
		{"empty summary", func(i *Intent) { i.Summary = " " }, []string{"summary is required"}},
		{"long summary", func(i *Intent) { i.Summary = strings.Repeat("あ", MaxSummaryRunes+1) }, []string{"summary is too long"}},
		{"too many moods", func(i *Intent) { i.Mood = make([]string, MaxMoodTags+1) }, []string{"mood has", "mood[0] is empty"}},
		{"no candidates", func(i *Intent) { i.Candidates = nil }, []string{"candidates must have 1..3"}},
		{"too many candidates", func(i *Intent) { i.Candidates = append(i.Candidates, i.Candidates[0], i.Candidates[0]) }, []string{"candidates must have 1..3"}},
		{"empty title", func(i *Intent) { i.Candidates[0].Title = "" }, []string{"candidates[0]: title is required"}},
		{"bad tonic", func(i *Intent) { i.Candidates[0].Key.Tonic = 12 }, []string{"key.tonic must be 0..11"}},
		{"bad mode", func(i *Intent) { i.Candidates[0].Key.Mode = music.Mode(9) }, []string{"key.mode must be major, minor, dorian or mixolydian"}},
		{"tempo too slow", func(i *Intent) { i.Candidates[0].TempoBPM = 20 }, []string{"tempoBpm must be 40..220"}},
		{"tempo too fast", func(i *Intent) { i.Candidates[1].TempoBPM = 300 }, []string{"candidates[1]: tempoBpm"}},
		{"empty progression", func(i *Intent) { i.Candidates[0].Progression = nil }, []string{"progression must have 1..16"}},
		{"too long progression", func(i *Intent) { i.Candidates[0].Progression = make([]string, MaxChords+1) }, []string{"progression must have 1..16"}},
		{"bad numeral", func(i *Intent) { i.Candidates[0].Progression = []string{"i", "X", "V"} }, []string{"progression[1]", "not a roman numeral"}},
		{"bad beats", func(i *Intent) { i.Candidates[0].BeatsPerChord = 3 }, []string{"beatsPerChord must be one of"}},
		{"empty reason", func(i *Intent) { i.Candidates[0].Reason = "" }, []string{"reason is required"}},
		{"two problems at once", func(i *Intent) { i.Summary = ""; i.Candidates[1].Reason = "" }, []string{"summary is required", "candidates[1]: reason is required"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			tt.mutate(&in)
			err := in.Validate()
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q should mention %q", err, w)
				}
			}
		})
	}
}

func TestToProgression(t *testing.T) {
	prog, err := valid().Candidates[1].ToProgression()
	if err != nil {
		t.Fatal(err)
	}
	want := music.Progression{
		{Degree: 1, Seventh: true, Beats: 4}, {Degree: 6, Seventh: true, Beats: 4},
		{Degree: 2, Seventh: true, Beats: 4}, {Degree: 5, Seventh: true, Beats: 4},
	}
	if len(prog) != len(want) {
		t.Fatalf("got %v", prog)
	}
	for i := range want {
		if prog[i] != want[i] {
			t.Errorf("step %d = %+v, want %+v", i, prog[i], want[i])
		}
	}
	// The result must render and analyse without error (end of the chain).
	tl, err := music.FromProgression(valid().Candidates[1].Key, prog, 80)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := music.Analyze(tl); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaIsValidJSONAndMirrorsLimits(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(Schema, &schema); err != nil {
		t.Fatalf("schema is not JSON: %v", err)
	}
	// Structured outputs cannot express most limits, so the descriptions
	// carry them; they must state the same numbers Validate enforces.
	for _, want := range []string{
		fmt.Sprintf("%d 個まで", MaxMoodTags),
		fmt.Sprintf("%d〜%d 個", MinCandidates, MaxCandidates),
		fmt.Sprintf("%d〜%d", MinTempoBPM, MaxTempoBPM),
		fmt.Sprintf("%d 個までのコード", MaxChords),
	} {
		if !strings.Contains(string(Schema), want) {
			t.Errorf("schema description lost the limit %q", want)
		}
	}
	if schema["additionalProperties"] != false {
		t.Error("schema must forbid additional properties so the model cannot smuggle fields")
	}
}

func TestParseRoundTrip(t *testing.T) {
	data, err := json.Marshal(valid())
	if err != nil {
		t.Fatal(err)
	}
	in, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if in.Candidates[0].Key.Mode != music.Minor || in.Candidates[0].Progression[1] != "VI" {
		t.Fatalf("round trip lost data: %+v", in)
	}
	if _, err := Parse([]byte(`{"summary": 1}`)); err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Errorf("type mismatch should be a JSON error, got %v", err)
	}
	if _, err := Parse([]byte(`{"summary":"x","mood":[],"candidates":[]}`)); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("empty candidates should be invalid, got %v", err)
	}
}

func TestPatternIdsAreValidated(t *testing.T) {
	in := valid()
	in.Candidates[0].Accompaniment = "arpeggio"
	in.Candidates[0].Drums = "eight-beat"
	if err := in.Validate(); err != nil {
		t.Fatalf("known ids rejected: %v", err)
	}
	in.Candidates[0].Accompaniment = "techno"
	in.Candidates[1].Drums = "trap"
	err := in.Validate()
	if err == nil || !strings.Contains(err.Error(), "accompaniment must be one of") || !strings.Contains(err.Error(), "drums must be one of") {
		t.Fatalf("unknown ids should be reported: %v", err)
	}
	if len(AccompanimentIDs()) < 5 || len(DrumIDs()) < 3 {
		t.Fatalf("catalogue ids: %v %v", AccompanimentIDs(), DrumIDs())
	}
	if !strings.Contains(string(Schema), `"arpeggio"`) || !strings.Contains(string(Schema), `"eight-beat"`) {
		t.Fatal("schema enums must list the catalogue ids")
	}
}

func TestParseIgnoresTheCaseOfEnumValues(t *testing.T) {
	// The structured-outputs docs say enum case is not guaranteed.
	data := []byte(`{"summary":"s","mood":[],"candidates":[{"title":"t",
	  "key":{"tonic":9,"mode":"Minor"},"tempoBpm":80,"progression":["i","iv"],
	  "beatsPerChord":4,"reason":"r","accompaniment":"Arpeggio","drums":"EIGHT-BEAT"}]}`)
	in, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	c := in.Candidates[0]
	if c.Key.Mode != music.Minor || c.Accompaniment != "arpeggio" || c.Drums != "eight-beat" {
		t.Fatalf("enum values not normalised: %+v", c)
	}
	if c.Progression[0] != "i" {
		t.Fatalf("roman numerals carry meaning in their case and must stay as written: %v", c.Progression)
	}
}

func TestFeelIsValidatedAndInTheSchema(t *testing.T) {
	in := valid()
	in.Candidates[0].Feel = "swing"
	if err := in.Validate(); err != nil {
		t.Fatalf("swing is allowed: %v", err)
	}
	in.Candidates[0].Feel = "shuffle"
	if err := in.Validate(); err == nil || !strings.Contains(err.Error(), "feel must be one of") {
		t.Fatalf("unknown feel must be reported, got %v", err)
	}
	if !strings.Contains(string(Schema), `"swing"`) {
		t.Fatal("the schema must offer the feels")
	}
	parsed, err := Parse([]byte(`{"summary":"s","mood":[],"candidates":[{"title":"t","key":{"tonic":0,"mode":"major"},"tempoBpm":120,"progression":["I","V"],"beatsPerChord":4,"reason":"r","accompaniment":"pulse","feel":"Swing"}]}`))
	if err != nil || parsed.Candidates[0].Feel != "swing" {
		t.Fatalf("feel case is normalised: %+v %v", parsed, err)
	}
}

func TestLabelsThatTheKeyWouldChangeAreRejectedWithTheFix(t *testing.T) {
	in := valid() // A minor
	in.Candidates[0].Progression = []string{"i", "II", "V"}
	err := in.Validate()
	if err == nil || !strings.Contains(err.Error(), `"II" would sound as Bdim (ii°) in A minor; write "ii°"`) {
		t.Fatalf("a major II must be rejected with the engine's spelling, got %v", err)
	}
	in.Candidates[0].Progression = []string{"i", "iv", "V"}
	if err := in.Validate(); err != nil {
		t.Fatalf("iv and the harmonic V are fine: %v", err)
	}
}

func TestParseSpellsLabelsAsTheEngineDoes(t *testing.T) {
	data := []byte(`{"summary":"s","mood":[],"candidates":[{"title":"t","key":{"tonic":0,"mode":"major"},"tempoBpm":100,"progression":["I","vii","viim7b5","V7"],"beatsPerChord":4,"reason":"r"}]}`)
	in, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(in.Candidates[0].Progression, " ")
	if got != "I vii° viiø7 V7" {
		t.Fatalf("labels not canonicalised: %q", got)
	}
}

// TestSixEightCandidate: 6/8 is counted in dotted-quarter beats, so a chord
// lasts half a bar (1), a bar (2) or two bars (4); the patterns must be the
// 6/8 ones, and swing is refused because the beat already divides in three.
func TestSixEightCandidate(t *testing.T) {
	ok := func(c *Candidate) {
		c.Meter, c.BeatsPerChord, c.Accompaniment, c.Drums = "6/8", 2, "rocking-6", "six-eight"
	}
	tests := []struct {
		name   string
		mutate func(*Candidate)
		want   string // "" means valid
	}{
		{"one bar", func(c *Candidate) {}, ""},
		{"half a bar", func(c *Candidate) { c.BeatsPerChord = 1 }, ""},
		{"two bars", func(c *Candidate) { c.BeatsPerChord = 4 }, ""},
		{"three beats", func(c *Candidate) { c.BeatsPerChord = 3 }, "beatsPerChord must be one of [1 2 4] in 6/8"},
		{"a 3/4 pattern", func(c *Candidate) { c.Accompaniment = "waltz" }, "accompaniment must be one of"},
		{"4/4 drums", func(c *Candidate) { c.Drums = "eight-beat" }, "drums must be one of"},
		{"swing", func(c *Candidate) { c.Feel = "swing" }, "swing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			ok(&in.Candidates[0])
			tt.mutate(&in.Candidates[0])
			err := in.Validate()
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("should be valid: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error %v should mention %q", err, tt.want)
			}
		})
	}
	if got := valid().Candidates[0]; got.TimeSignature() != music.FourFour {
		t.Errorf("an omitted meter is 4/4")
	}
	c := valid().Candidates[0]
	ok(&c)
	if c.TimeSignature() != music.SixEight {
		t.Errorf("6/8 should map to music.SixEight")
	}
}

// TestPerChordBeats (BL-109): beats gives each chord its own length in
// beats of the meter. Each length is one the meter allows; a chord
// shorter than a bar starts at a multiple of its length within the bar
// (so it never crosses a bar line), a longer one starts on a bar line, and
// the progression ends on a bar line.
func TestPerChordBeats(t *testing.T) {
	tests := []struct {
		name  string
		meter string
		prog  []string
		beats []int
		want  string // "" means valid
	}{
		{"hold I two bars, then two beats each", "", []string{"i", "VI", "III", "VII"}, []int{8, 4, 2, 2}, ""},
		{"omitted uses beatsPerChord", "", []string{"i", "VI", "III", "VII"}, nil, ""},
		{"length mismatch", "", []string{"i", "VI", "III", "VII"}, []int{8, 4}, "beats has 2 lengths for 4 chords"},
		{"length the meter does not allow", "", []string{"i", "VI", "III", "VII"}, []int{3, 1, 4, 8}, "beats[0] must be one of [1 2 4 8]"},
		{"half bar on beat 2", "", []string{"i", "VI", "III", "VII"}, []int{1, 2, 1, 4}, "beats[1]: a chord of 2 beats must start on beat 1 or 3"},
		{"bar-long chord mid-bar", "", []string{"i", "VI", "III", "VII"}, []int{2, 4, 2, 4}, "beats[1]: a chord of 4 beats must start on a bar line"},
		{"ends mid-bar", "", []string{"i", "VI", "III", "VII"}, []int{4, 4, 4, 2}, "the progression ends mid-bar"},
		{"6/8 half bars", "6/8", []string{"i", "VI", "III", "VII"}, []int{2, 1, 1, 2}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			c := &in.Candidates[0]
			c.Progression, c.Beats = tt.prog, tt.beats
			if tt.meter == "6/8" {
				c.Meter, c.BeatsPerChord = "6/8", 2
			}
			err := in.Validate()
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("should be valid: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error %v should mention %q", err, tt.want)
			}
		})
	}
	c := valid().Candidates[0]
	c.Beats = []int{8, 4, 2, 2}
	prog, err := c.ToProgression()
	if err != nil {
		t.Fatal(err)
	}
	var got []int
	for _, s := range prog {
		got = append(got, s.Beats)
	}
	if fmt.Sprint(got) != "[8 4 2 2]" {
		t.Errorf("steps carry the per-chord beats: %v", got)
	}
}

// TestRetryHints (#133): an error the model makes again and again carries
// the fix, so the retry can act on it: a misspelt label gets the label one
// deletion away, and a length or pattern of another meter names that meter.
func TestRetryHints(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Candidate)
		want   string
	}{
		{"misspelt maj7", func(c *Candidate) {
			c.Key = music.Key{Tonic: music.C, Mode: music.Major}
			c.Progression = []string{"Imaxj7", "IV"}
		},
			`did you mean "Imaj7"?`},
		{"waltz length without the meter", func(c *Candidate) { c.BeatsPerChord = 3 },
			`3 is a length in 3/4; set meter to "3/4" to use it`},
		{"waltz pattern without the meter", func(c *Candidate) { c.Accompaniment = "arpeggio-3" },
			`"arpeggio-3" is a 3/4 pattern; set meter to "3/4" to use it`},
		{"6/8 drums without the meter", func(c *Candidate) { c.Drums = "jig" },
			`"jig" is a 6/8 pattern; set meter to "6/8" to use it`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			tt.mutate(&in.Candidates[0])
			err := in.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %v should say %q", err, tt.want)
			}
		})
	}
	in := valid()
	in.Candidates[0].Progression = []string{"Q7"}
	if err := in.Validate(); err == nil || strings.Contains(err.Error(), "did you mean") {
		t.Errorf("no suggestion when no label is one deletion away: %v", err)
	}
	// A label that reads but names another chord already says what to
	// write; a deletion ("VImaj7") would point at a different chord.
	in = valid()
	in.Candidates[0].Key = music.Key{Tonic: music.D, Mode: music.Minor}
	in.Candidates[0].Progression = []string{"i", "VIImaj7"}
	err := in.Validate()
	if err == nil || !strings.Contains(err.Error(), `write "VII7"`) || strings.Contains(err.Error(), "did you mean") {
		t.Errorf("a label that reads keeps only the engine's correction: %v", err)
	}
}

// TestSlashChordCandidates (BL-115): the slash chords the prompt shows as
// examples pass validation, and a bass degree out of range is refused.
func TestSlashChordCandidates(t *testing.T) {
	for _, tt := range []struct {
		key  music.Key
		prog []string
	}{
		{music.Key{Tonic: music.C, Mode: music.Major}, []string{"I", "I/7", "vi", "vi/5", "IV", "I/3", "ii7", "V7"}},
		{music.Key{Tonic: music.A, Mode: music.Minor}, []string{"i", "i/#7", "i/7", "i/#6"}},
		{music.Key{Tonic: music.C, Mode: music.Major}, []string{"I", "IV/1", "V7/V/#4", "V"}},
	} {
		in := valid()
		in.Candidates[0].Key, in.Candidates[0].Progression = tt.key, tt.prog
		if err := in.Validate(); err != nil {
			t.Errorf("%v should be valid: %v", tt.prog, err)
		}
	}
	in := valid()
	in.Candidates[0].Progression = []string{"i", "i/9"}
	if err := in.Validate(); err == nil || !strings.Contains(err.Error(), "scale degree 1..7") {
		t.Errorf("i/9 should be refused with the reason: %v", err)
	}
}

// TestStyleBindsThePatterns (BL-082): a candidate's style must be known,
// and its patterns and feel must belong to the style; the error says what
// would fit. A candidate without a style is not checked (older answers).
func TestStyleBindsThePatterns(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Candidate)
		want   string // "" means valid
	}{
		{"celtic reel", func(c *Candidate) { c.Style, c.Accompaniment, c.Drums = "celtic", "reel", "bodhran" }, ""},
		{"celtic on the eight-beat", func(c *Candidate) { c.Style, c.Accompaniment, c.Drums = "celtic", "strum", "eight-beat" },
			`accompaniment "strum" does not fit the celtic style`},
		{"jazz with swing", func(c *Candidate) { c.Style, c.Accompaniment, c.Drums, c.Feel = "jazz", "comp", "jazz-ride", "swing" }, ""},
		{"unknown style", func(c *Candidate) { c.Style = "polka" }, "style must be one of"},
		{"no style", func(c *Candidate) { c.Accompaniment, c.Drums = "strum", "eight-beat" }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			tt.mutate(&in.Candidates[0])
			err := in.Validate()
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("should be valid: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error %v should mention %q", err, tt.want)
			}
		})
	}
	if !strings.Contains(string(Schema), `"style"`) || !slices.Contains(StyleIDs(), "celtic") {
		t.Error("the schema should ask for a style from the catalogue")
	}
}

// TestPushes (BL-085): a candidate may push chords, one entry per chord;
// the first chord cannot push, and the steps carry the pushes.
func TestPushes(t *testing.T) {
	tests := []struct {
		name   string
		pushes []string
		want   string // "" means valid
	}{
		{"push the last chord an eighth", []string{"none", "none", "none", "eighth"}, ""},
		{"omitted pushes nothing", nil, ""},
		{"length mismatch", []string{"none", "eighth"}, "pushes has 2 entries for 4 chords"},
		{"unknown push", []string{"none", "quarter", "none", "none"}, `pushes[1] must be one of [none eighth sixteenth], got "quarter"`},
		{"first chord", []string{"eighth", "none", "none", "none"}, "pushes[0]: the first chord has nothing to push into"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			in.Candidates[0].Pushes = tt.pushes
			err := in.Validate()
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("should be valid: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error %v should mention %q", err, tt.want)
			}
		})
	}
	c := valid().Candidates[0]
	c.Pushes = []string{"none", "sixteenth", "none", "eighth"}
	prog, err := c.ToProgression()
	if err != nil {
		t.Fatal(err)
	}
	var got []music.Push
	for _, s := range prog {
		got = append(got, s.Push)
	}
	if !slices.Equal(got, []music.Push{music.PushNone, music.PushSixteenth, music.PushNone, music.PushEighth}) {
		t.Errorf("steps carry the pushes: %q", got)
	}
	if !strings.Contains(string(Schema), `"pushes"`) {
		t.Error("the schema offers pushes")
	}
}

// TestDrumGrids (BL-077): a candidate may write its drums as a grid; the
// engine's checks become retry errors, a grid and a drums id do not mix,
// and basedOn names catalogue drums that fit the meter and the style.
func TestDrumGrids(t *testing.T) {
	backbeat := &GridSpec{Division: 2, Voices: []GridLine{{"kick", "x... x.x."}, {"snare", "..x. ..x."}, {"hat", "xxxx xxxx"}}}
	tests := []struct {
		name string
		edit func(c *Candidate)
		want string // "" means valid
	}{
		{"a grid in pop", func(c *Candidate) { c.Style, c.DrumGrid = "pop", backbeat }, ""},
		{"based on the eight-beat", func(c *Candidate) {
			g := *backbeat
			g.BasedOn = "eight-beat"
			c.Style, c.DrumGrid = "pop", &g
		}, ""},
		{"grid and drums", func(c *Candidate) { c.DrumGrid, c.Drums = backbeat, "eight-beat" }, "write drums or drumGrid, not both"},
		{"bad grid", func(c *Candidate) {
			c.DrumGrid = &GridSpec{Division: 2, Voices: []GridLine{{"kick", "x..."}}}
		}, `drumGrid: pattern: grid voices["kick"] has 4 steps, want 8`},
		{"a voice twice", func(c *Candidate) {
			c.DrumGrid = &GridSpec{Division: 2, Voices: []GridLine{{"kick", "x......."}, {"kick", "....x..."}}}
		}, `voice "kick" is written twice`},
		{"basedOn of another style", func(c *Candidate) {
			g := *backbeat
			g.BasedOn = "jazz-ride"
			c.Style, c.DrumGrid = "pop", &g
		}, `drums "jazz-ride" does not fit the pop style`},
		{"basedOn unknown", func(c *Candidate) {
			g := *backbeat
			g.BasedOn = "polka"
			c.DrumGrid = &g
		}, `drumGrid: pattern: grid basedOn "polka" is not drums written for 4/4`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid()
			tt.edit(&in.Candidates[0])
			err := in.Validate()
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("should be valid: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error %v should mention %q", err, tt.want)
			}
		})
	}
	if !strings.Contains(string(Schema), `"drumGrid"`) || !strings.Contains(string(Schema), `"cross-stick"`) {
		t.Error("the schema offers the grid and its voices")
	}
}

// TestNoSuggestionForSlashLabels (review): V/IV in C is rejected with the
// engine's fix ("write I"); no one-deletion guess may contradict it.
func TestNoSuggestionForSlashLabels(t *testing.T) {
	in := valid()
	in.Candidates[0].Key = music.Key{Tonic: music.C, Mode: music.Major}
	in.Candidates[0].Progression = []string{"I", "V/IV", "IV", "V"}
	err := in.Validate()
	if err == nil || strings.Contains(err.Error(), "did you mean") {
		t.Errorf("want the engine's fix alone, got %v", err)
	}
}

// TestOneErrorPerBadPattern (review): an unknown drums id is reported once
// with the ids to use, not again as a style misfit.
func TestOneErrorPerBadPattern(t *testing.T) {
	in := valid()
	in.Candidates[0].Style, in.Candidates[0].Drums = "jazz", "jazz-ridee"
	err := in.Validate()
	if err == nil || !strings.Contains(err.Error(), "drums must be one of") || strings.Contains(err.Error(), "does not fit the jazz style") {
		t.Errorf("one error for the bad drums id, got %v", err)
	}
}

// TestParseIgnoresTheCaseOfEveryEnum (review of BL-062): style, pushes and
// the grid's voices and basedOn are lowercased like the pattern ids, so a
// re-cased enum is not a retry.
func TestParseIgnoresTheCaseOfEveryEnum(t *testing.T) {
	data := []byte(`{"summary":"s","mood":[],"candidates":[{"title":"t","style":"Pop",
	  "key":{"tonic":0,"mode":"major"},"tempoBpm":120,"progression":["I","V","vi","IV"],
	  "beatsPerChord":4,"reason":"r","pushes":["NONE","none","none","Eighth"],
	  "drumGrid":{"division":2,"basedOn":"Eight-Beat","voices":[{"voice":"Kick","steps":"x... x..."},{"voice":"SNARE","steps":"..x. ..x."}]}}]}`)
	in, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	c := in.Candidates[0]
	if c.Style != "pop" || c.Pushes[3] != "eighth" || c.DrumGrid.BasedOn != "eight-beat" || c.DrumGrid.Voices[0].Voice != "kick" {
		t.Errorf("enums not normalised: %+v %+v", c, c.DrumGrid)
	}
}

// TestDominantHint (review): "II7" in C major gets the secondary dominant
// it most likely meant, alongside the engine's own fix.
func TestDominantHint(t *testing.T) {
	in := valid()
	in.Candidates[0].Key = music.Key{Tonic: music.C, Mode: music.Major}
	in.Candidates[0].Progression = []string{"I", "II7", "V7", "I"}
	err := in.Validate()
	if err == nil || !strings.Contains(err.Error(), `write "V7/V"`) {
		t.Errorf("want the V7/V hint, got %v", err)
	}
}

// TestEachPatternMistakeIsReportedOnce (review): a bad feel or a bad grid
// basedOn gets its own error with the ids to use, not a second one as a
// style misfit; drums written beside a grid are still held to the style;
// an out-of-range tempo does not also ask for a coarser grid.
func TestEachPatternMistakeIsReportedOnce(t *testing.T) {
	grid := func(basedOn string, division int) *GridSpec {
		steps := strings.Repeat("x", division*4)
		return &GridSpec{Division: division, BasedOn: basedOn, Voices: []GridLine{{Voice: "hat", Steps: steps}}}
	}
	cases := []struct {
		name    string
		edit    func(c *Candidate)
		want    string
		wantNot string
	}{
		{"unknown feel", func(c *Candidate) { c.Style, c.Feel = "jazz", "shuffle" }, "feel must be one of", "does not fit the jazz style"},
		{"basedOn for another meter", func(c *Candidate) { c.Style, c.DrumGrid = "jazz", grid("waltz", 2) }, "basedOn", "does not fit the jazz style"},
		{"drums beside a grid", func(c *Candidate) { c.Style, c.Drums, c.DrumGrid = "jazz", "eight-beat", grid("", 2) }, "does not fit the jazz style", ""},
		{"tempo out of range", func(c *Candidate) { c.TempoBPM, c.DrumGrid = 300, grid("", 4) }, "tempoBpm must be", "coarser"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid()
			tc.edit(&in.Candidates[0])
			err := in.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
			if tc.wantNot != "" && strings.Contains(err.Error(), tc.wantNot) {
				t.Errorf("reported twice (%q): %v", tc.wantNot, err)
			}
		})
	}
}

// TestParseTrimsEnums (review): " arpeggio" or "swing " is the id, not a retry.
func TestParseTrimsEnums(t *testing.T) {
	data := []byte(`{"summary":"s","mood":[],"candidates":[{"title":"t",
	  "key":{"tonic":0,"mode":"major"},"tempoBpm":120,"progression":["I","V","vi","IV"],
	  "beatsPerChord":4,"reason":"r","accompaniment":" arpeggio","feel":"swing "}]}`)
	in, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if c := in.Candidates[0]; c.Accompaniment != "arpeggio" || c.Feel != "swing" {
		t.Errorf("enums not trimmed: %q %q", c.Accompaniment, c.Feel)
	}
}

// TestTritoneSubstitutesInTheIntent (BL-129): subV7/x is a label the model
// may write, and a substitute written by its root (bII7) gets the label to
// use instead.
func TestTritoneSubstitutesInTheIntent(t *testing.T) {
	in := valid()
	in.Candidates[0].Key = music.Key{Tonic: music.C, Mode: music.Major}
	in.Candidates[0].Progression = []string{"ii7", "subV7/I", "Imaj7", "Imaj7"}
	if err := in.Validate(); err != nil {
		t.Fatalf("subV7/I rejected: %v", err)
	}
	in.Candidates[0].Progression = []string{"ii7", "bII7", "Imaj7", "Imaj7"}
	err := in.Validate()
	if err == nil || !strings.Contains(err.Error(), `write "subV7/I"`) {
		t.Errorf("want the subV7/I hint for bII7, got %v", err)
	}
	in.Candidates[0].Progression = []string{"I", "bVI7", "V7", "I"}
	if err := in.Validate(); err == nil || !strings.Contains(err.Error(), `write "subV7/V"`) {
		t.Errorf("want the subV7/V hint for bVI7, got %v", err)
	}
}

// TestPassingDiminishedInTheIntent (BL-131): vii°7/x is a label the model
// may write; a diminished seventh written on a sharpened root (#i°7) gets
// the label to use; and a minor key's V7/V is read as V7/v.
func TestPassingDiminishedInTheIntent(t *testing.T) {
	in := valid()
	in.Candidates[0].Key = music.Key{Tonic: music.C, Mode: music.Major}
	in.Candidates[0].Progression = []string{"I", "vii°7/ii", "ii7", "V7"}
	if err := in.Validate(); err != nil {
		t.Fatalf("vii°7/ii rejected: %v", err)
	}
	in.Candidates[0].Progression = []string{"I", "#i°7", "ii7", "V7"}
	if err := in.Validate(); err == nil || !strings.Contains(err.Error(), `write "vii°7/ii"`) {
		t.Errorf("want the vii°7/ii hint for #i°7, got %v", err)
	}
	minor := valid()
	minor.Candidates[0].Progression = []string{"i", "V7/V", "V7", "i"}
	if err := minor.Validate(); err != nil {
		t.Errorf("V7/V in A minor rejected: %v", err)
	}
}

// TestOpenChordsAndBluesDominantsInTheIntent (BL-133): sus4, 7sus4, power
// chords and a major key's blues dominants are labels the model may write;
// a sus4 that needs a note outside the key is still refused.
func TestOpenChordsAndBluesDominantsInTheIntent(t *testing.T) {
	in := valid()
	in.Candidates[0].Key = music.Key{Tonic: music.C, Mode: music.Major}
	for _, prog := range [][]string{
		{"I", "Vsus4", "V", "I"},
		{"I", "IV", "V7sus4", "V7"},
		{"I5", "bVII5", "IV5", "I5"},
		{"I7", "IV7", "I7", "V7"},
	} {
		in.Candidates[0].Progression = prog
		if err := in.Validate(); err != nil {
			t.Errorf("%v rejected: %v", prog, err)
		}
	}
	in.Candidates[0].Progression = []string{"I", "IVsus4", "I"}
	if err := in.Validate(); err == nil || !strings.Contains(err.Error(), "IVsus4") {
		t.Errorf("IVsus4 in C major needs Bb and must be refused, got %v", err)
	}
}
