/**
 * Display names for pitch classes and chords. Mirrors the backend's
 * PitchClass.String and ChordQuality.Symbol (backend/internal/music) so
 * that the UI prints exactly what the explanation text refers to.
 */
import type {
  Chord,
  ChordQuality,
  HarmonyEvent,
  Key,
  Pitch,
  PitchClass,
} from '@/shared/api/client';

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

/**
 * Tonic names by key signature, mirroring music.Key.TonicName: the spelling
 * with fewer accidentals, ties and Db/C# going to flats; minor keys take
 * their parent major's signature (Eb minor from Gb major, G# minor from B).
 */
const MAJOR_TONICS = ['C', 'Db', 'D', 'Eb', 'E', 'F', 'Gb', 'G', 'Ab', 'A', 'Bb', 'B'] as const;
const MINOR_TONICS = ['C', 'C#', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'G#', 'A', 'Bb', 'B'] as const;

/** "Bb major", "C# minor": the key as notation spells it. */
export function keyName(key: Key): string {
  const table = key.mode === 'minor' ? MINOR_TONICS : MAJOR_TONICS;
  return `${table[((key.tonic % 12) + 12) % 12] ?? 'C'} ${key.mode}`;
}

/** The engine's key-spelled symbol when present; older saved timelines fall back to sharps. */
export function harmonySymbol(h: Pick<HarmonyEvent, 'chord' | 'symbol'>): string {
  return h.symbol ?? chordSymbol(h.chord);
}

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
