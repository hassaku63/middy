package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hassaku63/middy/backend/internal/music"
	"github.com/hassaku63/middy/backend/internal/music/pattern"
)

// maxRequestBody bounds JSON request bodies. Progressions are tiny; this
// is a safety limit, not a product limit.
const maxRequestBody = 64 << 10

// defaultBeatsPerChord is used when a progression step omits "beats".
const defaultBeatsPerChord = 4

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
	// Feel is "straight" (default) or "swing"; it needs a pattern to act on.
	Feel string `json:"feel,omitempty"`
}

type progressionLabel struct {
	Numeral string `json:"numeral"`
	Beats   *int   `json:"beats,omitempty"`
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

	prog, err := toProgression(req.Key, req.Progression)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	tl, err := music.FromProgression(req.Key, prog, req.TempoBPM)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	tl, err = pattern.Render(tl, pattern.Choice{Accompaniment: req.Accompaniment, Drums: req.Drums, Feel: pattern.Feel(req.Feel)})
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	tl, err = music.Analyze(tl)
	if err != nil {
		// FromProgression only emits diatonic chords, so this is a bug, not user error.
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, timelineResponse{Timeline: tl})
}

// patternsResponse is the catalogue for GET /patterns.
type patternsResponse struct {
	Accompaniments []pattern.Accompaniment `json:"accompaniments"`
	Drums          []pattern.Drums         `json:"drums"`
	Feels          []pattern.FeelInfo      `json:"feels"`
}

func handlePatterns(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, patternsResponse{
		Accompaniments: pattern.Accompaniments(),
		Drums:          pattern.DrumPatterns(),
		Feels:          pattern.Feels(),
	})
}

// toProgression converts request labels into domain steps, reporting
// every bad label at once. The key decides what an uppercase V means in
// minor (music.StepFromLabel).
func toProgression(key music.Key, labels []progressionLabel) (music.Progression, error) {
	if len(labels) == 0 {
		return nil, errors.New("progression must contain at least one chord")
	}
	var errs []error
	prog := make(music.Progression, 0, len(labels))
	for i, l := range labels {
		beats := defaultBeatsPerChord
		if l.Beats != nil {
			beats = *l.Beats
		}
		step, err := music.StepFromLabel(key, l.Numeral, beats)
		if err != nil {
			errs = append(errs, fmt.Errorf("progression[%d]: %w", i, err))
			continue
		}
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
	dec := json.NewDecoder(io.LimitReader(r.Body, maxRequestBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if dec.More() {
		return errors.New("invalid JSON body: trailing data")
	}
	return nil
}
