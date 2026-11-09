import type { HarmonyEvent } from '@/shared/api/client';

/** The degree the map lights: only a chord of the key sits on it, not a borrowed or applied one. */
export function currentDegree(event: HarmonyEvent | undefined): number | null {
  const a = event?.analysis;
  return a && !a.borrowed && a.appliedTo === undefined ? a.degree : null;
}
