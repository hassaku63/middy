import type { HarmonyEvent } from '@/shared/api/client';

/** The chord the panel should describe: pinned one, else the one under the playhead. */
export function explainedIndex(
  harmony: HarmonyEvent[],
  positionTick: number,
  selected: number | null,
): number {
  if (selected !== null && selected >= 0 && selected < harmony.length) return selected;
  return harmony.findIndex((h) => positionTick >= h.start && positionTick < h.start + h.duration);
}

/** The degree to light: only a chord of the key sits on the map, not a borrowed or applied one. */
export function currentDegree(event: HarmonyEvent | undefined): number | null {
  const a = event?.analysis;
  return a && !a.borrowed && a.appliedTo === undefined ? a.degree : null;
}
