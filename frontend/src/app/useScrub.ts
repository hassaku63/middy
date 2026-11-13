import { useRef } from 'react';
import type { UsePlayer } from '@/features/player/usePlayer';
import type { Timeline } from '@/shared/api/client';

/**
 * Dragging the playhead along the timeline (BL-335), as a media player's
 * seek bar: playing, the sound holds while the drag goes on, so the
 * drawing does not move under the pointer, and play goes on from where
 * the drag ended. Stopped, the drag only moves where play will start.
 */
export function useScrub(player: UsePlayer, timeline: Timeline) {
  const resume = useRef(false);
  return (phase: 'start' | 'end') => {
    if (phase === 'start') {
      resume.current = player.state === 'playing';
      if (resume.current) player.pause({ quiet: true });
      return;
    }
    if (!resume.current) return;
    resume.current = false;
    void player.play(timeline);
  };
}
