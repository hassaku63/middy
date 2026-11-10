package pattern

import (
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Validate reports every problem at once, each with where it is (BL-177).
func TestValidateReportsEveryProblem(t *testing.T) {
	nine := music.TimeSignature{Numerator: 9, Denominator: 8}
	good := Drums{ID: "d", Name: "D", Description: "x", Meter: &fourFour, Hits: []DrumHit{{0, eighth, Kick, 90}}}
	for _, tc := range []struct {
		name string
		cat  Catalogue
		want []string
	}{
		{"a good catalogue", Catalogue{Drums: []Drums{good}}, nil},
		{"missing fields", Catalogue{Drums: []Drums{{Meter: &fourFour}}}, []string{
			`drums[0] "": id is required`, "name is required", "description is required", "hits are required",
		}},
		{"a repeated id", Catalogue{Drums: []Drums{good, good}}, []string{`drums[1] "d": id is used twice`}},
		{"no meter", Catalogue{Drums: []Drums{{ID: "d", Name: "D", Description: "x", Hits: good.Hits}}}, []string{"must declare their meter"}},
		{"an unsupported meter", Catalogue{Drums: []Drums{{ID: "d", Name: "D", Description: "x", Meter: &nine, Hits: good.Hits}}}, []string{"9/8 is not supported"}},
		{"bad hits", Catalogue{Drums: []Drums{{ID: "d", Name: "D", Description: "x", Meter: &threeFour, Hits: []DrumHit{
			{3 * quarter, eighth, Kick, 90}, {0, 0, Kick, 90}, {0, eighth, 200, 90}, {0, eighth, Kick, 0},
		}}}}, []string{
			"hits[0] at 1440 for 240 does not fit a bar of 1440", "hits[1] at 0 for 0", "hits[2] note 200", "hits[3] velocity 0",
		}},
		{"a bad accompaniment", Catalogue{Accompaniments: []Accompaniment{{ID: "a", Name: "A", Description: "x", Hits: []Hit{
			{0, 0, "hum", 80}, {0, 5 * quarter, VoiceChord, 80},
		}}}}, []string{`hits[0] voice "hum" is not a voice`, "hits[1] at 0 for 2400 does not fit a bar of 1920"},
		},
		{"an accompaniment and drums may share an id", Catalogue{
			Accompaniments: []Accompaniment{{ID: "d", Name: "D", Description: "x", Hits: []Hit{{0, 0, VoiceChord, 80}}}},
			Drums:          []Drums{good},
		}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cat.Validate()
			if len(tc.want) == 0 {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() = nil")
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error lacks %q:\n%v", w, err)
				}
			}
		})
	}
}

func TestDecodeCatalogueRefusesUnknownFields(t *testing.T) {
	if _, err := decodeCatalogue([]byte(`{"drums":[{"id":"d","tempo":120}]}`)); err == nil || !strings.Contains(err.Error(), "tempo") {
		t.Fatalf("decodeCatalogue = %v, want an unknown field error", err)
	}
}
