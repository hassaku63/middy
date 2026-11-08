import { useCallback, useState } from 'react';

/**
 * The degree map under trial (BL-099), shown with ?degrees. Toggling it from
 * the ?lab switcher rewrites the URL so a reload or a shared link keeps it.
 * Off by default; the owner decides whether to adopt it.
 */
export function useDegreeMapTrial() {
  const [on, setOn] = useState(() => new URLSearchParams(window.location.search).has('degrees'));

  const toggle = useCallback((next: boolean) => {
    setOn(next);
    const url = new URL(window.location.href);
    if (next) url.searchParams.set('degrees', '');
    else url.searchParams.delete('degrees');
    window.history.replaceState(null, '', url);
  }, []);

  return { on, toggle };
}
