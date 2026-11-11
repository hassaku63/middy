import { useEffect, useState, type RefObject } from 'react';

/**
 * The rendered size of an element, followed as it changes (#402 F). Null
 * until measured, and where ResizeObserver is missing (tests), so callers
 * keep their natural size. It reads the element's own box, not the
 * window's (the layout rules forbid branching on the window, ADR-0042).
 */
export function useBoxSize(
  ref: RefObject<Element | null>,
  enabled: boolean,
): { width: number; height: number } | null {
  const [size, setSize] = useState<{ width: number; height: number } | null>(null);
  useEffect(() => {
    const el = ref.current;
    if (!enabled || !el || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(([entry]) => {
      if (!entry) return;
      const { width, height } = entry.contentRect;
      setSize((old) =>
        old && Math.abs(old.width - width) < 1 && Math.abs(old.height - height) < 1
          ? old
          : { width, height },
      );
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [ref, enabled]);
  return enabled ? size : null;
}
