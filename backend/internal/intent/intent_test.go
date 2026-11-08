package intent

import (
	"encoding/json"
	"fmt"
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
		{"bad mode", func(i *Intent) { i.Candidates[0].Key.Mode = music.Mode(9) }, []string{"key.mode must be major or minor"}},
		{"tempo too slow", func(i *Intent) { i.Candidates[0].TempoBPM = 20 }, []string{"tempoBpm must be 40..220"}},
		{"tempo too fast", func(i *Intent) { i.Candidates[1].TempoBPM = 300 }, []string{"candidates[1]: tempoBpm"}},
		{"empty progression", func(i *Intent) { i.Candidates[0].Progression = nil }, []string{"progression must have 1..16"}},
		{"too long progression", func(i *Intent) { i.Candidates[0].Progression = make([]string, MaxChords+1) }, []string{"progression must have 1..16"}},
		{"bad numeral", func(i *Intent) { i.Candidates[0].Progression = []string{"I", "bVII", "V"} }, []string{"progression[1]", "not a roman numeral"}},
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
