package smf

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/hassaku63/middy/backend/internal/music"
)

func oneNote() music.Timeline {
	tl := music.NewTimeline(music.Key{Tonic: music.C, Mode: music.Major}, 120)
	tl.Tracks = []music.Track{{
		Name: "piano", Channel: 0, Program: 0,
		Notes: []music.NoteEvent{{Pitch: 60, Velocity: 100, Start: 0, Duration: 480}},
	}}
	return tl
}

// TestWriteExactBytes pins the byte layout for the smallest useful file so
// that any change to the writer is a visible, deliberate diff.
func TestWriteExactBytes(t *testing.T) {
	got, err := Write(oneNote())
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		'M', 'T', 'h', 'd', 0, 0, 0, 6, 0, 0, 0, 1, 0x01, 0xE0, // header, format 0, 1 track, 480 ppq
		'M', 'T', 'r', 'k', 0, 0, 0, 37, // track length
		0x00, 0xFF, 0x51, 0x03, 0x07, 0xA1, 0x20, // tempo 500000 us (120 BPM)
		0x00, 0xFF, 0x58, 0x04, 0x04, 0x02, 0x18, 0x08, // 4/4
		0x00, 0xFF, 0x59, 0x02, 0x00, 0x00, // C major
		0x00, 0xC0, 0x00, // program 0 on channel 0
		0x00, 0x90, 60, 100, // note on
		0x83, 0x60, 0x80, 60, 0, // 480 ticks later: note off
		0x00, 0xFF, 0x2F, 0x00, // end of track
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("bytes differ\n got % X\nwant % X", got, want)
	}
}

func TestVarLen(t *testing.T) {
	tests := []struct {
		v    uint32
		want []byte
	}{
		{0, []byte{0x00}}, {127, []byte{0x7F}}, {128, []byte{0x81, 0x00}},
		{480, []byte{0x83, 0x60}}, {16383, []byte{0xFF, 0x7F}}, {16384, []byte{0x81, 0x80, 0x00}},
		{0x0FFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0x7F}},
	}
	for _, tt := range tests {
		var b bytes.Buffer
		writeVarLen(&b, tt.v)
		if !bytes.Equal(b.Bytes(), tt.want) {
			t.Errorf("varlen(%d) = % X, want % X", tt.v, b.Bytes(), tt.want)
		}
	}
}

func TestKeySignatures(t *testing.T) {
	tests := []struct {
		key  music.Key
		want []byte
	}{
		{music.Key{Tonic: music.C, Mode: music.Major}, []byte{0, 0}},
		{music.Key{Tonic: music.G, Mode: music.Major}, []byte{1, 0}},
		{music.Key{Tonic: music.F, Mode: music.Major}, []byte{0xFF, 0}},  // -1
		{music.Key{Tonic: music.Fs, Mode: music.Major}, []byte{0xFA, 0}}, // spelled Gb major, -6
		{music.Key{Tonic: music.Ds, Mode: music.Minor}, []byte{0xFA, 1}}, // Eb minor, -6
		{music.Key{Tonic: music.D, Mode: music.Dorian}, []byte{0, 0}},    // C major's signature
		{music.Key{Tonic: music.G, Mode: music.Mixolydian}, []byte{0, 0}},
		{music.Key{Tonic: music.E, Mode: music.Dorian}, []byte{2, 0}}, // D major's
		{music.Key{Tonic: music.A, Mode: music.Mixolydian}, []byte{2, 0}},
		{music.Key{Tonic: music.Cs, Mode: music.Major}, []byte{0xFB, 0}}, // Db major, -5
		{music.Key{Tonic: music.A, Mode: music.Minor}, []byte{0, 1}},
		{music.Key{Tonic: music.E, Mode: music.Minor}, []byte{1, 1}},
		{music.Key{Tonic: music.D, Mode: music.Minor}, []byte{0xFF, 1}},
	}
	for _, tt := range tests {
		if got := keySignatureBytes(tt.key); !bytes.Equal(got, tt.want) {
			t.Errorf("%s: % X, want % X", tt.key, got, tt.want)
		}
	}
}

func TestTempoAndTimeSignature(t *testing.T) {
	if got := tempoBytes(90); !bytes.Equal(got, []byte{0x0A, 0x2C, 0x2B}) { // 666667 us
		t.Errorf("tempo 90 = % X", got)
	}
	// The click is the beat: a dotted quarter (36 clocks) in 6/8.
	if got := timeSignatureBytes(music.SixEight); !bytes.Equal(got, []byte{6, 3, 36, 8}) {
		t.Errorf("6/8 = % X", got)
	}
	if got := timeSignatureBytes(music.ThreeFour); !bytes.Equal(got, []byte{3, 2, 24, 8}) {
		t.Errorf("3/4 = % X", got)
	}
}

