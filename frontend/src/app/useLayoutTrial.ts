import { useCallback, useState } from 'react';

/**
 * The screen layout under trial (BL-184, #279): today's, with the
 * conversation above the result, or the desk (?layout=desk), with the
 * conversation in a left column beside the stage on wide screens.
 * Choosing from the ?lab switcher rewrites the URL so a reload or a shared
 * link keeps it. The one the owner does not choose goes.
 */
export function useLayoutTrial() {
  const params = new URLSearchParams(window.location.search);
  const [desk, setDesk] = useState(() => params.get('layout') === 'desk');
  const lab = params.has('lab');

  const choose = useCallback((next: boolean) => {
    setDesk(next);
    const url = new URL(window.location.href);
    if (next) url.searchParams.set('layout', 'desk');
    else url.searchParams.delete('layout');
    window.history.replaceState(null, '', url);
  }, []);

  return { desk, choose, lab };
}
