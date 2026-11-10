import { useCallback, useLayoutEffect, useRef, useState } from 'react';

/** A horizontal slice of the drawing, in its own units (computeLayout's pixels). */
export type Slice = { x: number; width: number };

/** How a point is kept in view: turned to like a page, or held at the centre. */
export type Keep = 'page' | 'centre';

/** The share of the view kept free before the point, when a page is turned. */
const LEAD = 0.1;

/**
 * Where to scroll so that x stays in view, all in screen pixels: null when
 * it already is. A page turns when x comes near either edge, putting x just
 * after the left edge; centre always puts it in the middle.
 */
export function scrollFor(left: number, visible: number, x: number, keep: Keep): number | null {
  if (keep === 'centre') return x - visible / 2;
  const lead = visible * LEAD;
  if (x >= left + lead && x <= left + visible - lead) return null;
  return x - lead;
}

/**
 * The scrolling box of a long timeline (#360): scrolls it to keep target
 * (in drawing units) in view whenever target moves, and reports the slice
 * in view for the overview to outline. Between target moves the person
 * scrolls freely.
 */
export function useTimelineScroll(width: number, target: number | null, keep: Keep) {
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
      const left = scrollFor(box.scrollLeft, box.clientWidth, target * scale, keep);
      if (left !== null) box.scrollLeft = left;
    }
    measure();
  }, [target, keep, width, measure]);

  return { ref, view, onScroll: measure };
}
