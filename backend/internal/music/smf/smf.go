// Package smf writes a music.Timeline as a Standard MIDI File (format 0).
//
// Format 0 puts every event on one track, which is the simplest file any
// DAW or player accepts. Tracks of the timeline keep their channels and
// programs, so instruments survive the round trip; track names do not
// (a format 1 writer is a backlog idea if that ever matters).
package smf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Write encodes the timeline as SMF format 0 bytes. The timeline must be
// valid (music.Timeline.Validate); callers at the API boundary check that.
func Write(tl music.Timeline) ([]byte, error) {
	if err := tl.Validate(); err != nil {
		return nil, fmt.Errorf("smf: %w", err)
	}
	if tl.PPQ > 0x7FFF {
		return nil, fmt.Errorf("smf: ppq %d exceeds the 15-bit SMF division", tl.PPQ)
	}

	var track bytes.Buffer
	writeMeta(&track, 0, 0x51, tempoBytes(tl.QuarterBPM()))
	writeMeta(&track, 0, 0x58, timeSignatureBytes(tl.TimeSignature))
	writeMeta(&track, 0, 0x59, keySignatureBytes(tl.Key))

	events := collectEvents(tl)
	var last music.Tick
	for _, e := range events {
		writeVarLen(&track, uint32(e.tick-last))
		last = e.tick
		track.Write(e.data)
	}
	writeMeta(&track, 0, 0x2F, nil) // end of track

	var out bytes.Buffer
	out.WriteString("MThd")
	_ = binary.Write(&out, binary.BigEndian, uint32(6))
	_ = binary.Write(&out, binary.BigEndian, uint16(0))      // format 0
	_ = binary.Write(&out, binary.BigEndian, uint16(1))      // one track
	_ = binary.Write(&out, binary.BigEndian, uint16(tl.PPQ)) // ticks per quarter
	out.WriteString("MTrk")
	_ = binary.Write(&out, binary.BigEndian, uint32(track.Len()))
	out.Write(track.Bytes())
	return out.Bytes(), nil
}

// event is a channel message at an absolute tick. kind orders events that
// share a tick: program changes first, then note-offs, then note-ons, so a
// repeated pitch is released before it is struck again.
type event struct {
	tick music.Tick
	kind int
	seq  int
	data []byte
}

func collectEvents(tl music.Timeline) []event {
	var events []event
	seq := 0
	push := func(tick music.Tick, kind int, data ...byte) {
		events = append(events, event{tick: tick, kind: kind, seq: seq, data: data})
		seq++
	}
	for _, tr := range tl.Tracks {
		ch := byte(tr.Channel)
		push(0, 0, 0xC0|ch, byte(tr.Program))
		for _, n := range tr.Notes {
			push(n.Start, 2, 0x90|ch, byte(n.Pitch), byte(n.Velocity))
			push(n.End(), 1, 0x80|ch, byte(n.Pitch), 0)
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].tick != events[j].tick {
			return events[i].tick < events[j].tick
		}
		if events[i].kind != events[j].kind {
			return events[i].kind < events[j].kind
		}
		return events[i].seq < events[j].seq
	})
	return events
}

func writeMeta(w *bytes.Buffer, delta uint32, typ byte, payload []byte) {
	writeVarLen(w, delta)
	w.WriteByte(0xFF)
	w.WriteByte(typ)
	writeVarLen(w, uint32(len(payload)))
	w.Write(payload)
}

// writeVarLen encodes a MIDI variable-length quantity (7 bits per byte,
// high bit set on all but the last byte).
func writeVarLen(w *bytes.Buffer, v uint32) {
	buf := []byte{byte(v & 0x7F)}
	for v >>= 7; v > 0; v >>= 7 {
		buf = append([]byte{byte(v&0x7F) | 0x80}, buf...)
	}
	w.Write(buf)
}

// tempoBytes is microseconds per quarter note as a 24-bit integer.
func tempoBytes(bpm float64) []byte {
	us := uint32(math.Round(60_000_000 / bpm))
	return []byte{byte(us >> 16), byte(us >> 8), byte(us)}
}

// timeSignatureBytes is numerator, log2(denominator), MIDI clocks per
// metronome click, 32nd notes per quarter (8). The click is the meter's
// beat at 24 clocks per quarter: 24 in 4/4 and 3/4, 36 (a dotted quarter)
// in 6/8.
func timeSignatureBytes(ts music.TimeSignature) []byte {
	dd := byte(0)
	for d := ts.Denominator; d > 1; d >>= 1 {
		dd++
	}
	return []byte{byte(ts.Numerator), dd, byte(music.BeatTicks(ts, 24)), 8}
}

// keySignatureBytes is the number of sharps (positive) or flats (negative)
// in -7..7 and 0 for major, 1 for minor. A mode uses its parent major's
// signature (D dorian has none, as C major); the key with six either way
// is written with flats, as the app spells it (Gb major, Eb minor).
func keySignatureBytes(k music.Key) []byte {
	mi := byte(0)
	if k.Mode == music.Minor {
		mi = 1
	}
	// Position on the circle of fifths: tonic = 7*fifths mod 12.
	fifths := (int(k.ParentMajor().Normalize()) * 7) % 12
	if fifths >= 6 {
		fifths -= 12
	}
	return []byte{byte(int8(fifths)), mi}
}
