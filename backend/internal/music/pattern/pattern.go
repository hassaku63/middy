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

var accompaniments = []Accompaniment{
	{
		ID: "block", Name: "ブロック", Program: 0,
		Description: "コードの構成音をまとめて、コードが変わるまで伸ばす。進行そのものを聴くための形。",
		Hits:        []Hit{{0, 0, VoiceChord, 80}}, // any meter: to the end of the bar
	},
	{
		ID: "pulse", Meter: &fourFour, Name: "4 分刻み", Program: 0,
		Description: "毎拍にコードを打つ。拍の位置がはっきりし、テンポ感が出る。",
		Hits: []Hit{
			{0, quarter, VoiceChord, 84}, {quarter, quarter, VoiceChord, 72},
			{2 * quarter, quarter, VoiceChord, 80}, {3 * quarter, quarter, VoiceChord, 72},
		},
	},
	{
		ID: "ballad", Meter: &fourFour, Name: "バラード", Program: 0,
		Description: "1 拍目と 3 拍目にベース、2 拍目と 4 拍目にコード。重心が交互に動き、ゆったり揺れる。",
		Hits: []Hit{
			{0, quarter, VoiceBass, 88}, {quarter, quarter, VoiceChord, 70},
			{2 * quarter, quarter, VoiceBass, 80}, {3 * quarter, quarter, VoiceChord, 70},
		},
	},
	{
		ID: "arpeggio", Meter: &fourFour, Name: "アルペジオ", Program: 0,
		Description: "構成音を 8 分音符で下から順に弾く。和音が線になって流れる。",
		Hits: []Hit{
			{0, eighth, VoiceRoot, 82}, {eighth, eighth, VoiceThird, 70},
			{2 * eighth, eighth, VoiceFifth, 74}, {3 * eighth, eighth, VoiceOctave, 70},
			{4 * eighth, eighth, VoiceRoot, 78}, {5 * eighth, eighth, VoiceThird, 68},
			{6 * eighth, eighth, VoiceFifth, 72}, {7 * eighth, eighth, VoiceOctave, 68},
		},
	},
	{
		ID: "strum", Meter: &fourFour, Name: "ストローク", Program: 24,
		Description: "ギターのストロークのように、表拍に長く、裏拍に短く刻む。8 ビートの基本。",
		Hits: []Hit{
			{0, quarter, VoiceChord, 86}, {quarter + eighth, eighth, VoiceChord, 66},
			{2 * quarter, quarter, VoiceChord, 80}, {3 * quarter, eighth, VoiceChord, 70},
			{3*quarter + eighth, eighth, VoiceChord, 64},
		},
	},
	{
		ID: "reel", Meter: &fourFour, Name: "リールの裏打ち", Program: 25,
		Description: "1・3 拍にベース、各拍の裏に短い和音。ギターがリールの細かな 8 分を後ろから押し、ケルトの舞曲の推進力を作る。",
		Hits: []Hit{
			{0, quarter, VoiceBass, 86}, {eighth, eighth * 3 / 4, VoiceChord, 64},
			{quarter + eighth, eighth * 3 / 4, VoiceChord, 60},
			{2 * quarter, quarter, VoiceBass, 80}, {2*quarter + eighth, eighth * 3 / 4, VoiceChord, 62},
			{3*quarter + eighth, eighth * 3 / 4, VoiceChord, 58},
		},
	},
	{
		ID: "comp", Meter: &fourFour, Name: "4 ビートのコンピング", Program: 0,
		Description: "1・3 拍にベース、1 拍目と 2 拍目の裏に短い和音（チャールストンのリズム）。和音を短く切って間を空け、ジャズのピアノが伴奏する形。ノリはスウィングと組み合わせる。",
		Hits: []Hit{
			{0, quarter, VoiceBass, 84}, {0, eighth * 3 / 4, VoiceChord, 70},
			{quarter + eighth, eighth * 3 / 4, VoiceChord, 64},
			{2 * quarter, quarter, VoiceBass, 76},
		},
	},
	{
		ID: "walking", Meter: &fourFour, Name: "ウォーキング・ベース", Program: 0,
		Description: "ベースが 4 分音符で歩く。和音の 1 拍目に根音、間に和音の音、和音の最後の拍で次の和音の根音の半音下から入る。上では短い和音がチャールストンのリズムで鳴る。ジャズの 4 ビート。ノリはスウィングと組み合わせる。",
		Hits: []Hit{
			{0, quarter * 9 / 10, VoiceWalk, 84}, {quarter, quarter * 9 / 10, VoiceWalk, 70},
			{2 * quarter, quarter * 9 / 10, VoiceWalk, 78}, {3 * quarter, quarter * 9 / 10, VoiceWalk, 70},
			{0, eighth * 3 / 4, VoiceChord, 64}, {quarter + eighth, eighth * 3 / 4, VoiceChord, 58},
		},
	},
	{
		ID: "drone", Name: "ドローン", Program: 20,
		Description: "主音と 5 度を低く保ち、その上で和音を伸ばす（リードオルガン）。和音が動いても主音が鳴り続けるので、バグパイプやフィドルの開放弦のような、ケルトや民謡の土台になる。",
		Hits:        []Hit{{0, 0, VoiceDrone, 70}, {0, 0, VoiceChord, 62}}, // any meter: to the end of the bar
	},
	{
		ID: "waltz", Meter: &threeFour, Name: "ワルツ", Program: 0,
		Description: "1 拍目にベース、2・3 拍目にコード。「ズン・チャッ・チャッ」と 1 小節ごとに回る、3 拍子の基本。",
		Hits: []Hit{
			{0, quarter, VoiceBass, 90}, {quarter, quarter, VoiceChord, 68}, {2 * quarter, quarter, VoiceChord, 64},
		},
	},
	{
		ID: "arpeggio-3", Meter: &threeFour, Name: "3 拍のアルペジオ", Program: 0,
		Description: "構成音を 8 分音符で上って下りる（根音・3 度・5 度・オクターブ・5 度・3 度）。3 拍子の揺れが線になって流れる。",
		Hits: []Hit{
			{0, eighth, VoiceRoot, 82}, {eighth, eighth, VoiceThird, 68}, {2 * eighth, eighth, VoiceFifth, 72},
			{3 * eighth, eighth, VoiceOctave, 70}, {4 * eighth, eighth, VoiceFifth, 70}, {5 * eighth, eighth, VoiceThird, 66},
		},
	},
	{
		ID: "rocking-6", Meter: &sixEight, Name: "6/8 の揺れ", Program: 0,
		Description: "付点 4 分の拍の頭にベース、残りの 8 分 2 つにコード。「ズン・チャ・チャ」が 1 小節に 2 回来て、ゆりかごのように揺れる。",
		Hits: []Hit{
			{0, 3 * eighth, VoiceBass, 88}, {eighth, eighth, VoiceChord, 62}, {2 * eighth, eighth, VoiceChord, 58},
			{3 * eighth, 3 * eighth, VoiceBass, 80}, {4 * eighth, eighth, VoiceChord, 60}, {5 * eighth, eighth, VoiceChord, 56},
		},
	},
	{
		ID: "arpeggio-6", Meter: &sixEight, Name: "6/8 のアルペジオ", Program: 0,
		Description: "構成音を 8 分音符 6 つで上って下りる（根音・3 度・5 度・オクターブ・5 度・3 度）。2 拍目の頭にオクターブが来て、1 小節で波が 1 往復する。",
		Hits: []Hit{
			{0, eighth, VoiceRoot, 82}, {eighth, eighth, VoiceThird, 64}, {2 * eighth, eighth, VoiceFifth, 66},
			{3 * eighth, eighth, VoiceOctave, 76}, {4 * eighth, eighth, VoiceFifth, 64}, {5 * eighth, eighth, VoiceThird, 62},
		},
	},
}

