import { useCallback, useLayoutEffect, useRef, useState } from 'react';

/** A horizontal slice of the drawing, in its own units (computeLayout's pixels). */
export type Slice = { x: number; width: number };

/** How a point is kept in view: turned to like a page, or held at the centre. */
export type Keep = 'page' | 'centre';

/** The share of the view kept free before the point, when a page is turned. */
const LEAD = 0.1;

/**
 * Where to scroll so that x stays in view, all in screen pixels: null when
 * it already is. inset is the width at the box's left that covers the
 * drawing (the sticky row names), so "in view" and "the middle" are of the
 * drawing that shows. A page turns when x comes near either edge, putting
 * x just after the left edge; centre always puts it in the middle (the
 * browser stops the scroll at both ends, so at the start and the end the
 * point moves off the middle instead; BL-326).
 */
export function scrollFor(
  left: number,
  visible: number,
  x: number,
  keep: Keep,
  inset = 0,
): number | null {
  const shown = visible - inset;
  if (keep === 'centre') return x - inset - shown / 2;
  const lead = inset + shown * LEAD;
  if (x >= left + lead && x <= left + visible - shown * LEAD) return null;
  return x - lead;
}

/**
 * The scrolling box of a long timeline (#360): scrolls it to keep target
 * (in drawing units) in view whenever target moves, and reports the slice
 * in view for the overview to outline. Between target moves the person
 * scrolls freely.
 */
export function useTimelineScroll(width: number, target: number | null, keep: Keep, inset = 0) {
  const ref = useRef<HTMLDivElement>(null);
  const [view, setView] = useState<Slice>({ x: 0, width });

  const measure = useCallback(() => {
    const box = ref.current;
    if (!box?.scrollWidth) return; // not laid out (jsdom)
    const scale = box.scrollWidth / width;
    const next = { x: box.scrollLeft / scale, width: box.clientWidth / scale };
    setView((prev) => (prev.x === next.x && prev.width === next.width ? prev : next));
  }, [width]);

  useLayoutEffect(() => {
    const box = ref.current;
    if (box?.scrollWidth && target !== null) {
      const scale = box.scrollWidth / width;
      const left = scrollFor(box.scrollLeft, box.clientWidth, target * scale, keep, inset * scale);
      if (left !== null) box.scrollLeft = left;
    }
    measure();
  }, [target, keep, width, inset, measure]);

  return { ref, view, onScroll: measure };
}
