import { useEffect, type RefObject } from 'react';

/**
 * Keeps x (unscaled drawing pixels) inside the visible part of a horizontal
 * scroller: when the playhead leaves the left 10–70 % band, the view jumps
 * so it sits at 20 %. Jumps, not smooth scrolling, so reading stays stable.
 */
export function useFollowPlayhead(ref: RefObject<HTMLElement | null>, x: number): void {
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const left = el.scrollLeft;
    const w = el.clientWidth;
    if (x < left + w * 0.1 || x > left + w * 0.7) {
      el.scrollLeft = Math.max(0, x - w * 0.2);
    }
  }, [ref, x]);
}
