import { useCallback, useState } from 'react';

/** The layouts under trial (#279); "stacked" is today's. */
export type Layout = 'stacked' | 'desk' | 'wide';

const layouts: readonly Layout[] = ['stacked', 'desk', 'wide'];

/**
 * The screen layout under trial (BL-184, BL-197, #279): today's, with the
 * conversation above the result; the desk (?layout=desk), with the
 * conversation in a left column beside the stage on wide screens; or the
 * wide desk (?layout=wide), the same column with the timeline across the
 * whole stage and the explanations side by side below it (#314).
 * Choosing from the ?lab switcher rewrites the URL so a reload or a shared
 * link keeps it. The ones the owner does not choose go.
 */
export function useLayoutTrial() {
  const params = new URLSearchParams(window.location.search);
  const [mode, setMode] = useState<Layout>(() => {
    const asked = params.get('layout');
    return layouts.find((l) => l === asked) ?? 'stacked';
  });
  const lab = params.has('lab');
  // Where the desk's input sits, under trial too (BL-218): pinned to the
  // foot of the column (today's), or right after the conversation
  // (?input=follow), so it rises into the space a short one leaves and
  // sticks to the foot once the conversation fills the column.
  const [followInput, setFollowInput] = useState(() => params.get('input') === 'follow');

  const chooseInput = useCallback((follow: boolean) => {
    setFollowInput(follow);
    const url = new URL(window.location.href);
    if (follow) url.searchParams.set('input', 'follow');
    else url.searchParams.delete('input');
    window.history.replaceState(null, '', url);
  }, []);

  const choose = useCallback((next: Layout) => {
    setMode(next);
    const url = new URL(window.location.href);
    if (next === 'stacked') url.searchParams.delete('layout');
    else url.searchParams.set('layout', next);
    window.history.replaceState(null, '', url);
  }, []);

  // The wide desk is a desk too: the left column, the frame and the keys.
  return {
    mode,
    desk: mode !== 'stacked',
    wide: mode === 'wide',
    choose,
    lab,
    followInput,
    chooseInput,
  };
}