// TestEventOrdering reads the file back with a minimal parser and checks
// that multi-track timelines merge correctly and that a repeated pitch is
// released before it is struck again.
func TestEventOrdering(t *testing.T) {
	tl := music.NewTimeline(music.Key{Tonic: music.A, Mode: music.Minor}, 100)
	tl.Tracks = []music.Track{
		{Name: "piano", Channel: 0, Program: 0, Notes: []music.NoteEvent{
			{Pitch: 60, Velocity: 80, Start: 0, Duration: 480},
			{Pitch: 60, Velocity: 80, Start: 480, Duration: 480}, // same pitch back to back
		}},
		{Name: "bass", Channel: 1, Program: 33, Notes: []music.NoteEvent{
			{Pitch: 45, Velocity: 90, Start: 240, Duration: 960},
		}},
	}
	data, err := Write(tl)
	if err != nil {
		t.Fatal(err)
	}
	evs := parseTrack(t, data)
	want := []string{
		"0 C0 00", "0 C1 21",
		"0 90 3C 50",
		"240 91 2D 5A",
		"480 80 3C 00", "480 90 3C 50", // off before on at the same tick
		"960 80 3C 00",
		"1200 81 2D 00",
	}
	if len(evs) != len(want) {
		t.Fatalf("got %d events %v, want %d", len(evs), evs, len(want))
	}
	for i := range want {
		if evs[i] != want[i] {
			t.Errorf("event %d = %q, want %q", i, evs[i], want[i])
		}
	}
}

func TestWriteRejectsInvalidTimeline(t *testing.T) {
	tl := oneNote()
	tl.Tracks[0].Notes[0].Velocity = 0
	if _, err := Write(tl); err == nil {
		t.Fatal("invalid timeline must be rejected")
	}
}

// parseTrack returns channel events as "tick STATUS DATA..." strings,
// skipping meta events. It understands just enough SMF to check ordering.
func parseTrack(t *testing.T, data []byte) []string {
	t.Helper()
	if string(data[:4]) != "MThd" || string(data[14:18]) != "MTrk" {
		t.Fatalf("bad chunks")
	}
	length := binary.BigEndian.Uint32(data[18:22])
	track := data[22 : 22+length]
	var out []string
	tick := uint32(0)
	i := 0
	readVar := func() uint32 {
		var v uint32
		for {
			b := track[i]
			i++
			v = v<<7 | uint32(b&0x7F)
			if b&0x80 == 0 {
				return v
			}
		}
	}
	for i < len(track) {
		tick += readVar()
		status := track[i]
		i++
		switch {
		case status == 0xFF:
			typ := track[i]
			i++
			n := int(readVar())
			i += n
			if typ == 0x2F {
				return out
			}
		case status&0xF0 == 0xC0:
			out = append(out, formatEvent(tick, status, track[i]))
			i++
		default:
			out = append(out, formatEvent(tick, status, track[i], track[i+1]))
			i += 2
		}
	}
	t.Fatal("missing end of track")
	return nil
}

func formatEvent(tick uint32, status byte, data ...byte) string {
	s := formatUint(tick) + " " + hex2(status)
	for _, d := range data {
		s += " " + hex2(d)
	}
	return s
}

func hex2(b byte) string {
	const digits = "0123456789ABCDEF"
	return string([]byte{digits[b>>4], digits[b&0xF]})
}

func formatUint(v uint32) string {
	if v == 0 {
		return "0"
	}
	var buf []byte
	for ; v > 0; v /= 10 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
	}
	return string(buf)
}

// TestSixEightTempoIsPerDottedQuarter: 60 BPM in 6/8 is 60 dotted quarters a
// minute, so the file's tempo (microseconds per quarter) is that of 90.
func TestSixEightTempoIsPerDottedQuarter(t *testing.T) {
	tl := oneNote()
	tl.TimeSignature = music.SixEight
	tl.TempoBPM = 60
	data, err := Write(tl)
	if err != nil {
		t.Fatal(err)
	}
	tempo := append([]byte{0xFF, 0x51, 0x03}, tempoBytes(90)...)
	if !bytes.Contains(data, tempo) {
		t.Errorf("6/8 at 60 BPM should write the tempo of 90 quarters a minute: % X", data)
	}
}

// An event further from the previous one than a MIDI variable-length
// quantity can hold is refused, not written as a wrong delta.
func TestWriteRefusesAGapSMFCannotHold(t *testing.T) {
	tl, err := music.FromProgression(music.Key{Tonic: music.C, Mode: music.Major}, music.Progression{{Degree: 1, Beats: 1}}, 120)
	if err != nil {
		t.Fatal(err)
	}
	tl.Tracks[0].Notes[0].Start = 2 * maxVarLen
	if _, err := Write(tl); err == nil || !strings.Contains(err.Error(), "SMF allows") {
		t.Fatalf("want the gap refused, got %v", err)
	}
}
