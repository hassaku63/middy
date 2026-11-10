// Package pattern turns a harmony-only timeline into something that
// grooves: an accompaniment template (how the chord tones are played
// across a bar) and a drum template (GM percussion on channel 10). Both
// are one-bar figures at music.DefaultPPQ in the meter they declare (4/4,
// 3/4 or 6/8, ADR-0017), repeated over every harmony event; a chord shorter
// than a bar simply gets the first part.
//
// Templates are data so that the explanation layer (backlog BL-022) can
// describe them and the LLM (BL-023) can pick them by id.
package pattern

import (
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/hassaku63/middy/backend/internal/music"
)

// Voice selects which chord tones a hit plays, relative to the chord's
// close root-position voicing at music.ChordOctave.
type Voice string

// Voices.
const (
	VoiceChord  Voice = "chord"  // every chord tone together
	VoiceBass   Voice = "bass"   // the root one octave below the voicing
	VoiceRoot   Voice = "root"   // the root of the voicing
	VoiceThird  Voice = "third"  // the second chord tone
	VoiceFifth  Voice = "fifth"  // the third chord tone
	VoiceTop    Voice = "top"    // the highest chord tone (seventh when present)
	VoiceOctave Voice = "octave" // the root one octave above the voicing
	// VoiceDrone is the key's tonic and fifth below the chords: it does
	// not follow the chord, so it keeps sounding as the harmony moves.
	VoiceDrone Voice = "drone"
	// VoiceWalk is a walking bass (BL-086): the root, the chord's tones,
	// and a semitone into the next chord's root on the chord's last walk.
	VoiceWalk Voice = "walk"
)

// Hit is one event of an accompaniment template, positioned within a bar.
// A zero Duration lasts to the end of the bar, so a template that fits any
// meter (block) needs no bar length of its own.
type Hit struct {
	Start    music.Tick `json:"start"`
	Duration music.Tick `json:"duration"`
	Voice    Voice      `json:"voice"`
	Velocity int        `json:"velocity"`
}

// Accompaniment is a one-bar figure for the chord track.
type Accompaniment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Program     int    `json:"program"` // GM program for the track
	// Meter is the time signature the hits are written for; nil fits any.
	Meter *music.TimeSignature `json:"meter,omitempty"`
	Hits  []Hit                `json:"hits"`
}

// Fits reports whether the template can play in ts.
func (a Accompaniment) Fits(ts music.TimeSignature) bool { return a.Meter == nil || *a.Meter == ts }

// DrumHit is one percussion event within a bar. Note uses the GM drum map.
type DrumHit struct {
	Start    music.Tick  `json:"start"`
	Duration music.Tick  `json:"duration"`
	Note     music.Pitch `json:"note"`
	Velocity int         `json:"velocity"`
}

// Drums is a one-bar percussion figure.
type Drums struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Meter is the time signature the hits are written for.
	Meter *music.TimeSignature `json:"meter,omitempty"`
	Hits  []DrumHit            `json:"hits"`
}

// Fits reports whether the figure can play in ts.
func (d Drums) Fits(ts music.TimeSignature) bool { return d.Meter == nil || *d.Meter == ts }

// Durations in ticks at DefaultPPQ; bar is a 4/4 bar.
const (
	bar     = music.Tick(4 * music.DefaultPPQ)
	quarter = music.Tick(music.DefaultPPQ)
	eighth  = music.Tick(music.DefaultPPQ / 2)
)

var (
	fourFour  = music.FourFour
	threeFour = music.ThreeFour
	sixEight  = music.SixEight
)

// GM drum map notes used by the templates.
const (
	Kick      music.Pitch = 36
	SideStick music.Pitch = 37
	Snare     music.Pitch = 38
	ClosedHat music.Pitch = 42
	PedalHat  music.Pitch = 44
	LowTom    music.Pitch = 45
	OpenHat   music.Pitch = 46
	MidTom    music.Pitch = 47
	Ride      music.Pitch = 51
	RideBell  music.Pitch = 53
)

// DrumChannel is the General MIDI percussion channel (0-based).
const DrumChannel = 9

// ErrUnknownPattern is returned by the lookup functions.
var ErrUnknownPattern = errors.New("pattern: unknown pattern id")

