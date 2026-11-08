package pattern

import (
	"errors"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func progression(t *testing.T, key music.Key, beats int, degrees ...int) music.Timeline {
	t.Helper()
	var prog music.Progression
	for _, d := range degrees {
		prog = append(prog, music.Step{Degree: d, Beats: beats})
	}
	tl, err := music.FromProgression(key, prog, 100)
	if err != nil {
		t.Fatal(err)
	}
	return tl
}

func mustAcc(t *testing.T, id string) Accompaniment {
	t.Helper()
	a, err := FindAccompaniment(id)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func notesIn(tl music.Timeline, track string) []music.NoteEvent {
	for _, tr := range tl.Tracks {
		if tr.Name == track {
			return tr.Notes
		}
	}
	return nil
}

func TestCatalogueIsWellFormed(t *testing.T) {
	ids := map[string]bool{}
	for _, a := range Accompaniments() {
		if a.ID == "" || a.Name == "" || a.Description == "" || len(a.Hits) == 0 || ids[a.ID] {
			t.Errorf("bad accompaniment %+v", a)
		}
		ids[a.ID] = true
		length := bar
		if a.Meter != nil {
			length = music.BarTicks(*a.Meter, music.DefaultPPQ)
		}
		for _, h := range a.Hits {
			anyMeter := a.Meter == nil && h.Duration == 0 // lasts to the end of whatever bar
			if h.Start < 0 || h.Start+h.Duration > length || (h.Duration <= 0 && !anyMeter) || h.Velocity < 1 || h.Velocity > 127 || len(pitches([]music.Pitch{60, 64, 67}, h.Voice)) == 0 {
				t.Errorf("%s: bad hit %+v", a.ID, h)
			}
		}
	}
	for _, d := range DrumPatterns() {
		if d.ID == "" || d.Name == "" || d.Description == "" || len(d.Hits) == 0 {
			t.Errorf("bad drums %+v", d)
		}
		if d.Meter == nil {
			t.Errorf("%s: drums must declare their meter", d.ID)
			continue
		}
		for _, h := range d.Hits {
			if h.Start < 0 || h.Start+h.Duration > music.BarTicks(*d.Meter, music.DefaultPPQ) || h.Duration <= 0 || !h.Note.Valid() {
				t.Errorf("%s: bad hit %+v", d.ID, h)
			}
		}
	}
	if _, err := FindAccompaniment("nope"); !errors.Is(err, ErrUnknownPattern) {
		t.Errorf("unknown accompaniment: %v", err)
	}
	if _, err := FindDrums("nope"); !errors.Is(err, ErrUnknownPattern) {
		t.Errorf("unknown drums: %v", err)
	}
}

func TestBlockMatchesFromProgression(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1, 5, 6, 4)
	got, err := Apply(tl, mustAcc(t, "block"), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := tl.Tracks[0].Notes
	gotNotes := notesIn(got, music.ChordTrackName)
	if len(gotNotes) != len(want) {
		t.Fatalf("%d notes, want %d", len(gotNotes), len(want))
	}
	for i := range want {
		if gotNotes[i] != want[i] {
			t.Fatalf("note %d = %+v, want %+v", i, gotNotes[i], want[i])
		}
	}
	if len(got.Tracks) != 1 {
		t.Fatalf("no drums requested, got %d tracks", len(got.Tracks))
	}
	// The input must be untouched.
	if &tl.Tracks[0].Notes[0] == &got.Tracks[0].Notes[0] {
		t.Fatal("Apply must not alias the input")
	}
}

func TestPulseAndBalladShapes(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1) // one bar of C
	pulse, err := Apply(tl, mustAcc(t, "pulse"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(notesIn(pulse, music.ChordTrackName)); n != 4*3 {
		t.Fatalf("pulse: %d notes, want 12 (4 triads)", n)
	}

	ballad, err := Apply(tl, mustAcc(t, "ballad"), nil)
	if err != nil {
		t.Fatal(err)
	}
	notes := notesIn(ballad, music.ChordTrackName)
	if notes[0].Pitch != 48 || notes[0].Start != 0 { // C3 bass on beat 1
		t.Fatalf("ballad first note = %+v, want bass C3 at 0", notes[0])
	}
	if n := len(notes); n != 1+3+1+3 {
		t.Fatalf("ballad: %d notes, want 8", n)
	}
}

func TestArpeggioVoicesAndStrumProgram(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1)
	arp, err := Apply(tl, mustAcc(t, "arpeggio"), nil)
	if err != nil {
		t.Fatal(err)
	}
	var seq []music.Pitch
	for _, n := range notesIn(arp, music.ChordTrackName) {
		seq = append(seq, n.Pitch)
	}
	want := []music.Pitch{60, 64, 67, 72, 60, 64, 67, 72}
	if len(seq) != len(want) {
		t.Fatalf("arpeggio = %v", seq)
	}
	for i := range want {
		if seq[i] != want[i] {
			t.Fatalf("arpeggio = %v, want %v", seq, want)
		}
	}

	strum, err := Apply(tl, mustAcc(t, "strum"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if strum.Tracks[0].Program != 24 {
		t.Fatalf("strum should use a guitar program, got %d", strum.Tracks[0].Program)
	}
}

func TestShortChordsClipTheTemplate(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.A, Mode: music.Minor}, 2, 1, 4) // two half-bar chords
	got, err := Apply(tl, mustAcc(t, "arpeggio"), nil)
	if err != nil {
		t.Fatal(err)
	}
	notes := notesIn(got, music.ChordTrackName)
	if len(notes) != 8 { // 4 eighths per half bar, two chords
		t.Fatalf("%d notes, want 8", len(notes))
	}
	for _, n := range notes {
		if n.End() > 4*music.DefaultPPQ {
			t.Fatalf("note runs past the timeline: %+v", n)
		}
	}
	if notes[4].Pitch != music.NewPitch(music.D, 4) || notes[4].Start != 2*music.DefaultPPQ {
		t.Fatalf("second chord should start on D4 at beat 3, got %+v", notes[4])
	}
}

func TestDrumsRepeatPerBarOnChannelTen(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1, 5)
	d, err := FindDrums("eight-beat")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Apply(tl, mustAcc(t, "block"), &d)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tracks) != 2 || got.Tracks[1].Channel != DrumChannel || got.Tracks[1].Name != "drums" {
		t.Fatalf("drum track missing or misplaced: %+v", got.Tracks)
	}
	perc := got.Tracks[1].Notes
	if len(perc) != 2*len(d.Hits) {
		t.Fatalf("%d drum hits, want %d", len(perc), 2*len(d.Hits))
	}
	kicks := 0
	for _, n := range perc {
		if n.Pitch == Kick {
			kicks++
		}
	}
	if kicks != 4 {
		t.Fatalf("expected 4 kicks over two bars, got %d", kicks)
	}
	if got.LengthTicks() != tl.LengthTicks() {
		t.Fatalf("drums must not extend the timeline: %d vs %d", got.LengthTicks(), tl.LengthTicks())
	}
}

func TestApplyRejectsUnsupportedInput(t *testing.T) {
	tl := progression(t, music.Key{Tonic: music.C, Mode: music.Major}, 4, 1)
	tl.TimeSignature = music.TimeSignature{Numerator: 9, Denominator: 8}
	if _, err := Apply(tl, mustAcc(t, "block"), nil); err == nil || !strings.Contains(err.Error(), "4/4, 3/4 and 6/8") {
		t.Fatalf("9/8 should be rejected for now, got %v", err)
	}
	empty := music.NewTimeline(music.Key{Tonic: music.C, Mode: music.Major}, 100)
	if _, err := Apply(empty, mustAcc(t, "block"), nil); err == nil || !strings.Contains(err.Error(), "no harmony") {
		t.Fatalf("empty harmony should be rejected, got %v", err)
	}
}
