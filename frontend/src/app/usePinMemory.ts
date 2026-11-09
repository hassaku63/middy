import { useCallback, useEffect, useRef } from 'react';
import type { IdeaCandidate } from '@/shared/api/client';

/**
 * Remembers the chord pinned on each example, so that going back to an
 * earlier one shows it as it was left (BL-186, #279). Playback is not
 * remembered: switching examples stops it, and play starts from the top.
 */
export function usePinMemory(open: IdeaCandidate | null, pinned: number | null) {
  const pins = useRef(new WeakMap<IdeaCandidate, number | null>());
  useEffect(() => {
    if (open) pins.current.set(open, pinned);
  }, [open, pinned]);
  return useCallback((c: IdeaCandidate) => pins.current.get(c) ?? null, []);
}
