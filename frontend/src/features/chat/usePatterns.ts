import { useEffect, useState } from 'react';
import { fetchPatterns, type PatternsResponse } from '@/shared/api/client';

const EMPTY: PatternsResponse = { accompaniments: [], drums: [], feels: [], styles: [] };

/**
 * Loads the pattern catalogue. Until it is in, the lists stay empty; a
 * failed load is tried again each time `reachable` turns true (the backend
 * health probe answered), so a backend that was down at start-up does not
 * leave the catalogue empty for the whole session.
 */
export function usePatterns(reachable = true): PatternsResponse {
  const [patterns, setPatterns] = useState<PatternsResponse>(EMPTY);
  const loaded = patterns !== EMPTY;
  useEffect(() => {
    if (!reachable || loaded) return;
    let cancelled = false;
    fetchPatterns()
      .then((p) => !cancelled && setPatterns(p))
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, [reachable, loaded]);
  return patterns;
}
