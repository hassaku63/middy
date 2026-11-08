import { useCallback, useState } from 'react';
import { createTimelineFromProgression, type PitchClass, type Timeline } from '@/shared/api/client';
import { rebuildRequest } from '@/shared/lib/rebuildRequest';

export type KeyChange = {
  busy: boolean;
  error: string | null;
  /** Renders `timeline` in `tonic`; resolves to the new timeline, or null on failure. */
  change: (timeline: Timeline, tonic: PitchClass) => Promise<Timeline | null>;
  /** Forgets the last error, e.g. when another example opens. */
  clear: () => void;
};

/** Re-renders the open example in another key through the engine, never the LLM. */
export function useKeyChange(): KeyChange {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const change = useCallback(async (timeline: Timeline, tonic: PitchClass) => {
    setBusy(true);
    setError(null);
    try {
      return await createTimelineFromProgression(rebuildRequest(timeline, tonic));
    } catch {
      setError('キーを変えられませんでした');
      return null;
    } finally {
      setBusy(false);
    }
  }, []);

  const clear = useCallback(() => setError(null), []);

  return { busy, error, change, clear };
}
