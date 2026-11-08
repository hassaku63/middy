package music

// Rhythm is the explanation data for how a timeline grooves: one entry
// per applied pattern (accompaniment, drums). It is filled by
// internal/music/pattern.Apply and consumed by the explanation panel;
// nothing in it sounds.
type Rhythm struct {
	Accompaniment *RhythmPart `json:"accompaniment,omitempty"`
	Drums         *RhythmPart `json:"drums,omitempty"`
	// Feel is present when the eighths are not straight (e.g. swing).
	Feel *RhythmFeel `json:"feel,omitempty"`
}

// RhythmFeel explains how the eighths are played.
type RhythmFeel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Explanation string `json:"explanation"`
}

// RhythmPart describes one bar of a pattern in beginner terms.
type RhythmPart struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Hits per bar, split by where they land: on a beat (1 2 3 4), on the
	// eighth between beats ("and"), or finer.
	Hits       int  `json:"hits"`
	OnBeat     int  `json:"onBeat"`
	OffBeat    int  `json:"offBeat"`
	Finer      int  `json:"finer"`
	Syncopated bool `json:"syncopated"`
	// Explanation is a short Japanese paragraph the UI shows as-is.
	Explanation string `json:"explanation"`
}
