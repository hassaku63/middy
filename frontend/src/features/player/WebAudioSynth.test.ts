import { describe, expect, it } from 'vitest';
import { drumVoice, envelope, instrumentVoice, midiToHz, WebAudioSynth } from './WebAudioSynth';

describe('midiToHz', () => {
  it('maps A4 to 440 Hz and C4 to 261.63 Hz', () => {
    expect(midiToHz(69)).toBe(440);
    expect(midiToHz(60)).toBeCloseTo(261.63, 2);
    expect(midiToHz(81)).toBe(880);
  });
});

describe('drumVoice', () => {
  it('voices the metals apart, as the lanes name them (#404 S4)', () => {
    const metal = (pitch: number) => {
      const v = drumVoice(pitch);
      if (v.kind !== 'hat') throw new Error(`${pitch} is not a metal`);
      return v;
    };
    const closed = metal(42);
    expect(metal(49).length).toBeGreaterThan(1); // the crash rings
    expect(metal(46).length).toBeGreaterThan(closed.length * 3); // the open hat rings on
    expect(metal(44).dark).toBe(true); // the foot's hat is dull
    expect(metal(44).length).toBeLessThanOrEqual(closed.length);
    expect(metal(51).ping).toBeGreaterThan(0); // the ride pings
    expect(metal(53).ping).toBeGreaterThan(0); // and its bell
    expect(metal(51)).not.toEqual(closed);
    expect(metal(81)).toEqual(closed); // anything else: a closed hat
  });

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

describe('envelope (review of BL-120)', () => {
  it('releases a short stab at its end instead of holding it through the decay', () => {
    const keys = instrumentVoice(0);
    const stab = envelope(keys, 1, 0.12, 0.5);
    expect(stab.release).toBeCloseTo(1.12);
    expect(stab.decayEnd).toBeCloseTo(1.12);
    const long = envelope(keys, 1, 2, 0.5);
    expect(long.decayEnd).toBeCloseTo(1 + keys.attack + keys.decay);
    expect(long.release).toBeCloseTo(3);
  });

  it('never cuts the attack short', () => {
    const strings = instrumentVoice(48);
    expect(envelope(strings, 0, 0.05, 0.5).release).toBeCloseTo(strings.attack);
  });

  it('releases a note that never decayed from its peak, without a step (review)', () => {
    const strings = instrumentVoice(48);
    const short = envelope(strings, 0, 0.005, 0.5);
    expect(short.atRelease).toBe(0.5);
    const long = envelope(strings, 0, 3, 0.5);
    expect(long.atRelease).toBe(long.held);
  });
});

/** Just enough of an AudioContext to record what allOff() schedules. */
function fakeContext() {
  const param = (value = 1) => ({
    value,
    events: [] as [string, number, number][],
    setValueAtTime(v: number, t: number) {
      this.events.push(['set', v, t]);
    },
    linearRampToValueAtTime(v: number, t: number) {
      this.events.push(['linear', v, t]);
    },
    exponentialRampToValueAtTime(v: number, t: number) {
      this.events.push(['exp', v, t]);
    },
    cancelScheduledValues(t: number) {
      this.events.push(['cancel', 0, t]);
    },
  });
  const node = () => ({ connect: (n: unknown) => n });
  const oscs: { stops: number[] }[] = [];
  const ctx = {
    currentTime: 10,
    sampleRate: 8000,
    destination: node(),
    createGain: () => ({ ...node(), gain: param() }),
    createBiquadFilter: () => ({ ...node(), type: '', frequency: param() }),
    createOscillator: () => {
      const osc = {
        ...node(),
        type: '',
        frequency: param(),
        detune: param(),
        stops: [] as number[],
        start: () => undefined,
        stop(t: number) {
          this.stops.push(t);
        },
        addEventListener: () => undefined,
      };
      oscs.push(osc);
      return osc;
    },
  };
  return { ctx: ctx as unknown as AudioContext, oscs };
}

describe('WebAudioSynth.allOff (review)', () => {
  it('silences a note scheduled ahead at once instead of ramping it down from full gain', () => {
    const { ctx, oscs } = fakeContext();
    const synth = new WebAudioSynth(ctx);
    synth.play({
      pitch: 60,
      velocity: 100,
      time: 10.02,
      startSec: 0.02,
      durationSec: 0.5,
      channel: 0,
      program: 0,
    });
    const before = oscs.length;
    synth.allOff();
    expect(before).toBeGreaterThan(0);
    for (const osc of oscs) expect(osc.stops.at(-1)).toBe(10); // never starts
  });
});
