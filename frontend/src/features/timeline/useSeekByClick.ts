import { useState, type MouseEvent, type PointerEvent, type RefObject } from 'react';
import { beatAt, type Layout } from './layout';

export type SeekByClick = {
  /** The beat under a mouse or pen, in ticks: where a click would go. */
  hint: number | null;
  handlers: {
    onClick?: (e: MouseEvent<SVGSVGElement>) => void;
    onPointerMove?: (e: PointerEvent<SVGSVGElement>) => void;
    onPointerLeave?: () => void;
  };
};

/**
 * Clicking the drawing anywhere but on a chord moves the playhead to the
 * start of the beat under it (BL-328, ADR-0067); a chord still pins. A
 * click, not a press: the browser does not click after a finger panned the
 * box (WCAG 2.5.2). A mouse or a pen shows where it would go first.
 */
export function useSeekByClick(
  ref: RefObject<SVGSVGElement | null>,
  layout: Layout,
  onSeek: ((tick: number) => void) | undefined,
): SeekByClick {
  const [hint, setHint] = useState<number | null>(null);
  if (!onSeek) return { hint: null, handlers: {} };
  const at = (e: MouseEvent<SVGSVGElement>) => {
    const box = ref.current?.getBoundingClientRect();
    if (!box || box.width === 0) return null;
    if (e.target instanceof Element && e.target.closest('[role=option]')) return null;
    return beatAt(layout, ((e.clientX - box.left) * layout.width) / box.width);
  };
  return {
    hint,
    handlers: {
      onClick: (e) => {
        const tick = at(e);
        if (tick !== null) onSeek(tick);
      },
      onPointerMove: (e) => {
        setHint(e.pointerType === 'touch' ? null : at(e));
      },
      onPointerLeave: () => {
        setHint(null);
      },
    },
  };
}
