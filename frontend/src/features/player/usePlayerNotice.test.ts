import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import type { PlayerNotice } from './usePlayer';
import { usePlayerNotice } from './usePlayerNotice';

const timeline = {
  ppq: 480,
  timeSignature: { numerator: 4, denominator: 4 },
} as Timeline;

describe('usePlayerNotice (#500)', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  const render = (notice: PlayerNotice | null, playing = false) =>
    renderHook(
      (p: { notice: PlayerNotice | null; playing: boolean }) =>
        usePlayerNotice(p.notice, timeline, p.playing),
      { initialProps: { notice, playing } },
    );

  it.each([
    [{ kind: 'pause', tick: 1920 + 480, n: 1 }, '2 小節目 2 拍目で一時停止'],
    [{ kind: 'stop', tick: 0, n: 1 }, '停止。先頭に戻りました'],
    [{ kind: 'end', tick: 0, n: 1 }, '最後まで再生しました'],
  ] as const)('says %o as %s at once', (notice, words) => {
    const { result } = render(notice);
    act(() => {
      vi.advanceTimersByTime(0);
    });
    expect(result.current).toBe(words);
  });

  it('says a seek once the playhead settles, and nothing once it plays again', () => {
    const { result, rerender } = render({ kind: 'seek', tick: 3840, n: 1 });
    act(() => {
      vi.advanceTimersByTime(300);
    });
    expect(result.current).toBe('');
    act(() => {
      vi.advanceTimersByTime(400);
    });
    expect(result.current).toBe('3 小節目 1 拍目へ');
    rerender({ notice: { kind: 'seek', tick: 1920, n: 2 }, playing: true });
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current).toBe('');
  });
});
