import { useCallback, useState } from 'react';

/**
 * How the long timeline's detail moves, under trial (BL-127): following
 * the playhead (B, the default since the owner leans to it) or by four-bar
 * pages (A, ?zoom=page). Choosing from the ?lab switcher rewrites the URL
 * so a reload or a shared link keeps it. The one the owner does not choose
 * goes.
 */
export function useZoomTrial() {
  const params = new URLSearchParams(window.location.search);
  const [follow, setFollow] = useState(() => params.get('zoom') !== 'page');
  const lab = params.has('lab');

  const choose = useCallback((next: boolean) => {
    setFollow(next);
    const url = new URL(window.location.href);
    if (next) url.searchParams.delete('zoom');
    else url.searchParams.set('zoom', 'page');
    window.history.replaceState(null, '', url);
  }, []);

  return { follow, choose, lab };
}
