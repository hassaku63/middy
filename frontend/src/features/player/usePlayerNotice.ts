import { useEffect, useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import { spokenPosition } from '@/shared/lib/timing';
import type { PlayerNotice } from './usePlayer';

/** How long a run of seeks (a drag) settles before its end is said. */
const SEEK_SETTLE_MS = 600;

/**
 * The words a screen reader hears from the player (#500): only when the
 * music is silent, so nothing is said over it. A pause, a stop and the end
 * are said at once; a seek once the playhead has settled; nothing while it
 * plays, and a pending word is dropped when it starts again.
 */
export function usePlayerNotice(
  notice: PlayerNotice | null,
  timeline: Timeline | null,
  playing: boolean,
): string {
  const [said, setSaid] = useState('');
  useEffect(() => {
    if (playing || !notice || !timeline) return;
    const where = spokenPosition(notice.tick, timeline.ppq, timeline.timeSignature);
    const words = {
      pause: `${where}で一時停止`,
      stop: '停止。先頭に戻りました',
      end: '最後まで再生しました',
      seek: `${where}へ`,
    }[notice.kind];
    const delay = notice.kind === 'seek' ? SEEK_SETTLE_MS : 0;
    const timer = setTimeout(() => {
      setSaid(words);
    }, delay);
    return () => {
      clearTimeout(timer);
    };
  }, [notice, timeline, playing]);
  // While it plays, the region is empty: nothing to say over the music.
  return playing ? '' : said;
}
