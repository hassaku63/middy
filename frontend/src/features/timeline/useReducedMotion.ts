import { useSyncExternalStore } from 'react';

const QUERY = '(prefers-reduced-motion: reduce)';

/**
 * Whether the person asked the system for less motion. Read, not sized:
 * the layout does not branch on it (docs/dev/conventions.md), only how the
 * timeline keeps the playhead in view (BL-326).
 */
export function useReducedMotion(): boolean {
  return useSyncExternalStore(subscribe, current, () => false);
}

function media(): MediaQueryList | undefined {
  return typeof window !== 'undefined' && typeof window.matchMedia === 'function'
    ? window.matchMedia(QUERY)
    : undefined;
}

function subscribe(onChange: () => void): () => void {
  const m = media();
  m?.addEventListener('change', onChange);
  return () => m?.removeEventListener('change', onChange);
}

function current(): boolean {
  return media()?.matches ?? false;
}
