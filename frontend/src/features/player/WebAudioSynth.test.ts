import { describe, expect, it } from 'vitest';
import { drumVoice, midiToHz } from './WebAudioSynth';

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
