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

/** How far, in screen pixels, the box may sit from where it was put before it counts as moved by hand. */
const MOVED = 2;

/**
 * The scrolling box of a long timeline (#360): scrolls it to keep target
 * (in drawing units) in view whenever target moves, and reports the slice
 * in view for the overview to outline. Between target moves the person
 * scrolls freely. Playing, scrolling by hand pauses the following, as a
 * DAW's follow does (BL-332, ADR-0068): resume() or the next play takes it
 * up again.
 */
export function useTimelineScroll(
  width: number,
  target: number | null,
  keep: Keep,
  inset = 0,
  playing = false,
) {
  const ref = useRef<HTMLDivElement>(null);
  const [view, setView] = useState<Slice>({ x: 0, width });
  const [paused, setPaused] = useState(false);
  // Playing or stopping takes the following up again.
  const [wasPlaying, setWasPlaying] = useState(playing);
  if (wasPlaying !== playing) {
    setWasPlaying(playing);
    setPaused(false);
  }
  /** Where this hook last left the box; elsewhere, a hand moved it. */
  const placed = useRef<number | null>(null);
  const movedByHand = useCallback((box: HTMLDivElement) => {
    const at = placed.current;
    return at !== null && Math.abs(box.scrollLeft - at) > MOVED;
  }, []);

  const measure = useCallback(() => {
    const box = ref.current;
    if (!box?.scrollWidth) return; // not laid out (jsdom)
    const scale = box.scrollWidth / width;
    const next = { x: box.scrollLeft / scale, width: box.clientWidth / scale };
    setView((prev) => (prev.x === next.x && prev.width === next.width ? prev : next));
  }, [width]);

  useLayoutEffect(() => {
    const box = ref.current;
    if (!box) return;
    if (playing && !paused && movedByHand(box)) {
      setPaused(true); // a scroll the scroll event has not reported yet
    } else if (box.scrollWidth && target !== null && !(playing && paused)) {
      const scale = box.scrollWidth / width;
      const left = scrollFor(box.scrollLeft, box.clientWidth, target * scale, keep, inset * scale);
      if (left !== null) box.scrollLeft = left;
    }
    placed.current = box.scrollLeft;
    measure();
  }, [target, keep, width, inset, measure, playing, paused, movedByHand]);

  const onScroll = useCallback(() => {
    const box = ref.current;
    if (box && playing && movedByHand(box)) setPaused(true);
    // Stopped, a hand may put the box anywhere: play follows from there.
    if (box && !playing) placed.current = box.scrollLeft;
    measure();
  }, [measure, playing, movedByHand]);

  const resume = useCallback(() => {
    setPaused(false);
  }, []);

  /**
   * Scrolls the box where the person asked to look from the overview strip
   * (ADR-0069): x in the middle, or, given from, the box's left edge at x.
   * Playing, that pauses the following as a scroll by hand does. Returns
   * the box's left edge in drawing units, where the browser stopped it.
   */
  const look = useCallback(
    (x: number, from: 'centre' | 'left' = 'centre') => {
      const box = ref.current;
      if (!box?.scrollWidth) return 0;
      const scale = box.scrollWidth / width;
      const left =
        from === 'left'
          ? x * scale
          : scrollFor(box.scrollLeft, box.clientWidth, x * scale, 'centre', inset * scale);
      if (left !== null) box.scrollLeft = left; // centre always gives one
      placed.current = box.scrollLeft;
      if (playing) setPaused(true);
      measure();
      return box.scrollLeft / scale;
    },
    [width, inset, playing, measure],
  );

  return { ref, view, onScroll, paused: playing && paused, resume, look };
}
