import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { positionStore } from '@/features/player/position';
import type { UsePlayer } from '@/features/player/usePlayer';
import type { Timeline } from '@/shared/api/client';
import { useScrub } from './useScrub';

const tl = (tempoBpm: number): Timeline => ({
  ppq: 480,
  tempoBpm,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [],
});

function playing() {
  const play = vi.fn<(t: Timeline) => Promise<void>>(() => Promise.resolve());
  const pause = vi.fn();
  const player: UsePlayer = {
    state: 'playing',
    position: positionStore(),
    notice: null,
    error: null,
    play,
    pause,
    stop: vi.fn(),
    seek: vi.fn(),
  };
  return { player, play, pause };
}

describe('useScrub', () => {
  it('holds the sound while dragging and plays on from where the drag ended', () => {
    const { player, play, pause } = playing();
    const first = tl(120);
    const { result } = renderHook(() => useScrub(player, first));
    result.current('start');
    expect(pause).toHaveBeenCalledWith({ quiet: true });
    result.current('end');
    expect(play).toHaveBeenCalledWith(first);
  });

  it('does not play the example the drag began on once another is open (#500)', () => {
    const { player, play } = playing();
    const { result, rerender } = renderHook(({ t }) => useScrub(player, t), {
      initialProps: { t: tl(120) },
    });
    const began = result.current;
    began('start');
    rerender({ t: tl(90) });
    began('end'); // a drawing taken away ends its drag with the scrub it had
    expect(play).not.toHaveBeenCalled();
  });
});