var drums = []Drums{
	{
		ID: "eight-beat", Meter: &fourFour, Name: "8 ビート",
		Description: "キックが 1・3 拍、スネアが 2・4 拍、ハイハットが 8 分で刻む。ポップスの土台。",
		Hits: append(hats(eighth, 8, 60),
			DrumHit{0, eighth, Kick, 100}, DrumHit{quarter, eighth, Snare, 96},
			DrumHit{2 * quarter, eighth, Kick, 96}, DrumHit{3 * quarter, eighth, Snare, 96},
		),
	},
	{
		ID: "ballad", Meter: &fourFour, Name: "バラード",
		Description: "キックが 1 拍、スネアが 3 拍だけ。間が広く、ゆっくり揺れる。",
		Hits: append(hats(quarter, 4, 48),
			DrumHit{0, eighth, Kick, 96}, DrumHit{2 * quarter, eighth, Snare, 90},
		),
	},
	{
		ID: "bossa", Meter: &fourFour, Name: "ボサノバ風",
		Description: "キックが 1 拍と 2 拍裏、3 拍と 4 拍裏。表と裏が交互に来て、前に流れる。",
		Hits: append(hats(eighth, 8, 56),
			DrumHit{0, eighth, Kick, 96}, DrumHit{quarter + eighth, eighth, Kick, 84},
			DrumHit{2 * quarter, eighth, Kick, 92}, DrumHit{3*quarter + eighth, eighth, Kick, 84},
			DrumHit{quarter, eighth, RideBell, 70}, DrumHit{3 * quarter, eighth, RideBell, 70},
		),
	},
	{
		ID: "jazz-ride", Meter: &fourFour, Name: "ジャズのライド",
		Description: "ライドシンバルが 1・2・2 拍の裏・3・4・4 拍の裏を刻み（チン・チキ・チン・チキ）、足のハイハットが 2・4 拍で閉じる。キックは 4 分でごく弱く支えるだけで、スネアの強い 2・4 拍はない。スウィングと組み合わせる、ジャズの 4 ビートの土台。",
		Hits: []DrumHit{
			{0, eighth, Ride, 72}, {quarter, eighth, Ride, 66}, {quarter + eighth, eighth, Ride, 54},
			{2 * quarter, eighth, Ride, 70}, {3 * quarter, eighth, Ride, 66}, {3*quarter + eighth, eighth, Ride, 54},
			{quarter, eighth, PedalHat, 58}, {3 * quarter, eighth, PedalHat, 58},
			{0, eighth, Kick, 34}, {quarter, eighth, Kick, 30}, {2 * quarter, eighth, Kick, 32}, {3 * quarter, eighth, Kick, 30},
		},
	},
	{
		ID: "jazz-waltz", Meter: &threeFour, Name: "ジャズ・ワルツのライド",
		Description: "ライドシンバルが 1・2・2 拍の裏・3 拍を刻み、足のハイハットが 2 拍で閉じる。キックは 1 拍をごく弱く支えるだけ。3 拍子のままスウィングさせる、ジャズ・ワルツの土台。",
		Hits: []DrumHit{
			{0, eighth, Ride, 72}, {quarter, eighth, Ride, 64}, {quarter + eighth, eighth, Ride, 54}, {2 * quarter, eighth, Ride, 66},
			{quarter, eighth, PedalHat, 56}, {0, eighth, Kick, 34},
		},
	},
	{
		ID: "bodhran", Meter: &fourFour, Name: "ボーラン風",
		Description: "低いタムが 1・3 拍、サイドスティックが 2・4 拍、中くらいのタムが残りの 8 分を軽く刻む。スネアもハイハットも使わず、アイルランドの片面太鼓（ボーラン）のように控えめに転がる。",
		Hits: []DrumHit{
			{0, eighth, LowTom, 90}, {eighth, eighth, MidTom, 48}, {quarter, eighth, MidTom, 60},
			{quarter, eighth, SideStick, 56}, {quarter + eighth, eighth, MidTom, 48},
			{2 * quarter, eighth, LowTom, 84}, {2*quarter + eighth, eighth, MidTom, 48},
			{3 * quarter, eighth, MidTom, 60}, {3 * quarter, eighth, SideStick, 52},
			{3*quarter + eighth, eighth, MidTom, 48},
		},
	},
	{
		ID: "waltz", Meter: &threeFour, Name: "ワルツ",
		Description: "キックが 1 拍目、スネアが 2・3 拍目を軽く打つ。1 拍目の重みで、3 拍ごとに回る。",
		Hits: append(hats(quarter, 3, 44),
			DrumHit{0, eighth, Kick, 96}, DrumHit{quarter, eighth, Snare, 60}, DrumHit{2 * quarter, eighth, Snare, 56},
		),
	},
	{
		ID: "six-eight", Meter: &sixEight, Name: "6/8 のバラード",
		Description: "キックが 1 拍目、スネアが 2 拍目（4 つめの 8 分）、ハイハットが 8 分を 6 つ刻む。大きく 2 つに揺れ、その中が 3 つに分かれて聞こえる。",
		Hits: append(hats(eighth, 6, 46),
			DrumHit{0, eighth, Kick, 96}, DrumHit{3 * eighth, eighth, Snare, 84},
		),
	},
	{
		ID: "jig", Meter: &sixEight, Name: "ジグ",
		Description: "キックが 2 つの拍の頭、軽いスネアが各拍の 3 つめの 8 分。「タタタ・タタタ」と回り、拍の終わりから次の拍へ跳ねる、ケルトの舞曲の推進力。",
		Hits: append(hats(eighth, 6, 50),
			DrumHit{0, eighth, Kick, 98}, DrumHit{2 * eighth, eighth, Snare, 58},
			DrumHit{3 * eighth, eighth, Kick, 90}, DrumHit{5 * eighth, eighth, Snare, 58},
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
	out.Tracks = []music.Track{chordTrack(tl, acc)}
	out.Rhythm = &music.Rhythm{Accompaniment: describeAccompaniment(acc, ts)}
	if dr != nil {
		out.Tracks = append(out.Tracks, drumTrack(tl, *dr))
		out.Rhythm.Drums = describeDrums(*dr, ts)
	}
	if err := out.Validate(); err != nil {
		return music.Timeline{}, fmt.Errorf("pattern: %w", err)
	}
	return out, nil
}

// chordTrack realises the accompaniment over every harmony event.
func chordTrack(tl music.Timeline, acc Accompaniment) music.Track {
	track := music.Track{Name: music.ChordTrackName, Channel: 0, Program: acc.Program}
	bar := music.BarTicks(tl.TimeSignature, tl.PPQ)
	led := music.LeadVoicings(chordsOf(tl.Harmony)) // BL-086
	for i := range tl.Harmony {
		track.Notes = append(track.Notes, chordNotes(tl, acc, i, led[i], bar)...)
	}
	track.Notes = music.PushNotes(tl.Harmony, track.Notes, tl.PPQ) // before the feel (BL-085)
	track.Notes = holdDrone(track.Notes, acc, tl.Key)              // after the pushes, which cut at each chord
	sortNotes(track.Notes)
	return track
}

// chordNotes realises the accompaniment over chord i, a bar of the
// template at a time from the chord's start; the led voicing plays the
// chord, and the walking bass is filled in once its hits are known.
func chordNotes(tl music.Timeline, acc Accompaniment, i int, led []music.Pitch, bar music.Tick) []music.NoteEvent {
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
			for _, p := range slashed(acc, h, pitches(tl.Key, voicing, led, h.Bass, hit.Voice), hit.Voice) {
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
