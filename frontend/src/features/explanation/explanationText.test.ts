import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { explanationLines } from './explanationText';

const timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [
    { romanNumeral: 'I', chord: { root: 0, quality: 'maj' }, start: 0, duration: 1920 },
    { romanNumeral: 'V', chord: { root: 7, quality: 'dom7' }, start: 1920, duration: 1920 },
  ],
  overview: { whole: '長調の明るさ。', flow: 'I から V へ。', degrees: [] },
  rhythm: {
    accompaniment: {
      id: 'arpeggio',
      name: 'アルペジオ',
      description: '分散和音。',
      explanation: '流れる。',
    },
    feel: { id: 'swing', name: 'スウィング', explanation: '跳ねる。' },
  },
} as unknown as Timeline;

describe('explanationLines (BL-300)', () => {
  it('writes the chords, the whole, the flow and the rhythm as lines', () => {
    expect(explanationLines(timeline)).toEqual([
      'コード: C - G7',
      '全体: 長調の明るさ。',
      '流れ: I から V へ。',
      '伴奏: アルペジオ。分散和音。流れる。',
      'ノリ: スウィング。跳ねる。',
    ]);
  });

  it('writes nothing for an empty example', () => {
    const bare: Timeline = { ...timeline, harmony: [] };
    delete bare.overview;
    delete bare.rhythm;
    expect(explanationLines(bare)).toEqual([]);
  });
});
