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

  it('stopped, a seek before the first play sets where it starts', async () => {
    const { e, play } = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    act(() => result.current.seek(960));
    expect(result.current.position.get()).toBe(960);
    await act(() => result.current.play(fourBeats));
    expect(play.mock.calls.map(([n]) => n.pitch)).toEqual([64]);
    act(() => result.current.stop());
  });

  it('pause stays where it is and play goes on from there (BL-334)', async () => {
    const { e } = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    await act(() => result.current.play(fourBeats));
    act(() => result.current.seek(1440));
    expect(result.current.state).toBe('playing');
    act(() => result.current.pause());
    expect(result.current.state).toBe('stopped');
    expect(result.current.position.get()).toBe(1440);
    await act(() => result.current.play(fourBeats));
    expect(result.current.position.get()).toBe(1440);
    act(() => result.current.pause());
  });

  it('stop goes back to the beginning, playing or not (BL-334)', async () => {
    const { e } = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    act(() => result.current.seek(960));
    await act(() => result.current.play(fourBeats));
    act(() => result.current.stop());
    expect(result.current.state).toBe('stopped');
    expect(result.current.position.get()).toBe(0);
    act(() => result.current.seek(960));
    act(() => result.current.stop());
    expect(result.current.position.get()).toBe(0);
    await act(() => result.current.play(fourBeats));
    expect(result.current.position.get()).toBe(0);
    act(() => result.current.stop());
  });
});

describe('what a screen reader hears from the player (#500)', () => {
  function engine() {
    const e: AudioEngine = {
      synth: { play: vi.fn(), allOff: vi.fn() },
      clock: { now: () => 0 },
      resume: () => Promise.resolve(),
    };
    return e;
  }

  it('tells a pause, a stop and a seek while silent, and not a quiet pause', async () => {
    const e = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    act(() => result.current.seek(240));
    expect(result.current.notice).toMatchObject({ kind: 'seek', tick: 240 });
    await act(() => result.current.play(timeline));
    act(() => result.current.pause());
    expect(result.current.notice).toMatchObject({ kind: 'pause', tick: 240 });
    await act(() => result.current.play(timeline));
    const before = result.current.notice?.n;
    act(() => result.current.pause({ quiet: true }));
    expect(result.current.notice?.n).toBe(before);
    act(() => result.current.stop());
    expect(result.current.notice).toMatchObject({ kind: 'stop', tick: 0 });
  });

  it('says nothing for a seek while it plays', async () => {
    const e = engine();
    const { result } = renderHook(() => usePlayer(() => e));
    await act(() => result.current.play(timeline));
    act(() => result.current.seek(120));
    expect(result.current.notice).toBeNull();
    act(() => result.current.pause());
  });
});

describe('a play that makes no sound (#500)', () => {
  it('says so when the audio device cannot be made, and plays on the next try', async () => {
    const play = vi.fn();
    const working: AudioEngine = {
      synth: { play, allOff: vi.fn() },
      clock: { now: () => 0 },
      resume: () => Promise.resolve(),
    };
    const create = vi
      .fn<() => AudioEngine>()
      .mockImplementationOnce(() => {
        throw new Error('no AudioContext');
      })
      .mockReturnValue(working);
    const { result } = renderHook(() => usePlayer(create));
    await act(() => result.current.play(timeline));
    expect(result.current.error).toMatch(/音を鳴らせませんでした/);
    expect(result.current.state).toBe('stopped');

    await act(() => result.current.play(timeline));
    expect(result.current.error).toBeNull();
    expect(result.current.state).toBe('playing');
    expect(play).toHaveBeenCalled();
  });

  it('says so when the browser will not start the audio', async () => {
    const engine: AudioEngine = {
      synth: { play: vi.fn(), allOff: vi.fn() },
      clock: { now: () => 0 },
      resume: () => Promise.reject(new Error('not allowed')),
    };
    const { result } = renderHook(() => usePlayer(() => engine));
    await act(() => result.current.play(timeline));
    expect(result.current.error).toMatch(/音を鳴らせませんでした/);
  });
});
