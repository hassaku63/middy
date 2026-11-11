import type { IdeaCandidate } from '@/shared/api/client';
import { harmonySymbol } from '@/shared/lib/notation';

/**
 * An example's numerals as the engine wrote them in its timeline, the same
 * as the chord lane shows (♭VII, not bVII; I, not I/1), so the card and the
 * drawing agree (#451). Falls back to the labels as written while there is
 * no harmony to read.
 */
export function numeralsOf(c: IdeaCandidate): string[] {
  return c.timeline.harmony.length > 0
    ? c.timeline.harmony.map((h) => h.romanNumeral)
    : c.progression;
}

/** The chords' names in the example's key (Bbmaj7, Bb/A, …), or none without harmony (#451). */
export function chordNamesOf(c: IdeaCandidate): string[] {
  return c.timeline.harmony.map(harmonySymbol);
}
