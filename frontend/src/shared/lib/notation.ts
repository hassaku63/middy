/**
 * Display names for pitch classes and chords. Mirrors the backend's
 * PitchClass.String and ChordQuality.Symbol (backend/internal/music) so
 * that the UI prints exactly what the explanation text refers to.
 */
import type {
  Chord,
  ChordQuality,
  HarmonicFunction,
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
  min6: 'm6',
  dim7: 'dim7',
  sus4: 'sus4',
  dom7sus4: '7sus4',
  power: '5',
};

/**
 * Tonic names by key signature, mirroring music.Key.TonicName: the spelling
 * with fewer accidentals, ties and Db/C# going to flats; minor keys take
 * their parent major's signature (Eb minor from Gb major, G# minor from B).
 */
const TONICS: Record<Key['mode'], readonly string[]> = {
  major: ['C', 'Db', 'D', 'Eb', 'E', 'F', 'Gb', 'G', 'Ab', 'A', 'Bb', 'B'],
  minor: ['C', 'C#', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'G#', 'A', 'Bb', 'B'],
  // Dorian takes the signature of the major a tone below, mixolydian of the major a fifth below.
  dorian: ['C', 'C#', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'Ab', 'A', 'Bb', 'B'],
  mixolydian: ['C', 'Db', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'Ab', 'A', 'Bb', 'B'],
};

/** "Bb major", "C# minor", "D dorian": the key as notation spells it. */
export function keyName(key: Key): string {
  return `${TONICS[key.mode][((key.tonic % 12) + 12) % 12] ?? 'C'} ${key.mode}`;
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

/** The Japanese name of a harmonic function, as the explanations say it. */
export const FUNCTION_NAME: Record<HarmonicFunction, string> = {
  T: 'トニック',
  S: 'サブドミナント',
  D: 'ドミナント',
};
