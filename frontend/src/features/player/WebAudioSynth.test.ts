import { describe, expect, it } from 'vitest';
import { midiToHz } from './WebAudioSynth';

describe('midiToHz', () => {
  it('maps A4 to 440 Hz and C4 to 261.63 Hz', () => {
    expect(midiToHz(69)).toBe(440);
    expect(midiToHz(60)).toBeCloseTo(261.63, 2);
    expect(midiToHz(81)).toBe(880);
  });
});
