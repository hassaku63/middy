import { useEffect, useRef } from 'react';
import type { PatternsResponse, PitchClass, Timeline } from '@/shared/api/client';
import { useKeyChange } from '@/features/key-change/useKeyChange';
import { useRegroove } from '@/features/regroove/useRegroove';

/**
 * The two changes to the open example that the engine makes without the
 * AI: another key (BL-078) and another style (BL-087). A change that
 * finishes after another example has opened is dropped, so it cannot put
 * the old example back over the new one.
 */
export function useExampleChanges(
  timeline: Timeline | null,
  show: (tl: Timeline) => void,
  styles: PatternsResponse['styles'],
) {
  const keyChange = useKeyChange();
  const regroove = useRegroove();
  const current = useRef(timeline);
  useEffect(() => {
    current.current = timeline;
  }, [timeline]);

  const changeKey = async (tonic: PitchClass) => {
    if (!timeline) return;
    regroove.clear(); // its before → after note was about this timeline
    const next = await keyChange.change(timeline, tonic);
    if (next && current.current === timeline) show(next);
  };

  const changeStyle = async (id: string) => {
    const style = styles.find((s) => s.id === id);
    if (!timeline || !style) return;
    const next = await regroove.change(timeline, style);
    if (next && current.current === timeline) show(next);
  };

  /** Forgets notes and errors about the previous example. */
  const forget = () => {
    regroove.clear();
    keyChange.clear();
  };

  return { keyChange, regroove, changeKey, changeStyle, forget };
}
