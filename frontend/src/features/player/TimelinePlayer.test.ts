import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { TimelinePlayer, type Synth } from './TimelinePlayer';

/** Four whole-note chords (C, G, Am, F) at 120 BPM: 8 seconds total. */
const fourBars: Timeline = {
  ppq: 480,
  tempoBpm: 120,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [
    {
      name: 'chords',
      channel: 0,
      program: 0,
      notes: [
        [60, 64, 67],
        [67, 71, 74],
        [69, 72, 76],
        [65, 69, 72],
      ].flatMap((pitches, bar) =>
        pitches.map((pitch) => ({ pitch, velocity: 80, start: bar * 1920, duration: 1920 })),
      ),
    },
  ],
  harmony: [],
};

function fakeSynth() {
  const played: { pitch: number; time: number; durationSec: number }[] = [];
  const synth: Synth & { played: typeof played; allOffCalls: number } = {
    played,
    allOffCalls: 0,
    play: (n) => {
      played.push({ pitch: n.pitch, time: n.time, durationSec: n.durationSec });
    },
    allOff: () => {
      synth.allOffCalls += 1;
    },
  };
  return synth;
}

describe('TimelinePlayer', () => {
  // Fake timers drive the scheduler; the clock follows Date.now so both agree.
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(1_000_000);
  });
  afterEach(() => {
    vi.useRealTimers();
  });
  const clock = { now: () => Date.now() / 1000 };

  function newPlayer(synth: Synth) {
    const p = new TimelinePlayer({ synth, clock, lookAheadSec: 0.1, intervalMs: 25 });
    p.load(fourBars);
    return p;
  }

  it('schedules each bar at the right audio time and ends after four bars', () => {
    const synth = fakeSynth();
    const p = newPlayer(synth);
    const t0 = clock.now();

    p.play();
    expect(p.isPlaying).toBe(true);
    // Only the first chord is within the 100 ms look-ahead.
    expect(synth.played.map((n) => n.pitch)).toEqual([60, 64, 67]);
    expect(synth.played[0]?.time).toBe(t0);
    expect(synth.played[0]?.durationSec).toBe(2);

    vi.advanceTimersByTime(2000);
    expect(synth.played.length).toBe(6);
    expect(synth.played[3]?.time).toBeCloseTo(t0 + 2, 6);

    vi.advanceTimersByTime(4000);
    expect(synth.played.length).toBe(12);
    expect(synth.played[9]?.time).toBeCloseTo(t0 + 6, 6);

    expect(p.isPlaying).toBe(true);
    vi.advanceTimersByTime(2100);
    expect(p.isPlaying).toBe(false);
    expect(p.positionTick).toBe(0);
    // A natural end does not cut notes off.
    expect(synth.allOffCalls).toBe(0);
  });

  it('reports position in ticks from the clock', () => {
    const p = newPlayer(fakeSynth());
    p.play();
    expect(p.positionTick).toBe(0);
    vi.advanceTimersByTime(1500); // 1.5 s at 120 BPM = 3 beats
    expect(p.positionTick).toBeCloseTo(1440, 6);
    vi.advanceTimersByTime(2500);
    expect(p.positionTick).toBeCloseTo(3840, 6);
  });

  it('stop silences the synth, rewinds and notifies subscribers', () => {
    const synth = fakeSynth();
    const p = newPlayer(synth);
    const events: [number, string][] = [];
    const unsubscribe = p.subscribe((tick, state) => events.push([tick, state]));

    p.play();
    vi.advanceTimersByTime(1000);
    p.stop();

    expect(p.isPlaying).toBe(false);
    expect(p.positionTick).toBe(0);
    expect(synth.allOffCalls).toBe(1);
    expect(events.at(-1)).toEqual([0, 'stopped']);
    expect(events.some(([tick, state]) => state === 'playing' && tick > 0)).toBe(true);

    unsubscribe();
    p.play();
    expect(events.at(-1)).toEqual([0, 'stopped']);
  });

  it('play is idempotent while playing and a no-op without a timeline', () => {
    const synth = fakeSynth();
    const empty = new TimelinePlayer({ synth, clock });
    empty.play();
    expect(empty.isPlaying).toBe(false);

    const p = newPlayer(synth);
    p.play();
    p.play();
    expect(synth.played.length).toBe(3);
  });

  it('loading a new timeline stops the current playback', () => {
    const synth = fakeSynth();
    const p = newPlayer(synth);
    p.play();
    p.load(fourBars);
    expect(p.isPlaying).toBe(false);
    expect(synth.allOffCalls).toBe(1);
  });
});
