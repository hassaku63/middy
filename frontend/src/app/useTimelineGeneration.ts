import { useCallback, useState } from 'react';
import {
  ApiError,
  createTimelineFromProgression,
  type FromProgressionRequest,
  type Timeline,
} from '@/shared/api/client';

export type TimelineGeneration = {
  timeline: Timeline | null;
  busy: boolean;
  error: string | null;
  /** Resolves to true when a new timeline was produced. */
  generate: (req: FromProgressionRequest) => Promise<boolean>;
};

/** Owns the "form -> API -> timeline" call so components stay side-effect free. */
export function useTimelineGeneration(): TimelineGeneration {
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const generate = useCallback(async (req: FromProgressionRequest) => {
    setBusy(true);
    setError(null);
    try {
      setTimeline(await createTimelineFromProgression(req));
      return true;
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'バックエンドに接続できません');
      return false;
    } finally {
      setBusy(false);
    }
  }, []);

  return { timeline, busy, error, generate };
}
