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
    // Only the API call is in the try: a fault in our own request building
    // or note is a bug, not a failed style change (conventions.md).
    const request = regrooveRequest(timeline, style.id, groove);
    setBusy(true);
    let next: Timeline;
    try {
      next = await createTimelineFromProgression(request);
    } catch {
      setError('スタイルを替えられませんでした');
      return null;
    } finally {
      setBusy(false);
    }
    setNote(grooveChanges(style.name, timeline.rhythm, next.rhythm));
    return next;
  }, []);

  const clear = useCallback(() => {
    setNote(null);
    setError(null);
  }, []);

  return { busy, error, note, change, clear };
}