// Accompaniments lists the catalogue in a stable order.
func Accompaniments() []Accompaniment {
	out := make([]Accompaniment, len(accompaniments))
	copy(out, accompaniments)
	return out
}

// AccompanimentsFor lists the templates that can play in ts.
func AccompanimentsFor(ts music.TimeSignature) []Accompaniment {
	var out []Accompaniment
	for _, a := range accompaniments {
		if a.Fits(ts) {
			out = append(out, a)
		}
	}
	return out
}

// DrumPatternsFor lists the figures that can play in ts.
func DrumPatternsFor(ts music.TimeSignature) []Drums {
	var out []Drums
	for _, d := range drums {
		if d.Fits(ts) {
			out = append(out, d)
		}
	}
	return out
}

// DrumPatterns lists the catalogue in a stable order.
func DrumPatterns() []Drums {
	out := make([]Drums, len(drums))
	copy(out, drums)
	return out
}

// FindAccompaniment looks a template up by id.
func FindAccompaniment(id string) (Accompaniment, error) {
	for _, a := range accompaniments {
		if a.ID == id {
			return a, nil
		}
	}
	return Accompaniment{}, fmt.Errorf("%w: accompaniment %q", ErrUnknownPattern, id)
}

// FindDrums looks a template up by id.
func FindDrums(id string) (Drums, error) {
	for _, d := range drums {
		if d.ID == id {
			return d, nil
		}
	}
	return Drums{}, fmt.Errorf("%w: drums %q", ErrUnknownPattern, id)
}

// Apply rebuilds the timeline's tracks from its Harmony: the chord track
// follows acc, and a percussion track follows dr when it is not nil. The
// input is not modified. The timeline must be in 4/4, 3/4 or 6/8 at
// DefaultPPQ, and the templates must be written for its meter (ADR-0017).
func Apply(tl music.Timeline, acc Accompaniment, dr *Drums) (music.Timeline, error) {
	return apply(tl, acc, dr, nil)
}

// apply is Apply with the style's tensions (BL-125), which colour the
// chord voice of every pattern.
func apply(tl music.Timeline, acc Accompaniment, dr *Drums, tensions *Tensions) (music.Timeline, error) {
	ts := tl.TimeSignature
	if (ts != fourFour && ts != threeFour && ts != sixEight) || tl.PPQ != music.DefaultPPQ {
		return music.Timeline{}, fmt.Errorf("pattern: templates are defined for 4/4, 3/4 and 6/8 at %d ppq, got %d/%d at %d",
			music.DefaultPPQ, ts.Numerator, ts.Denominator, tl.PPQ)
	}
	if !acc.Fits(ts) {
		return music.Timeline{}, fmt.Errorf("pattern: accompaniment %q is written for %d/%d, the timeline is in %d/%d",
			acc.ID, acc.Meter.Numerator, acc.Meter.Denominator, ts.Numerator, ts.Denominator)
	}
	if dr != nil && !dr.Fits(ts) {
		return music.Timeline{}, fmt.Errorf("pattern: drums %q are written for %d/%d, the timeline is in %d/%d",
			dr.ID, dr.Meter.Numerator, dr.Meter.Denominator, ts.Numerator, ts.Denominator)
	}
	if len(tl.Harmony) == 0 {
		return music.Timeline{}, errors.New("pattern: timeline has no harmony to play")
	}

	out := tl
	out.Tracks = []music.Track{chordTrack(tl, acc, tensions)}
	out.Rhythm = &music.Rhythm{Accompaniment: describeAccompaniment(acc, ts)}
	if dr != nil {
		out.Tracks = append(out.Tracks, drumTrack(tl, *dr))
		out.Rhythm.Drums = describeDrums(*dr, ts)
		out.Rhythm.Drums.Explanation += fillText(*dr, ts, tl.LengthTicks())
	}
	if err := out.Validate(); err != nil {
		return music.Timeline{}, fmt.Errorf("pattern: %w", err)
	}
	return out, nil
}

