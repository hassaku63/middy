import { useCallback, useState } from 'react';
import { isTimelineLayout, type TimelineLayout } from '@/features/timeline/lab/variants';

/**
 * The long-timeline layout under trial (BL-080), from ?timeline=<id>; ?lab
 * shows a switcher. Choosing one rewrites the URL so a reload or a shared
 * link keeps it. Default is today's layout.
 */
export function useTimelineLayout() {
  const params = new URLSearchParams(window.location.search);
  const fromUrl = params.get('timeline');
  const [layout, setLayout] = useState<TimelineLayout>(isTimelineLayout(fromUrl) ? fromUrl : 'fit');
  const lab = params.has('lab') || fromUrl !== null;

  const choose = useCallback((next: TimelineLayout) => {
    setLayout(next);
    const url = new URL(window.location.href);
    url.searchParams.set('timeline', next);
    window.history.replaceState(null, '', url);
  }, [setLayout]);

  return { layout, lab, choose };
}
