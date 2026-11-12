import { useRef, useState, type MouseEvent, type PointerEvent, type RefObject } from 'react';
import { beatAt, type Layout } from './layout';

/** A drag that moves the playhead begins and ends (the app pauses play meanwhile). */
export type Scrub = (phase: 'start' | 'end') => void;

export type SeekByClick = {
  /** The beat under a mouse or pen, in ticks: where a click would go. */
  hint: number | null;
  handlers: {
    onClick?: (e: MouseEvent<SVGSVGElement>) => void;
    onPointerDown?: (e: PointerEvent<SVGSVGElement>) => void;
    onPointerMove?: (e: PointerEvent<SVGSVGElement>) => void;
    onPointerUp?: () => void;
    onPointerCancel?: () => void;
    onPointerLeave?: () => void;
  };
};

/**
 * Clicking the drawing anywhere but on a chord moves the playhead to the
 * start of the beat under it (BL-328, ADR-0067); a chord still pins, when
 * chords pin. With a mouse or a pen the playhead moves on the press, and
 * keeps moving with a drag until the release (BL-335), as a media
 * player's seek bar; scrub tells the app when such a drag starts and ends.
 * A finger only clicks: the browser does not click after a finger panned
 * the box (WCAG 2.5.2). A mouse or a pen shows where it would go first.
 */
export function useSeekByClick(
  ref: RefObject<SVGSVGElement | null>,
  layout: Layout,
  onSeek: ((tick: number) => void) | undefined,
  chordsPin: boolean,
  scrub?: Scrub,
): SeekByClick {
  const [hint, setHint] = useState<number | null>(null);
  /** The beat a drag last moved to; null when no drag is on. */
  const dragging = useRef<number | null>(null);
  /** The click that follows a mouse's press is the press itself. */
  const pressed = useRef(false);
  if (!onSeek) return { hint: null, handlers: {} };
  const beat = (e: MouseEvent<SVGSVGElement>) => {
    const box = ref.current?.getBoundingClientRect();
    if (!box || box.width === 0) return null;
    return beatAt(layout, ((e.clientX - box.left) * layout.width) / box.width);
  };
  const at = (e: MouseEvent<SVGSVGElement>) =>
    chordsPin && e.target instanceof Element && e.target.closest('[role=option]') ? null : beat(e);
  const end = () => {
    if (dragging.current === null) return;
    dragging.current = null;
    scrub?.('end');
  };
  return {
    hint,
    handlers: {
      onClick: (e) => {
        if (pressed.current) {
          pressed.current = false;
          return;
        }
        const tick = at(e);
        if (tick !== null) onSeek(tick);
      },
      onPointerDown: (e) => {
        if (e.pointerType === 'touch' || e.button !== 0) return;
        const tick = at(e);
        if (tick === null) return;
        pressed.current = true;
        dragging.current = tick;
        scrub?.('start');
        onSeek(tick);
        e.currentTarget.setPointerCapture(e.pointerId);
      },
      onPointerMove: (e) => {
        if (dragging.current === null) {
          setHint(e.pointerType === 'touch' ? null : at(e));
          return;
        }
        const tick = beat(e); // over a chord too, once dragging
        if (tick === null || tick === dragging.current) return;
        dragging.current = tick;
        setHint(null);
        onSeek(tick);
      },
      onPointerUp: end,
      onPointerCancel: end,
      onPointerLeave: () => {
        setHint(null);
      },
    },
  };
}
