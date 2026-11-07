// Package pattern turns a harmony-only timeline into something that
// grooves: an accompaniment template (how the chord tones are played
// across a bar) and a drum template (GM percussion on channel 10). Both
// are one-bar figures in 4/4 at music.DefaultPPQ, repeated over every
// harmony event; a chord shorter than a bar simply gets the first part.
//
// Templates are data so that the explanation layer (backlog BL-022) can
// describe them and the LLM (BL-023) can pick them by id.
package pattern

import (
	"errors"
	"fmt"
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
)

// Hit is one event of an accompaniment template, positioned within a bar.
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
	Hits        []Hit  `json:"hits"`
}

// DrumHit is one percussion event within a bar. Note uses the GM drum map.
type DrumHit struct {
	Start    music.Tick  `json:"start"`
	Duration music.Tick  `json:"duration"`
	Note     music.Pitch `json:"note"`
	Velocity int         `json:"velocity"`
}

// Drums is a one-bar percussion figure.
type Drums struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Hits        []DrumHit `json:"hits"`
}

// Durations in ticks for 4/4 at DefaultPPQ.
const (
	bar     = music.Tick(4 * music.DefaultPPQ)
	quarter = music.Tick(music.DefaultPPQ)
	eighth  = music.Tick(music.DefaultPPQ / 2)
)

// GM drum map notes used by the templates.
const (
	Kick      music.Pitch = 36
	Snare     music.Pitch = 38
	ClosedHat music.Pitch = 42
	OpenHat   music.Pitch = 46
	RideBell  music.Pitch = 53
)

// DrumChannel is the General MIDI percussion channel (0-based).
const DrumChannel = 9

// ErrUnknownPattern is returned by the lookup functions.
var ErrUnknownPattern = errors.New("pattern: unknown pattern id")

var accompaniments = []Accompaniment{
	{
		ID: "block", Name: "ブロック", Program: 0,
		Description: "コードの構成音をまとめて、コードが変わるまで伸ばす。進行そのものを聴くための形。",
		Hits:        []Hit{{0, bar, VoiceChord, 80}},
	},
	{
		ID: "pulse", Name: "4 分刻み", Program: 0,
		Description: "毎拍にコードを打つ。拍の位置がはっきりし、テンポ感が出る。",
		Hits: []Hit{
			{0, quarter, VoiceChord, 84}, {quarter, quarter, VoiceChord, 72},
			{2 * quarter, quarter, VoiceChord, 80}, {3 * quarter, quarter, VoiceChord, 72},
		},
	},
	{
		ID: "ballad", Name: "バラード", Program: 0,
		Description: "1 拍目と 3 拍目にベース、2 拍目と 4 拍目にコード。重心が交互に動き、ゆったり揺れる。",
		Hits: []Hit{
			{0, quarter, VoiceBass, 88}, {quarter, quarter, VoiceChord, 70},
			{2 * quarter, quarter, VoiceBass, 80}, {3 * quarter, quarter, VoiceChord, 70},
		},
	},
	{
		ID: "arpeggio", Name: "アルペジオ", Program: 0,
		Description: "構成音を 8 分音符で下から順に弾く。和音が線になって流れる。",
		Hits: []Hit{
			{0, eighth, VoiceRoot, 82}, {eighth, eighth, VoiceThird, 70},
			{2 * eighth, eighth, VoiceFifth, 74}, {3 * eighth, eighth, VoiceOctave, 70},
			{4 * eighth, eighth, VoiceRoot, 78}, {5 * eighth, eighth, VoiceThird, 68},
			{6 * eighth, eighth, VoiceFifth, 72}, {7 * eighth, eighth, VoiceOctave, 68},
		},
	},
	{
		ID: "strum", Name: "ストローク", Program: 24,
		Description: "ギターのストロークのように、表拍に長く、裏拍に短く刻む。8 ビートの基本。",
		Hits: []Hit{
			{0, quarter, VoiceChord, 86}, {quarter + eighth, eighth, VoiceChord, 66},
			{2 * quarter, quarter, VoiceChord, 80}, {3 * quarter, eighth, VoiceChord, 70},
			{3*quarter + eighth, eighth, VoiceChord, 64},
		},
	},
}

