import type { ProgressionLabel } from '@/shared/api/client';

/**
 * Splits user text such as "I - V - vi - IV", "ii V7 I" or "I|IV|V|I" into
 * progression labels. Validation of each label is the backend's job
 * (music.ParseRomanNumeral); this only tokenises.
 */
export function parseProgressionText(text: string, beats: number): ProgressionLabel[] {
  return text
    .split(/[\s,|\-–—→]+/)
    .map((t) => t.trim())
    .filter((t) => t.length > 0)
    .map((numeral) => ({ numeral, beats }));
}
