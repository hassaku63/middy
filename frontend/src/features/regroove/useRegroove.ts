import { useCallback, useState } from 'react';
import {
  createTimelineFromProgression,
  type PatternsResponse,
  type Timeline,
} from '@/shared/api/client';
import { grooveFor, regrooveRequest } from '@/shared/lib/rebuildRequest';
import { grooveChanges } from './grooveChanges';

type Style = PatternsResponse['styles'][number];

export type Regroove = {
  busy: boolean;
  error: string | null;
  /** What the last change did, before → after; null until a style is chosen. */
  note: string | null;
  /** Plays `timeline` in `style`; resolves to the new timeline, or null on failure. */
  change: (timeline: Timeline, style: Style) => Promise<Timeline | null>;
  /** Forgets the note, e.g. when another example opens. */
  clear: () => void;
};

/** Re-renders the open example in another style through the engine, never the LLM (BL-087). */
export function useRegroove(): Regroove {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [note, setNote] = useState<string | null>(null);

  const change = useCallback(async (timeline: Timeline, style: Style) => {
    setError(null);
    const groove = grooveFor(style, timeline);
    if (!groove) {
      setError(`${style.name}はこの拍子の型を持っていません`);
      return null;
    }
    setBusy(true);
    try {
      const next = await createTimelineFromProgression(regrooveRequest(timeline, style.id, groove));
      setNote(grooveChanges(style.name, timeline.rhythm, next.rhythm));
      return next;
    } catch {
      setError('スタイルを替えられませんでした');
      return null;
    } finally {
      setBusy(false);
    }
  }, []);

  const clear = useCallback(() => {
    setNote(null);
    setError(null);
  }, []);

  return { busy, error, note, change, clear };
}