// chordTrack realises the accompaniment over every harmony event.
func chordTrack(tl music.Timeline, acc Accompaniment, tensions *Tensions) music.Track {
	track := music.Track{Name: music.ChordTrackName, Channel: 0, Program: acc.Program}
	bar := music.BarTicks(tl.TimeSignature, tl.PPQ)
	led := music.LeadVoicings(chordsOf(tl.Harmony)) // BL-086
	for i := range tl.Harmony {
		adds := tensionsOf(tensions, tl.Key, tl.Harmony[i])
		chord := colour(led[i], adds)
		track.Notes = append(track.Notes, chordNotes(tl, acc, i, led[i], chord, adds, bar)...)
	}
	track.Notes = music.PushNotes(tl.Harmony, track.Notes, tl.PPQ) // before the feel (BL-085)
	track.Notes = holdDrone(track.Notes, acc, tl.Key)              // after the pushes, which cut at each chord
	sortNotes(track.Notes)
	return track
}

// chordNotes realises the accompaniment over chord i, a bar of the
// template at a time from the chord's start; the led voicing plays the
// chord, and the walking bass is filled in once its hits are known. An
// arpeggio, which has no chord hit to colour, plays the style's first
// tension in place of its octave (#307).
func chordNotes(tl music.Timeline, acc Accompaniment, i int, led, chord []music.Pitch, adds []music.PitchClass, bar music.Tick) []music.NoteEvent {
	h := tl.Harmony[i]
	voicing := h.Chord.Voicing(music.ChordOctave)
	end := h.Start + h.Duration
	var notes, walks []music.NoteEvent
	for barStart := h.Start; barStart < end; barStart += bar {
		for _, hit := range acc.Hits {
			start := barStart + hit.Start
			if start >= end {
				continue
			}
			dur := hit.Duration
			if dur == 0 {
				dur = bar - hit.Start // to the end of the bar
			}
			n := music.NoteEvent{Velocity: hit.Velocity, Start: start, Duration: min(dur, end-start)}
			if hit.Voice == VoiceWalk {
				walks = append(walks, n)
				continue
			}
			ps := pitches(tl.Key, voicing, led, h.Bass, hit.Voice)
			if hit.Voice == VoiceChord {
				ps = chord // with the style's tensions (BL-125)
			}
			if hit.Voice == VoiceOctave && len(adds) > 0 && arpeggiated(acc) {
				ps = []music.Pitch{above(voicing[0].Transpose(12), adds[0])}
			}
			for _, p := range slashed(acc, h, ps, hit.Voice) {
				n.Pitch = p
				notes = append(notes, n)
			}
		}
	}
	for k, p := range walkLine(tl.Harmony, i, len(walks)) {
		walks[k].Pitch = p
	}
	return append(notes, walks...)
}

// arpeggiated reports a pattern that spells its chords out with an octave
// and never strikes them whole: the arpeggios, where a style's tension
// takes the octave's place (#307).
func arpeggiated(acc Accompaniment) bool {
	return !slices.ContainsFunc(acc.Hits, func(h Hit) bool { return h.Voice == VoiceChord }) &&
		slices.ContainsFunc(acc.Hits, func(h Hit) bool { return h.Voice == VoiceOctave })
}

// above is the lowest pitch of class pc at or above p.
func above(p music.Pitch, pc music.PitchClass) music.Pitch {
	for p.Class() != pc {
		p++
	}
	return p
}

// slashed makes a slash chord's bass sound in a pattern with no bass
// voice of its own (BL-113): an arpeggio starts on it in place of the
// root, and a pattern with neither (block, strum, drone) plays it under
// the chord, as the plain block chords do.
func slashed(acc Accompaniment, h music.HarmonyEvent, ps []music.Pitch, v Voice) []music.Pitch {
	if h.Bass == nil || slices.ContainsFunc(acc.Hits, func(x Hit) bool { return x.Voice == VoiceBass || x.Voice == VoiceWalk }) {
		return ps
	}
	bass := music.NewPitch(*h.Bass, music.ChordOctave-1)
	hasRoot := slices.ContainsFunc(acc.Hits, func(x Hit) bool { return x.Voice == VoiceRoot })
	switch {
	case v == VoiceRoot:
		return []music.Pitch{bass}
	case v == VoiceChord && !hasRoot:
		return append([]music.Pitch{bass}, ps...)
	}
	return ps
}

