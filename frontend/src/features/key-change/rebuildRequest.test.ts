import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { rebuildRequest } from './rebuildRequest';

const base: Timeline = {
  ppq: 480,
  tempoBpm: 72,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 9, mode: 'minor' },
  tracks: [],
  harmony: [
    { start: 0, duration: 1920, chord: { root: 9, quality: 'min' }, romanNumeral: 'i' },
    { start: 1920, duration: 960, chord: { root: 4, quality: 'dom7' }, romanNumeral: 'V7' },
  ],
};

describe('rebuildRequest', () => {
  it('keeps numerals, lengths, tempo and mode, and changes only the tonic', () => {
    expect(rebuildRequest(base, 0)).toEqual({
      key: { tonic: 0, mode: 'minor' },
      tempoBpm: 72,
      progression: [
        { numeral: 'i', beats: 4 },
        { numeral: 'V7', beats: 2 },
      ],
    });
  });

  it('carries the patterns and the feel the timeline was rendered with', () => {
    const grooved: Timeline = {
      ...base,
      rhythm: {
        accompaniment: {
          id: 'arpeggio',
          name: 'アルペジオ',
          description: '',
          hits: 8,
          onBeat: 4,
          offBeat: 4,
          finer: 0,
          syncopated: true,
          explanation: '',
        },
        feel: { id: 'swing', name: 'スウィング', explanation: '' },
      },
    };
    expect(rebuildRequest(grooved, 2)).toMatchObject({
      accompaniment: 'arpeggio',
      feel: 'swing',
    });
    expect(rebuildRequest(grooved, 2)).not.toHaveProperty('drums');
  });
});
