import type { HarmonyEvent } from '@/shared/api/client';

/** Pixels per bar; every lane shares this scale so columns line up. */
export const BAR_WIDTH = 160;

/** Index of the harmony event sounding at `tick`, or -1. */
export function harmonyAt(harmony: HarmonyEvent[], tick: number): number {
  return harmony.findIndex((h) => tick >= h.start && tick < h.start + h.duration);
}
