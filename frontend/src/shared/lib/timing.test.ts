import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import {
  formatPosition,
  scheduleNotes,
  secondsToTicks,
  ticksPerBar,
  ticksToSeconds,
  timelineLengthTicks,
} from './timing';

const sig = { numerator: 4, denominator: 4 };

describe('tick/second conversions', () => {
  it('one quarter note at 120 BPM is half a second', () => {
    expect(ticksToSeconds(480, 480, 120)).toBe(0.5);
    expect(secondsToTicks(0.5, 480, 120)).toBe(480);
  });

  it('four bars at 90 BPM take 10.67 seconds', () => {
    expect(ticksToSeconds(4 * 4 * 480, 480, 90)).toBeCloseTo(10.667, 3);
  });

  it('knows the bar length for common signatures', () => {
    expect(ticksPerBar(480, sig)).toBe(1920);
    expect(ticksPerBar(480, { numerator: 3, denominator: 4 })).toBe(1440);
    expect(ticksPerBar(480, { numerator: 6, denominator: 8 })).toBe(1440);
  });
});

describe('formatPosition', () => {
  it('is 1-based bar.beat', () => {
    expect(formatPosition(0, 480, sig)).toBe('1.1');
    expect(formatPosition(480, 480, sig)).toBe('1.2');
    expect(formatPosition(1919, 480, sig)).toBe('1.4');
    expect(formatPosition(1920, 480, sig)).toBe('2.1');
    expect(formatPosition(1440, 480, { numerator: 3, denominator: 4 })).toBe('2.1');
  });
});

describe('scheduleNotes', () => {
  const tl: Timeline = {
    ppq: 480,
    tempoBpm: 120,
    timeSignature: sig,
    key: { tonic: 0, mode: 'major' },
    tracks: [
      {
        name: 'a',
        channel: 0,
        program: 0,
        notes: [{ pitch: 64, velocity: 80, start: 960, duration: 480 }],
      },
      {
        name: 'b',
        channel: 1,
        program: 0,
        notes: [{ pitch: 60, velocity: 80, start: 0, duration: 1920 }],
      },
    ],
    harmony: [{ start: 0, duration: 3840, chord: { root: 0, quality: 'maj' }, romanNumeral: 'I' }],
  };

  it('flattens tracks and orders by start time in seconds', () => {
    expect(scheduleNotes(tl)).toEqual([
      { pitch: 60, velocity: 80, startSec: 0, durationSec: 2 },
      { pitch: 64, velocity: 80, startSec: 1, durationSec: 0.5 },
    ]);
  });

  it('length covers harmony annotations too', () => {
    expect(timelineLengthTicks(tl)).toBe(3840);
  });
});
