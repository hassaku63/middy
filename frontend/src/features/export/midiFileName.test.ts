import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { midiFileName } from './midiFileName';

const base: Timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [],
};
const h = (romanNumeral: string) => ({
  start: 0,
  duration: 1,
  chord: { root: 0, quality: 'maj' as const },
  romanNumeral,
});

describe('midiFileName', () => {
  it('includes key, mode and the progression', () => {
    expect(midiFileName({ ...base, harmony: ['I', 'V', 'vi', 'IV'].map(h) })).toBe(
      'middy-C-major-I-V-vi-IV.mid',
    );
  });

  it('is filesystem-safe for sharps and decorated numerals', () => {
    expect(
      midiFileName({
        ...base,
        key: { tonic: 6, mode: 'minor' },
        harmony: ['i', 'ii°', 'viiø7', 'V7'].map(h),
      }),
    ).toBe('middy-Fs-minor-i-iidim-viihd7-V7.mid');
  });

  it('keeps borrowed and secondary-dominant numerals in a safe form', () => {
    expect(midiFileName({ ...base, harmony: ['I', 'V7/V', 'V', '♭VII'].map(h) })).toBe(
      'middy-C-major-I-V7ofV-V-bVII.mid',
    );
  });

  it('omits the progression part when there is no harmony', () => {
    expect(midiFileName(base)).toBe('middy-C-major.mid');
  });
});
