package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

// maxRequestBody bounds JSON request bodies. Progressions are tiny; this
// is a safety limit, not a product limit.
const maxRequestBody = 64 << 10

// Request/response shapes for POST /api/v1/timelines/from-progression.
// They are the contract in docs/api/openapi.yaml; change both together.
type fromProgressionRequest struct {
	Key         music.Key          `json:"key"`
	TempoBPM    float64            `json:"tempoBpm"`
	Progression []progressionLabel `json:"progression"`
	// Accompaniment and Drums are pattern ids (GET /patterns). Empty means
	// block chords and no drums, exactly what FromProgression produces.
	Accompaniment string `json:"accompaniment,omitempty"`
	Drums         string `json:"drums,omitempty"`
	// DrumGrid writes the drums out instead of naming them (BL-123).
	DrumGrid *music.DrumGrid `json:"drumGrid,omitempty"`
	// Feel is "straight" (default) or "swing"; it needs a pattern to act on.
	Feel string `json:"feel,omitempty"`
	// Style may choose the chords' instrument (BL-110); the key change sends
	// back the style the example was rendered in.
	Style string `json:"style,omitempty"`
	// TimeSignature is 4/4 (default), 3/4 or 6/8; beats count in its beats (ADR-0017).
	TimeSignature *music.TimeSignature `json:"timeSignature,omitempty"`
}

type progressionLabel struct {
	Numeral string     `json:"numeral"`
	Beats   *int       `json:"beats,omitempty"`
	Push    music.Push `json:"push,omitempty"` // BL-085
}

type timelineResponse struct {
	Timeline music.Timeline `json:"timeline"`
}

// handleFromProgression renders a roman-numeral progression in a key and
// returns the analysed timeline. It is the whole M1 pipeline in one call:
// parse -> FromProgression -> Analyze.
func handleFromProgression(w http.ResponseWriter, r *http.Request) {
	var req fromProgressionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	ts := music.FourFour
	if req.TimeSignature != nil {
		ts = *req.TimeSignature
	}
	if err := music.SupportedMeter(ts); err != nil { // before the default lengths, which count its beats
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	prog, err := toProgression(req.Key, req.Progression, ts.BeatsPerBar())
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	tl, err := music.FromProgressionIn(ts, req.Key, prog, req.TempoBPM)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	tl, err = pattern.Render(tl, pattern.Choice{Accompaniment: req.Accompaniment, Drums: req.Drums, Grid: req.DrumGrid, Feel: pattern.Feel(strings.ToLower(req.Feel)), Style: req.Style})
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	tl, err = music.Analyze(tl)
	if err != nil {
		// Every chord the request may name is one Analyze can place, so
		// this is a bug, not user error: log it, tell the client little.
		slog.ErrorContext(r.Context(), "timelines: could not analyse the timeline", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "the timeline could not be analysed")
		return
	}
	writeJSON(w, http.StatusOK, timelineResponse{Timeline: tl})
}

// patternsResponse is the catalogue for GET /patterns.
type patternsResponse struct {
	Accompaniments []pattern.Accompaniment `json:"accompaniments"`
	Drums          []pattern.Drums         `json:"drums"`
	Feels          []pattern.FeelInfo      `json:"feels"`
	Styles         []pattern.Style         `json:"styles"`
}

func handlePatterns(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, patternsResponse{
		Accompaniments: pattern.Accompaniments(),
		Drums:          pattern.DrumPatterns(),
		Feels:          pattern.Feels(),
		Styles:         pattern.Styles(),
	})
}

// toProgression converts request labels into domain steps, reporting
// every bad label at once. The key decides what an uppercase V means in
// minor (music.StepFromLabel).
// toProgression reads the labels; a label without beats lasts one bar of
// barBeats beats.
func toProgression(key music.Key, labels []progressionLabel, barBeats int) (music.Progression, error) {
	if len(labels) == 0 {
		return nil, errors.New("progression must contain at least one chord")
	}
	var errs []error
	prog := make(music.Progression, 0, len(labels))
	for i, l := range labels {
		beats := barBeats
		if l.Beats != nil {
			beats = *l.Beats
		}
		// A label the key would change is refused, as for the AI (ADR-0013):
		// "Vmaj7" in C major names Gmaj7 and would otherwise sound as G7.
		if _, err := music.CanonicalLabel(key, l.Numeral); err != nil {
			errs = append(errs, fmt.Errorf("progression[%d]: %w", i, err))
			continue
		}
		step, err := music.StepFromLabel(key, l.Numeral, beats)
		if err != nil {
			errs = append(errs, fmt.Errorf("progression[%d]: %w", i, err))
			continue
		}
		step.Push = l.Push
		prog = append(prog, step)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return prog, nil
}

// decodeJSON reads a bounded JSON body and rejects unknown fields so that
// typos in field names surface as 400s instead of silently defaulting.
func decodeJSON(r *http.Request, v any) error {
	return decodeJSONUpTo(r, v, maxRequestBody)
}

// decodeJSONUpTo is decodeJSON with its own limit, for bodies that carry a
// whole timeline (a MIDI export). A body over the limit is an error that
// says so, not a silently cut one.
func decodeJSONUpTo(r *http.Request, v any, limit int64) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, limit+1))
	if err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if int64(len(body)) > limit {
		return fmt.Errorf("invalid JSON body: larger than %d bytes", limit)
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	// More() is false before a stray "}" or "]", so read once more: only the
	// end of the body may follow the value.
	if err := dec.Decode(&json.RawMessage{}); !errors.Is(err, io.EOF) {
		return errors.New("invalid JSON body: trailing data")
	}
	return nil
}
