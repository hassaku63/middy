import { useEffect, useRef } from 'react';
import type { PatternsResponse, PitchClass, Timeline } from '@/shared/api/client';
import { useKeyChange } from '@/features/key-change/useKeyChange';
import { useRegroove } from '@/features/regroove/useRegroove';

/**
 * The two changes to the open example that the engine makes without the
 * AI: another key (BL-078) and another style (BL-087). A change that
 * finishes after another example has opened is dropped, so it cannot put
 * the old example back over the new one, and the note or error it left is
 * cleared, since it was about the old example. Opening counts even when
 * the same saved example is opened again.
 */
export function useExampleChanges(
  timeline: Timeline | null,
  show: (tl: Timeline) => void,
  styles: PatternsResponse['styles'],
) {
  const keyChange = useKeyChange();
  const regroove = useRegroove();
  const current = useRef(timeline);
  const opened = useRef(0); // counts forget(), i.e. examples opened
  useEffect(() => {
    current.current = timeline;
  }, [timeline]);

  /** Shows `next` unless the example changed meanwhile; if another one opened, `clear`s the stale note or error. */
  const settle = (from: Timeline, at: number, next: Timeline | null, clear: () => void) => {
    if (opened.current !== at) {
      clear();
      return;
    }
    if (next && current.current === from) show(next);
  };

  const changeKey = async (tonic: PitchClass) => {
    if (!timeline) return;
    regroove.clear(); // its before → after note was about this timeline
    const at = opened.current;
    settle(timeline, at, await keyChange.change(timeline, tonic), keyChange.clear);
  };

  const changeStyle = async (id: string) => {
    const style = styles.find((s) => s.id === id);
    if (!timeline || !style) return;
    keyChange.clear(); // an earlier key error is not about this change
    const at = opened.current;
    settle(timeline, at, await regroove.change(timeline, style), regroove.clear);
  };

  /** Forgets notes and errors about the previous example. */
  const forget = () => {
    opened.current += 1;
    regroove.clear();
    keyChange.clear();
  };

  return { keyChange, regroove, changeKey, changeStyle, forget };
}
