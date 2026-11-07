import { useEffect, useState } from 'react';
import { fetchPatterns, type PatternsResponse } from '@/shared/api/client';

const EMPTY: PatternsResponse = { accompaniments: [], drums: [] };

/** Loads the pattern catalogue once; an unreachable backend leaves the lists empty. */
export function usePatterns(): PatternsResponse {
  const [patterns, setPatterns] = useState<PatternsResponse>(EMPTY);
  useEffect(() => {
    let cancelled = false;
    fetchPatterns()
      .then((p) => !cancelled && setPatterns(p))
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, []);
  return patterns;
}
