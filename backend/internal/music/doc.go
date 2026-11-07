// Package music is middy's deterministic music-theory core.
//
// It has no dependencies beyond the standard library and no I/O. Everything
// here is a pure function over small value types so that it can be tested
// exhaustively with tables, and reasoned about by humans and AI agents alike.
//
// Conventions:
//   - Pitches are MIDI note numbers (60 = C4, middle C).
//   - Pitch classes are 0..11 with C = 0.
//   - Time is measured in ticks; DefaultPPQ ticks make one quarter note.
//   - Scale degrees are 1-based (1 = tonic), matching music-theory usage.
//
// See docs/architecture/domain-model.md for the vocabulary and the roadmap
// for what is intentionally not here yet (harmonic function analysis,
// pattern library, SMF export, ...).
package music