var drums = []Drums{
	{
		ID: "eight-beat", Name: "8 ビート",
		Description: "キックが 1・3 拍、スネアが 2・4 拍、ハイハットが 8 分で刻む。ポップスの土台。",
		Hits: append(hats(eighth, 8, 60),
			DrumHit{0, eighth, Kick, 100}, DrumHit{quarter, eighth, Snare, 96},
			DrumHit{2 * quarter, eighth, Kick, 96}, DrumHit{3 * quarter, eighth, Snare, 96},
		),
	},
	{
		ID: "ballad", Name: "バラード",
		Description: "キックが 1 拍、スネアが 3 拍だけ。間が広く、ゆっくり揺れる。",
		Hits: append(hats(quarter, 4, 48),
			DrumHit{0, eighth, Kick, 96}, DrumHit{2 * quarter, eighth, Snare, 90},
		),
	},
	{
		ID: "bossa", Name: "ボサノバ風",
		Description: "キックが 1 拍と 2 拍裏、3 拍と 4 拍裏。表と裏が交互に来て、前に流れる。",
		Hits: append(hats(eighth, 8, 56),
			DrumHit{0, eighth, Kick, 96}, DrumHit{quarter + eighth, eighth, Kick, 84},
			DrumHit{2 * quarter, eighth, Kick, 92}, DrumHit{3*quarter + eighth, eighth, Kick, 84},
			DrumHit{quarter, eighth, RideBell, 70}, DrumHit{3 * quarter, eighth, RideBell, 70},
		),
	},
}

func hats(step music.Tick, count, velocity int) []DrumHit {
	out := make([]DrumHit, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, DrumHit{music.Tick(i) * step, step / 2, ClosedHat, velocity})
	}
	return out
}

// Accompaniments lists the catalogue in a stable order.
func Accompaniments() []Accompaniment {
	out := make([]Accompaniment, len(accompaniments))
	copy(out, accompaniments)
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
// input is not modified. Only 4/4 at DefaultPPQ is supported today.
func Apply(tl music.Timeline, acc Accompaniment, dr *Drums) (music.Timeline, error) {
	if tl.TimeSignature != (music.TimeSignature{Numerator: 4, Denominator: 4}) || tl.PPQ != music.DefaultPPQ {
		return music.Timeline{}, fmt.Errorf("pattern: templates are defined for 4/4 at %d ppq, got %d/%d at %d",
			music.DefaultPPQ, tl.TimeSignature.Numerator, tl.TimeSignature.Denominator, tl.PPQ)
	}
	if len(tl.Harmony) == 0 {
		return music.Timeline{}, errors.New("pattern: timeline has no harmony to play")
	}

	out := tl
	out.Tracks = []music.Track{chordTrack(tl, acc)}
	if dr != nil {
		out.Tracks = append(out.Tracks, drumTrack(tl, *dr))
	}
	if err := out.Validate(); err != nil {
		return music.Timeline{}, fmt.Errorf("pattern: %w", err)
	}
	return out, nil
}

// chordTrack realises the accompaniment over every harmony event.
func chordTrack(tl music.Timeline, acc Accompaniment) music.Track {
	track := music.Track{Name: music.ChordTrackName, Channel: 0, Program: acc.Program}
	for _, h := range tl.Harmony {
		voicing := h.Chord.Voicing(music.ChordOctave)
		end := h.Start + h.Duration
		for barStart := h.Start; barStart < end; barStart += bar {
			for _, hit := range acc.Hits {
				start := barStart + hit.Start
				if start >= end {
					continue
				}
				dur := min(hit.Duration, end-start)
				for _, p := range pitches(voicing, hit.Voice) {
					track.Notes = append(track.Notes, music.NoteEvent{Pitch: p, Velocity: hit.Velocity, Start: start, Duration: dur})
				}
			}
		}
	}
	sortNotes(track.Notes)
	return track
}

// drumTrack repeats the percussion figure for the length of the timeline.
func drumTrack(tl music.Timeline, dr Drums) music.Track {
	track := music.Track{Name: "drums", Channel: DrumChannel, Program: 0}
	length := tl.LengthTicks()
	for barStart := music.Tick(0); barStart < length; barStart += bar {
		for _, hit := range dr.Hits {
			if barStart+hit.Start >= length {
				continue
			}
			track.Notes = append(track.Notes, music.NoteEvent{Pitch: hit.Note, Velocity: hit.Velocity, Start: barStart + hit.Start, Duration: hit.Duration})
		}
	}
	sortNotes(track.Notes)
	return track
}

// pitches resolves a voice against a close voicing (root first).
func pitches(voicing []music.Pitch, v Voice) []music.Pitch {
	root := voicing[0]
	at := func(i int) music.Pitch {
		if i < len(voicing) {
			return voicing[i]
		}
		return voicing[len(voicing)-1]
	}
	switch v {
	case VoiceChord:
		return voicing
	case VoiceBass:
		return []music.Pitch{root.Transpose(-12)}
	case VoiceRoot:
		return []music.Pitch{root}
	case VoiceThird:
		return []music.Pitch{at(1)}
	case VoiceFifth:
		return []music.Pitch{at(2)}
	case VoiceTop:
		return []music.Pitch{voicing[len(voicing)-1]}
	case VoiceOctave:
		return []music.Pitch{root.Transpose(12)}
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
