/**
 * Display names for pitch classes and chords. Mirrors the backend's
 * PitchClass.String and ChordQuality.Symbol (backend/internal/music) so
 * that the UI prints exactly what the explanation text refers to.
 */
import type { Chord, ChordQuality, Pitch, PitchClass } from '@/shared/api/client';

const SHARP_NAMES = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B'] as const;

const QUALITY_SUFFIX: Record<ChordQuality, string> = {
  maj: '',
  min: 'm',
  dim: 'dim',
  aug: 'aug',
  maj7: 'maj7',
  min7: 'm7',
  dom7: '7',
  m7b5: 'm7b5',
};

export function pitchClassName(pc: PitchClass): string {
  return SHARP_NAMES[((pc % 12) + 12) % 12] ?? 'C';
}

/** Scientific pitch notation, C4 = 60. */
export function pitchName(pitch: Pitch): string {
  return `${pitchClassName(pitch)}${Math.floor(pitch / 12) - 1}`;
}

export function chordSymbol(chord: Chord): string {
  return `${pitchClassName(chord.root)}${QUALITY_SUFFIX[chord.quality]}`;
}
