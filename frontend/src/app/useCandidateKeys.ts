import { useEffect, useRef } from 'react';
import type { IdeaCandidate } from '@/shared/api/client';
import { typing } from './useChordKeys';

/**
 * 1, 2 and 3 open the latest answer's examples in turn (the desk under
 * trial, BL-186, #279). Digits typed into a field stay with the field.
 */
export function useCandidateKeys(
  enabled: boolean,
  candidates: IdeaCandidate[],
  open: (candidate: IdeaCandidate) => void,
) {
  // The latest callback, without re-adding the listener on every render.
  const openRef = useRef(open);
  useEffect(() => {
    openRef.current = open;
  });
  useEffect(() => {
    if (!enabled || candidates.length === 0) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.defaultPrevented || e.altKey || e.ctrlKey || e.metaKey || typing(e.target)) return;
      const candidate = /^[1-9]$/.test(e.key) ? candidates[Number(e.key) - 1] : undefined;
      if (!candidate) return;
      e.preventDefault();
      openRef.current(candidate);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [enabled, candidates]);
}
