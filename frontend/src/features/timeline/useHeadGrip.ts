import { useEffect, useRef, type MouseEvent, type PointerEvent, type RefObject } from 'react';
import type { Layout } from './layout';
import type { Scrub } from './useSeekByClick';

/** Half the width a finger can hold the playhead by, in drawing units (44 px, WCAG 2.5.5). */
export const GRIP_HALF = 22;

/**
 * A finger holding the playhead (BL-337, ADR-0072): the playhead stays in
 * the middle and the score follows the finger, as in a phone's video
 * editor; a swipe to the right brings earlier bars under the playhead,
 * like pulling the score back. The position moves beat by beat. A mouse
 * drags from anywhere instead (useSeekByClick), so this takes fingers only.
 */
export function useHeadGrip(
  ref: RefObject<SVGSVGElement | null>,
  layout: Layout,
  positionTick: number,
  onSeek: ((tick: number) => void) | undefined,
  scrub: Scrub | undefined,
) {
  /** Where the finger and the playhead were when it took hold; the beat last sought. */
  const held = useRef<{ x: number; tick: number; last: number } | null>(null);
  /** A tap on the playhead leaves it where it is: its click is not a seek. */
  const tapped = useRef(false);
  // While a finger holds the playhead, its moves are this drag's, not a pan
  // of the box. CSS touch-action does not reach into the SVG, and React's
  // touch listeners are passive, so this takes them natively.
  useEffect(() => {
    const svg = ref.current;
    if (!svg || !onSeek) return;
    const hold = (e: TouchEvent) => {
      if (held.current) e.preventDefault();
    };
    svg.addEventListener('touchmove', hold, { passive: false });
    return () => {
      svg.removeEventListener('touchmove', hold);
    };
  }, [ref, onSeek]);
  if (!onSeek) return null;
  const end = () => {
    if (!held.current) return;
    held.current = null;
    scrub?.('end', 'head');
  };
  return {
    onPointerDown: (e: PointerEvent<SVGRectElement>) => {
      if (e.pointerType !== 'touch') return;
      tapped.current = true;
      held.current = { x: e.clientX, tick: positionTick, last: positionTick };
      scrub?.('start', 'head');
      e.currentTarget.setPointerCapture(e.pointerId);
    },
    onPointerMove: (e: PointerEvent<SVGRectElement>) => {
      const h = held.current;
      const box = ref.current?.getBoundingClientRect();
      if (!h || !box?.width) return;
      const perPixel = (layout.perBar / layout.barWidth) * (layout.width / box.width);
      const raw = h.tick - (e.clientX - h.x) * perPixel;
      const lastBeat = layout.bars * layout.perBar - layout.perBeat;
      const tick = Math.min(
        lastBeat,
        Math.max(0, Math.floor(raw / layout.perBeat) * layout.perBeat),
      );
      if (tick === h.last) return;
      h.last = tick;
      onSeek(tick);
    },
    onPointerUp: end,
    onPointerCancel: end,
    onClick: (e: MouseEvent<SVGRectElement>) => {
      if (!tapped.current) return;
      tapped.current = false;
      e.stopPropagation();
    },
  };
}
