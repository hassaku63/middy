import { describe, expect, it } from 'vitest';
import { parseProgressionText } from './parseProgressionText';

describe('parseProgressionText', () => {
  it('accepts dashes, spaces, commas, pipes and arrows as separators', () => {
    const want = [
      { numeral: 'I', beats: 4 },
      { numeral: 'V', beats: 4 },
      { numeral: 'vi', beats: 4 },
      { numeral: 'IV', beats: 4 },
    ];
    expect(parseProgressionText('I-V-vi-IV', 4)).toEqual(want);
    expect(parseProgressionText('I - V - vi - IV', 4)).toEqual(want);
    expect(parseProgressionText('I, V, vi, IV', 4)).toEqual(want);
    expect(parseProgressionText('I | V | vi | IV', 4)).toEqual(want);
    expect(parseProgressionText('I → V → vi → IV', 4)).toEqual(want);
  });

  it('keeps unknown tokens for the backend to reject', () => {
    expect(parseProgressionText('I X', 2)).toEqual([
      { numeral: 'I', beats: 2 },
      { numeral: 'X', beats: 2 },
    ]);
  });

  it('returns nothing for blank input', () => {
    expect(parseProgressionText('   ', 4)).toEqual([]);
  });
});