// holdDrone ties the drone across chord changes (BL-108): a drone note
// that starts where the same drone note ends is the same note held on, so
// the tonic keeps sounding instead of being struck on every chord.
func holdDrone(notes []music.NoteEvent, acc Accompaniment, key music.Key) []music.NoteEvent {
	if !slices.ContainsFunc(acc.Hits, func(h Hit) bool { return h.Voice == VoiceDrone }) {
		return notes
	}
	drone := dronePitches(key)
	sortNotes(notes)
	out := make([]music.NoteEvent, 0, len(notes))
	held := map[music.Pitch]int{} // pitch → index in out of the note being held
	for _, n := range notes {
		if slices.Contains(drone, n.Pitch) {
			if i, ok := held[n.Pitch]; ok && out[i].Start+out[i].Duration == n.Start {
				out[i].Duration += n.Duration
				continue
			}
			held[n.Pitch] = len(out)
		}
		out = append(out, n)
	}
	return out
}

func chordsOf(harmony []music.HarmonyEvent) []music.Chord {
	out := make([]music.Chord, len(harmony))
	for i, h := range harmony {
		out[i] = h.Chord
	}
	return out
}

// drumTrack repeats the percussion figure for the length of the timeline.
func drumTrack(tl music.Timeline, dr Drums) music.Track {
	track := music.Track{Name: "drums", Channel: DrumChannel, Program: 0}
	length := tl.LengthTicks()
	bar := music.BarTicks(tl.TimeSignature, tl.PPQ)
	k := kitOf(dr)
	fillBar := fillBarHits(dr, k, tl.TimeSignature)
	for b, barStart := 0, music.Tick(0); barStart < length; b, barStart = b+1, barStart+bar {
		hits := dr.Hits
		if phraseEnd(b) && fills(dr) { // the figure up to the fill, then the fill (BL-126)
			hits = fillBar
			if k != kitHand && barStart+bar < length {
				track.Notes = append(track.Notes, music.NoteEvent{Pitch: Crash, Velocity: 96, Start: barStart + bar, Duration: eighth})
			}
		}
		for _, hit := range hits {
			if barStart+hit.Start >= length {
				continue
			}
			track.Notes = append(track.Notes, music.NoteEvent{Pitch: hit.Note, Velocity: hit.Velocity, Start: barStart + hit.Start, Duration: hit.Duration})
		}
	}
	sortNotes(track.Notes)
	return track
}

// droneLow is the lowest tonic of the drone (F2).
const droneLow music.Pitch = 41

// dronePitches are the key's tonic in F2..E3 and its fifth, which stays
// under the chords (C4 up).
func dronePitches(key music.Key) []music.Pitch {
	tonic := droneLow + music.Pitch((int(key.Tonic)-int(droneLow.Class())+12)%12)
	return []music.Pitch{tonic, tonic.Transpose(7)}
}

// pitches resolves a voice against a close voicing (root first).
func pitches(key music.Key, voicing, led []music.Pitch, bass *music.PitchClass, v Voice) []music.Pitch {
	root := voicing[0]
	at := func(i int) music.Pitch {
		if i < len(voicing) {
			return voicing[i]
		}
		return voicing[len(voicing)-1]
	}
	switch v {
	case VoiceChord:
		return led // voiced from the chord before (BL-086)
	case VoiceBass:
		if bass != nil { // a slash chord's own bass (ADR-0020)
			return []music.Pitch{music.NewPitch(*bass, music.ChordOctave-1)}
		}
		return []music.Pitch{root.Transpose(-12)}
	case VoiceRoot:
		return []music.Pitch{root}
	case VoiceThird:
		return []music.Pitch{at(1)}
	case VoiceFifth:
		return []music.Pitch{at(2)}
	case VoiceTop:
		return []music.Pitch{led[len(led)-1]}
	case VoiceOctave:
		return []music.Pitch{root.Transpose(12)}
	case VoiceDrone:
		return dronePitches(key)
	default:
		return nil
	}
}

func sortNotes(notes []music.NoteEvent) {
	sort.SliceStable(notes, func(i, j int) bool {
		if notes[i].Start != notes[j].Start {
			return notes[i].Start < notes[j].Start
		}
		return notes[i].Pitch < notes[j].Pitch
	})
}
