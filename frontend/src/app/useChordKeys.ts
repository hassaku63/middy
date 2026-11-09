import { useEffect, useRef } from 'react';
import type { HarmonyEvent } from '@/shared/api/client';
import { explainedIndex } from '@/features/explanation/explainedIndex';

type Options = {
  /** Off unless the desk under trial shows a timeline (BL-185). */
  enabled: boolean;
  harmony: HarmonyEvent[];
  positionTick: number;
  selected: number | null;
  onSelect: (index: number | null) => void;
};

/**
 * ← and → move the explanation to the chord before or after and pin it; Esc
 * lets it follow the playhead again (BL-185, #279). Keys typed into a field
 * stay with the field.
 */
export function useChordKeys({ enabled, harmony, positionTick, selected, onSelect }: Options) {
  // The playhead moves every frame; read it at the key press instead of
  // adding the listener again on every render.
  const now = useRef({ positionTick, selected, onSelect });
  useEffect(() => {
    now.current = { positionTick, selected, onSelect };
  });
  useEffect(() => {
    if (!enabled || harmony.length === 0) return;
    const onKey = (e: KeyboardEvent) => {
      const { positionTick, selected, onSelect } = now.current;
      if (e.defaultPrevented || e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
      if (typing(e.target)) return;
      if (e.key === 'Escape') {
        if (selected === null) return;
        onSelect(null);
      } else if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
        const from = explainedIndex(harmony, positionTick, selected);
        const step = e.key === 'ArrowRight' ? 1 : -1;
        const to = from < 0 ? 0 : Math.min(Math.max(from + step, 0), harmony.length - 1);
        onSelect(to);
      } else {
        return;
      }
      e.preventDefault();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [enabled, harmony]);
}

/** A field where keys type or move the caret or the value. */
export function typing(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false;
  return (
    target.isContentEditable ||
    target instanceof HTMLInputElement ||
    target instanceof HTMLTextAreaElement ||
    target instanceof HTMLSelectElement
  );
}
