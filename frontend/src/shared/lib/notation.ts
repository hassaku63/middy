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
  ModalRole,
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
  sus2: 'sus2',
};

/**
 * Tonic names by key signature, mirroring music.Key.TonicName: the spelling
 * with fewer accidentals, ties and Db/C# going to flats; minor keys take
 * their parent major's signature (Eb minor from Gb major, G# minor from B).
 */
const TONICS: Record<Key['mode'], readonly string[]> = {
  major: ['C', 'Db', 'D', 'Eb', 'E', 'F', 'Gb', 'G', 'Ab', 'A', 'Bb', 'B'],
  minor: ['C', 'C#', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'G#', 'A', 'Bb', 'B'],
  // Dorian takes the signature of the major a tone below, mixolydian of the major a fifth
  // below, phrygian of the major a major third below.
  dorian: ['C', 'C#', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'Ab', 'A', 'Bb', 'B'],
  mixolydian: ['C', 'Db', 'D', 'Eb', 'E', 'F', 'F#', 'G', 'Ab', 'A', 'Bb', 'B'],
  phrygian: ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'Bb', 'B'],
};

/** "Bb major", "C# minor", "D dorian": the key as notation spells it. */
export function keyName(key: Key): string {
  return `${TONICS[key.mode][((key.tonic % 12) + 12) % 12] ?? 'C'} ${key.mode}`;
}

/**
 * The engine's key-spelled symbol when present (older saved timelines fall
 * back to sharps), with the tensions that sound in brackets before any
 * slash bass: Fmaj7(9), C7(9,13), F(add9), Cmaj7(9)/E (BL-307).
 */
export function harmonySymbol(h: Pick<HarmonyEvent, 'chord' | 'symbol' | 'tensions'>): string {
  const symbol = h.symbol ?? chordSymbol(h.chord);
  if (!h.tensions?.length) return symbol;
  const colour = `(${h.tensions.join(',')})`;
  const slash = symbol.indexOf('/');
  return slash < 0 ? symbol + colour : symbol.slice(0, slash) + colour + symbol.slice(slash);
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

/**
 * Functions (T, S, D) are a theory of major and minor: a mode has no
 * leading tone to pull back to the tonic, so the screen does not label
 * its chords by function (ADR-0045). The engine still assigns them by
 * degree (ADR-0015).
 */
export function isFunctional(mode: Key['mode']): boolean {
  return mode === 'major' || mode === 'minor';
}

/**
 * The Japanese name of a chord's role in a mode (#366, ADR-0047), where
 * functions do not hold. "other" has no name: it is shown by its degree
 * alone.
 */
export const ROLE_NAME: Record<Exclude<ModalRole, 'other'>, string> = {
  tonic: '主和音',
  characteristic: '特徴和音',
  tritone: '三全音の和音',
};

/** The role's name, or '' for a chord with no named role. */
export function roleName(role: ModalRole | undefined): string {
  return role && role !== 'other' ? ROLE_NAME[role] : '';
}

/** The Japanese name of a harmonic function, as the explanations say it. */
export const FUNCTION_NAME: Record<HarmonicFunction, string> = {
  T: 'トニック',
  S: 'サブドミナント',
  D: 'ドミナント',
};
