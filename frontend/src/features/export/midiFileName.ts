import type { Timeline } from '@/shared/api/client';
import { pitchClassName } from '@/shared/lib/notation';

/** e.g. "middy-C-major-I-V-vi-IV.mid"; sharps become "s" so the name is filesystem-safe. */
export function midiFileName(tl: Timeline): string {
  const key = `${pitchClassName(tl.key.tonic).replace('#', 's')}-${tl.key.mode}`;
  const numerals = tl.harmony
    .map((h) => h.romanNumeral.replace('°', 'dim').replace('ø', 'hd'))
    .filter((n) => /^[A-Za-z0-9]+$/.test(n))
    .join('-');
  return `middy-${key}${numerals ? `-${numerals}` : ''}.mid`;
}
