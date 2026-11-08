import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { usePlayer, type AudioEngine } from './usePlayer';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 120,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [
    {
      name: 'chords',
      channel: 0,
      program: 0,
      notes: [{ pitch: 60, velocity: 80, start: 0, duration: 480 }],
    },
  ],
  harmony: [],
};

/** An engine whose resume() the test finishes by hand, like a slow device. */
function slowEngine() {
  let resumed: () => void = () => undefined;
  const play = vi.fn();
  const close = vi.fn(() => Promise.resolve());
  const engine: AudioEngine = {
    synth: { play, allOff: vi.fn() },
    clock: { now: () => 0 },
    resume: () => new Promise<void>((r) => (resumed = r)),
    close,
  };
  return { engine, play, close, resume: () => resumed() };
}

describe('usePlayer (review)', () => {
  it('does not start a play that was stopped while the device resumed', async () => {
    const { engine, play, resume } = slowEngine();
    const { result } = renderHook(() => usePlayer(() => engine));
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = result.current.play(timeline);
    });
    act(() => result.current.stop()); // e.g. another example opened
    await act(async () => {
      resume();
      await pending;
    });
    expect(result.current.state).toBe('stopped');
    expect(play).not.toHaveBeenCalled();
  });

  it('does not touch a closed engine when it unmounts while the device resumes', async () => {
    const { engine, play, close, resume } = slowEngine();
    const { result, unmount } = renderHook(() => usePlayer(() => engine));
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = result.current.play(timeline);
    });
    unmount();
    resume();
    await expect(pending).resolves.toBeUndefined();
    expect(close).toHaveBeenCalled();
    expect(play).not.toHaveBeenCalled();
  });

  it('plays once the device has resumed', async () => {
    const { engine, resume } = slowEngine();
    const { result } = renderHook(() => usePlayer(() => engine));
    let pending: Promise<void> = Promise.resolve();
    act(() => {
      pending = result.current.play(timeline);
    });
    await act(async () => {
      resume();
      await pending;
    });
    expect(result.current.state).toBe('playing');
    act(() => result.current.stop());
  });
});
