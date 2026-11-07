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
  /** Replaces the current timeline with one obtained elsewhere (e.g. an idea candidate). */
  open: (timeline: Timeline) => void;
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

  const open = useCallback((tl: Timeline) => {
    setTimeline(tl);
    setError(null);
  }, []);

  return { timeline, busy, error, generate, open };
}
