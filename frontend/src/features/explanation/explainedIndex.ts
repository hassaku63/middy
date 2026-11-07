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
