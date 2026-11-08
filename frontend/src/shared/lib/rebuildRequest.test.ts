import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { grooveFor, rebuildRequest, regrooveRequest } from './rebuildRequest';

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
  it('keeps numerals, lengths, tempo, meter and mode, and changes only the tonic', () => {
    expect(rebuildRequest(base, 0)).toEqual({
      key: { tonic: 0, mode: 'minor' },
      tempoBpm: 72,
      timeSignature: { numerator: 4, denominator: 4 },
      progression: [
        { numeral: 'i', beats: 4 },
        { numeral: 'V7', beats: 2 },
      ],
    });
  });

  it('keeps a waltz in 3/4 (ADR-0017)', () => {
    const waltz: Timeline = { ...base, timeSignature: { numerator: 3, denominator: 4 } };
    expect(rebuildRequest(waltz, 5).timeSignature).toEqual({ numerator: 3, denominator: 4 });
  });

  it('counts chord lengths in dotted-quarter beats in 6/8', () => {
    const sixEight: Timeline = {
      ...base,
      timeSignature: { numerator: 6, denominator: 8 },
      harmony: [
        { start: 0, duration: 1440, chord: { root: 9, quality: 'min' }, romanNumeral: 'i' },
        { start: 1440, duration: 720, chord: { root: 4, quality: 'dom7' }, romanNumeral: 'V7' },
      ],
    };
    expect(rebuildRequest(sixEight, 0).progression).toEqual([
      { numeral: 'i', beats: 2 },
      { numeral: 'V7', beats: 1 },
    ]);
  });

  it('carries the style, so the key change keeps its instrument (BL-110)', () => {
    const styled: Timeline = {
      ...base,
      rhythm: { style: { id: 'celtic', name: 'ケルト', instrument: 'ハープ' } },
    };
    expect(rebuildRequest(styled, 2).style).toBe('celtic');
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

describe('regrooveRequest (BL-087)', () => {
  const jazz = {
    id: 'jazz',
    name: 'ジャズ',
    description: '',
    accompaniments: ['comp'],
    drums: ['jazz-ride'],
    feels: ['swing' as const],
    grooves: [
      {
        meter: { numerator: 4, denominator: 4 },
        accompaniment: 'comp',
        drums: 'jazz-ride',
        feel: 'swing' as const,
      },
      {
        meter: { numerator: 6, denominator: 8 },
        accompaniment: 'block',
        feel: 'straight' as const,
      },
    ],
  };

  it("finds the style's groove for the timeline's meter", () => {
    expect(grooveFor(jazz, base)?.accompaniment).toBe('comp');
    const waltz: Timeline = { ...base, timeSignature: { numerator: 3, denominator: 4 } };
    expect(grooveFor(jazz, waltz)).toBeUndefined();
  });

  it('keeps the harmony, key and tempo and swaps only the groove', () => {
    const celtic: Timeline = {
      ...base,
      rhythm: { style: { id: 'celtic', name: 'ケルト', instrument: 'ハープ' } },
    };
    const groove = grooveFor(jazz, celtic);
    if (!groove) throw new Error('no groove');
    expect(regrooveRequest(celtic, 'jazz', groove)).toEqual({
      ...rebuildRequest(base, 9),
      style: 'jazz',
      accompaniment: 'comp',
      drums: 'jazz-ride',
      feel: 'swing',
    });
  });

  it('leaves the drums out when the style has none for the meter', () => {
    const sixEight: Timeline = { ...base, timeSignature: { numerator: 6, denominator: 8 } };
    const groove = grooveFor(jazz, sixEight);
    if (!groove) throw new Error('no groove');
    expect(regrooveRequest(sixEight, 'jazz', groove)).not.toHaveProperty('drums');
  });
});

describe('pushes survive a rebuild (BL-085)', () => {
  it('sends a pushed chord back with its push', () => {
    const pushed: Timeline = {
      ...base,
      harmony: [base.harmony[0]!, { ...base.harmony[1]!, push: 'eighth' }],
    };
    expect(rebuildRequest(pushed, 0).progression).toEqual([
      { numeral: 'i', beats: 4 },
      { numeral: 'V7', beats: 2, push: 'eighth' },
    ]);
  });
});

describe('a drum grid survives a rebuild (BL-123)', () => {
  it('sends the grid back instead of the drums id', () => {
    const drumGrid = { division: 2 as const, voices: { kick: 'x... x...', snare: '..x. ..x.' } };
    const gridded: Timeline = {
      ...base,
      rhythm: {
        drums: {
          id: 'grid',
          name: '格子で書いたグルーヴ',
          description: '',
          hits: 4,
          onBeat: 4,
          offBeat: 0,
          finer: 0,
          syncopated: false,
          explanation: '',
        },
        drumGrid,
      },
    };
    const req = rebuildRequest(gridded, 2);
    expect(req.drumGrid).toEqual(drumGrid);
    expect(req).not.toHaveProperty('drums');
  });
});
