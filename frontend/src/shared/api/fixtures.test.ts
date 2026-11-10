import { describe, expect, it } from 'vitest';
import type { Chord, Key, Timeline } from './client';
import notation from './fixtures/notation.json';
import timelineFixture from './fixtures/timeline.json';
import { chordSymbol, keyName, pitchClassName } from '@/shared/lib/notation';

/**
 * Cross-language parity (ADR-0007): the backend generates these fixtures
 * (`make fixtures`) and this test proves the frontend mirrors spell
 * things exactly the same way. If this fails after a backend change,
 * update notation.ts; if the JSON is stale, the Go test says so.
 */
describe('fixtures from backend/internal/fixtures', () => {
  it('pitch class names match', () => {
    for (const { pitchClass, name } of notation.pitchClasses) {
      expect(pitchClassName(pitchClass)).toBe(name);
    }
    expect(notation.pitchClasses).toHaveLength(12);
  });

  it('key names match for every tonic and mode', () => {
    for (const { key, name } of notation.keyNames) {
      expect(keyName(key as Key)).toBe(name);
    }
    expect(notation.keyNames).toHaveLength(12 * 5); // every tonic in each of the five modes
  });

  it('chord symbols match for every root and quality', () => {
    for (const { chord, symbol } of notation.chords) {
      expect(chordSymbol(chord as Chord)).toBe(symbol);
    }
    // One per root and quality the engine knows (music.Qualities()), so a
    // new quality cannot slip past the mirror in notation.ts.
    const qualities = new Set(notation.chords.map(({ chord }) => (chord as Chord).quality));
    expect(qualities.size).toBeGreaterThanOrEqual(10);
    expect(notation.chords).toHaveLength(12 * qualities.size);
  });

  it('roman numeral fixtures carry diatonic chords for C major and A minor', () => {
    const keys = new Set(
      notation.romanNumerals.map((r) => `${(r.key as Key).tonic}-${(r.key as Key).mode}`),
    );
    expect([...keys].sort()).toEqual(['0-major', '9-minor']);
    const cMajorTriads = notation.romanNumerals
      .filter((r) => (r.key as Key).tonic === 0 && !r.seventh)
      .map((r) => `${r.romanNumeral}=${chordSymbol(r.chord as Chord)}`);
    expect(cMajorTriads).toEqual(['I=C', 'ii=Dm', 'iii=Em', 'IV=F', 'V=G', 'vi=Am', 'vii°=Bdim']);
  });

  it('the analysed timeline fixture satisfies the Timeline type and the docs example', () => {
    const tl = timelineFixture as Timeline;
    expect(tl.key).toEqual({ tonic: 0, mode: 'major' });
    expect(tl.tempoBpm).toBe(90);
    expect(tl.harmony.map((h) => h.romanNumeral)).toEqual(['I', 'V', 'vi', 'IV']);
    expect(tl.harmony.map((h) => h.analysis?.function)).toEqual(['T', 'D', 'T', 'S']);
    // One phrase that ends on IV: V and vi inside it close nothing (#371).
    expect(tl.harmony.map((h) => h.analysis?.cadence)).toEqual([
      undefined,
      undefined,
      undefined,
      undefined,
    ]);
  });
});
