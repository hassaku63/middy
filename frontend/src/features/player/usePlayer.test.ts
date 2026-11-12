import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import type { Synth } from './TimelinePlayer';
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

describe('usePlayer from a point (BL-327)', () => {
  /** Four quarter notes at 120 BPM; the clock stands still at 0. */
  const fourBeats: Timeline = {
    ...timeline,
    tracks: [
      {
        name: 'chords',
        channel: 0,
        program: 0,
        notes: [60, 62, 64, 65].map((pitch, i) => ({
          pitch,
          velocity: 80,
          start: i * 480,
          duration: 480,
        })),
      },
    ],
  };

  function engine() {
    const play = vi.fn<Synth['play']>();
    const e: AudioEngine = {
      synth: { play, allOff: vi.fn() },
      clock: { now: () => 0 },
      resume: () => Promise.resolve(),
    };
    return { e, play };
  }

  it('stopped, a seek before the first play sets where it starts; stop goes back there', async () => {
    const { e, play } = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    act(() => result.current.seek(960));
    expect(result.current.positionTick).toBe(960);
    await act(() => result.current.play(fourBeats));
    expect(play.mock.calls.map(([n]) => n.pitch)).toEqual([64]);
    act(() => result.current.stop());
    expect(result.current.positionTick).toBe(960);
    await act(() => result.current.play(fourBeats));
    expect(result.current.positionTick).toBe(960);
    act(() => result.current.stop());
  });

  it('playing, a seek moves on from there and stop goes back to it', async () => {
    const { e } = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    await act(() => result.current.play(fourBeats));
    act(() => result.current.seek(1440));
    expect(result.current.state).toBe('playing');
    expect(result.current.positionTick).toBe(1440);
    act(() => result.current.stop());
    expect(result.current.positionTick).toBe(1440);
  });

  it('reset stops and sets play to start at the beginning', async () => {
    const { e } = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    act(() => result.current.seek(960));
    await act(() => result.current.play(fourBeats));
    act(() => result.current.reset());
    expect(result.current.state).toBe('stopped');
    expect(result.current.positionTick).toBe(0);
    await act(() => result.current.play(fourBeats));
    expect(result.current.positionTick).toBe(0);
    act(() => result.current.stop());
  });
});
