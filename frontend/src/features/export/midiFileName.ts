import type { Timeline } from '@/shared/api/client';
import { keyName } from '@/shared/lib/notation';

/**
 * e.g. "middy-C-major-I-V-vi-IV.mid". Sharps become "s", "♭VII" becomes
 * "bVII" and "V7/V" becomes "V7ofV", so the name is filesystem-safe.
 */
export function midiFileName(tl: Timeline): string {
  const key = keyName(tl.key).replace('#', 's').replace(' ', '-');
  const numerals = tl.harmony
    .map((h) =>
      h.romanNumeral.replace('°', 'dim').replace('ø', 'hd').replace('♭', 'b').replace('/', 'of'),
    )
    .filter((n) => /^[A-Za-z0-9]+$/.test(n))
    .join('-');
  return `middy-${key}${numerals ? `-${numerals}` : ''}.mid`;
}
