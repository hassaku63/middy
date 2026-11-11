import { describe, expect, it } from 'vitest';
import { chordSymbol, harmonySymbol, pitchClassName, pitchName } from './notation';

describe('notation', () => {
  it('names pitch classes with sharps and wraps', () => {
    expect(pitchClassName(0)).toBe('C');
    expect(pitchClassName(1)).toBe('C#');
    expect(pitchClassName(11)).toBe('B');
    expect(pitchClassName(12)).toBe('C');
  });

  it('prints scientific pitch names', () => {
    expect(pitchName(60)).toBe('C4');
    expect(pitchName(69)).toBe('A4');
    expect(pitchName(0)).toBe('C-1');
  });

  it('prints chord symbols like the backend', () => {
    expect(chordSymbol({ root: 7, quality: 'dom7' })).toBe('G7');
    expect(chordSymbol({ root: 9, quality: 'min' })).toBe('Am');
    expect(chordSymbol({ root: 11, quality: 'm7b5' })).toBe('Bm7b5');
    expect(chordSymbol({ root: 0, quality: 'maj' })).toBe('C');
  });
});

describe('harmonySymbol with tensions (BL-307)', () => {
  const chord = { root: 5, quality: 'maj7' as const };
  it('adds the sounding tensions in brackets, before a slash bass', () => {
    expect(harmonySymbol({ chord, symbol: 'Fmaj7' })).toBe('Fmaj7');
    expect(harmonySymbol({ chord, symbol: 'Fmaj7', tensions: ['9'] })).toBe('Fmaj7(9)');
    expect(harmonySymbol({ chord, symbol: 'C7', tensions: ['9', '13'] })).toBe('C7(9,13)');
    expect(harmonySymbol({ chord, symbol: 'Cmaj7/E', tensions: ['9'] })).toBe('Cmaj7(9)/E');
  });
});
