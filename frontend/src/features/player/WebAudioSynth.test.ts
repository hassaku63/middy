import { describe, expect, it } from 'vitest';
import { drumVoice, instrumentVoice, midiToHz } from './WebAudioSynth';

describe('midiToHz', () => {
  it('maps A4 to 440 Hz and C4 to 261.63 Hz', () => {
    expect(midiToHz(69)).toBe(440);
    expect(midiToHz(60)).toBeCloseTo(261.63, 2);
    expect(midiToHz(81)).toBe(880);
  });
});

describe('drumVoice', () => {
  it('voices the bodhrán-like strokes as drums, not as hats', () => {
    expect(drumVoice(45).kind).toBe('tom'); // low tom: the accents on 1 and 3
    expect(drumVoice(47).kind).toBe('tom'); // low-mid tom: the light eighths
    expect(drumVoice(37).kind).toBe('stick'); // side stick on 2 and 4
  });

  it('keeps the kit as it was', () => {
    expect(drumVoice(36).kind).toBe('kick');
    expect(drumVoice(38).kind).toBe('snare');
    expect(drumVoice(42).kind).toBe('hat');
    expect(drumVoice(53).kind).toBe('hat'); // ride bell
  });

  it('pitches a higher tom higher', () => {
    const low = drumVoice(45);
    const mid = drumVoice(47);
    if (low.kind !== 'tom' || mid.kind !== 'tom') throw new Error('toms expected');
    expect(mid.startHz).toBeGreaterThan(low.startHz);
    expect(low.endHz).toBeLessThan(low.startHz);
  });
});

describe('instrumentVoice (BL-110)', () => {
  it('sorts GM programs into families the synth can tell apart', () => {
    const family = (p: number) => instrumentVoice(p).family;
    expect([0, 4].map(family)).toEqual(['keys', 'keys']);
    expect([24, 25, 46].map(family)).toEqual(['plucked', 'plucked', 'plucked']); // guitars, harp
    expect([19, 20, 21].map(family)).toEqual(['organ', 'organ', 'organ']); // organs, accordion
    expect([48, 40].map(family)).toEqual(['strings', 'strings']);
    expect([73, 78].map(family)).toEqual(['flute', 'flute']);
    expect(family(120)).toBe('keys'); // anything else falls back to keys
  });

  it('gives each family its own shape: plucks die away, organs hold', () => {
    const pluck = instrumentVoice(25);
    const organ = instrumentVoice(20);
    expect(pluck.sustain).toBeLessThan(organ.sustain);
    expect(pluck.waves.map((w) => w.type)).not.toEqual(organ.waves.map((w) => w.type));
    expect(instrumentVoice(48).attack).toBeGreaterThan(instrumentVoice(0).attack); // strings swell
  });
});
