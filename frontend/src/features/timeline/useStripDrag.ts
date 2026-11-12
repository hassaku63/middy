import { useRef, type PointerEvent } from 'react';
import type { Slice } from './useTimelineScroll';

type Look = (x: number, from?: 'centre' | 'left') => number;

/**
 * The overview strip as a place to look (ADR-0069): pressing it brings the
 * box there, and dragging moves the box along (BL-331), as a DAW's
 * overview. Pressed on the frame (the part in view), the frame follows the
 * pointer from where it was held; pressed elsewhere, the box first centres
 * there. The playhead never moves. x is in drawing units.
 */
export function useStripDrag(look: Look, view: Slice, width: number) {
  /** Where the drag started: the pointer's x and the box's left edge. */
  const held = useRef<{ x: number; left: number } | null>(null);

  const at = (e: PointerEvent) => {
    const strip = e.target instanceof Element ? e.target.closest('svg') : null;
    const box = strip?.getBoundingClientRect();
    if (!strip || !box?.width) return null;
    return { strip, x: ((e.clientX - box.left) / box.width) * width };
  };

  return {
    onPointerDown: (e: PointerEvent<HTMLDivElement>) => {
      const p = at(e);
      if (!p || e.button !== 0) return;
      const onFrame = p.x >= view.x && p.x <= view.x + view.width;
      const left = onFrame ? view.x : look(p.x);
      held.current = { x: p.x, left };
      p.strip.setPointerCapture(e.pointerId);
    },
    onPointerMove: (e: PointerEvent<HTMLDivElement>) => {
      const start = held.current;
      const p = start && at(e);
      if (!start || !p) return;
      look(start.left + (p.x - start.x), 'left');
    },
    onPointerUp: () => {
      held.current = null;
    },
    onPointerCancel: () => {
      held.current = null;
    },
  };
}
