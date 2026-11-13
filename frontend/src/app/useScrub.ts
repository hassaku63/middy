import { useLayoutEffect, useRef } from 'react';
import type { UsePlayer } from '@/features/player/usePlayer';
import type { Timeline } from '@/shared/api/client';

/**
 * Dragging the playhead along the timeline (BL-335), as a media player's
 * seek bar: playing, the sound holds while the drag goes on, so the
 * drawing does not move under the pointer, and play goes on from where
 * the drag ended. Stopped, the drag only moves where play will start.
 * A drag that ends after another example opened does not play the one it
 * began on (#500).
 */
export function useScrub(player: UsePlayer, timeline: Timeline) {
  /** The example to play again when the drag ends; null when none. */
  const resume = useRef<Timeline | null>(null);
  const shown = useRef(timeline);
  // Before a drawing taken away ends its drag (its effects' clean-up).
  useLayoutEffect(() => {
    shown.current = timeline;
  });
  return (phase: 'start' | 'end') => {
    if (phase === 'start') {
      resume.current = player.state === 'playing' ? timeline : null;
      if (resume.current) player.pause({ quiet: true });
      return;
    }
    const again = resume.current;
    resume.current = null;
    if (again && again === shown.current) void player.play(again);
  };
}
