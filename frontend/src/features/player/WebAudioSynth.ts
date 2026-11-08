import type { Synth } from './TimelinePlayer';
import type { ScheduledNote } from '@/shared/lib/timing';

/** GM percussion channel (0-based). */
export const DRUM_CHANNEL = 9;

/** MIDI note number to frequency in Hz (A4 = 69 = 440 Hz). */
export function midiToHz(pitch: number): number {
  return 440 * 2 ** ((pitch - 69) / 12);
}

/** How the synth voices a GM percussion note. */
export type DrumVoice =
  | { kind: 'kick' }
  | { kind: 'tom'; startHz: number; endHz: number }
  | { kind: 'stick' }
  | { kind: 'snare' }
  | { kind: 'hat' };

/**
 * The voice for a GM drum note: kicks and toms are pitched membranes (a
 * sine that drops), the side stick a short click, the snare a noise burst
 * with a tone, and everything else a short bright noise (hats, bells). The
 * toms carry the bodhrán-like pattern, so they must not sound like hats.
 */
export function drumVoice(pitch: number): DrumVoice {
  if (pitch === 35 || pitch === 36) return { kind: 'kick' };
  if (pitch === 37) return { kind: 'stick' };
  if (pitch === 38 || pitch === 40) return { kind: 'snare' };
  // GM toms: 41 low floor .. 50 high tom; higher note, higher drum.
  if ([41, 43, 45, 47, 48, 50].includes(pitch)) {
    const startHz = 80 + (pitch - 41) * 14;
    return { kind: 'tom', startHz, endHz: startHz * 0.6 };
  }
  return { kind: 'hat' };
}

/** The instrument families the synth tells apart (BL-110); not a realistic rendering. */
export type InstrumentFamily = 'keys' | 'plucked' | 'organ' | 'strings' | 'flute';

type Wave = { type: OscillatorType; detune: number; level: number };

/** How a family sounds: oscillators, attack (s), the level it decays to, and the decay time (s). */
export type InstrumentVoice = {
  family: InstrumentFamily;
  waves: Wave[];
  attack: number;
  sustain: number;
  decay: number;
  /** Low-pass cutoff as a multiple of the note's frequency at the start. */
  brightness: number;
};

const VOICES: Record<InstrumentFamily, Omit<InstrumentVoice, 'family'>> = {
  keys: {
    waves: [
      { type: 'triangle', detune: 0, level: 1 },
      { type: 'sine', detune: 4, level: 0.5 },
    ],
    attack: 0.01,
    sustain: 0.35,
    decay: 0.25,
    brightness: 6,
  },
  plucked: {
    waves: [
      { type: 'sawtooth', detune: 0, level: 0.6 },
      { type: 'triangle', detune: -3, level: 0.6 },
    ],
    attack: 0.005,
    sustain: 0.08,
    decay: 0.45,
    brightness: 8,
  },
  organ: {
    waves: [
      { type: 'square', detune: 0, level: 0.35 },
      { type: 'sine', detune: 1200, level: 0.4 },
    ],
    attack: 0.03,
    sustain: 0.85,
    decay: 0.1,
    brightness: 5,
  },
  strings: {
    waves: [
      { type: 'sawtooth', detune: -6, level: 0.45 },
      { type: 'sawtooth', detune: 6, level: 0.45 },
    ],
    attack: 0.15,
    sustain: 0.8,
    decay: 0.2,
    brightness: 4,
  },
  flute: {
    waves: [
      { type: 'sine', detune: 0, level: 1 },
      { type: 'triangle', detune: 1200, level: 0.15 },
    ],
    attack: 0.06,
    sustain: 0.85,
    decay: 0.1,
    brightness: 3,
  },
};

/** The family of a GM program (0-based): pianos, guitars and the harp, organs, strings, pipes. */
export function instrumentVoice(program: number): InstrumentVoice {
  let family: InstrumentFamily = 'keys';
  if ((program >= 24 && program <= 31) || program === 46 || (program >= 104 && program <= 107)) {
    family = 'plucked';
  } else if (program >= 16 && program <= 23) {
    family = 'organ';
  } else if ((program >= 40 && program <= 45) || (program >= 48 && program <= 51)) {
    family = 'strings';
  } else if (program >= 72 && program <= 79) {
    family = 'flute';
  }
  return { family, ...VOICES[family] };
}

/**
 * A small subtractive synth: a couple of oscillators through a low-pass
 * filter, shaped per instrument family (instrumentVoice). Good enough to
 * hear a guitar from an organ; sample-based instruments are a backlog
 * idea, not a requirement.
 */
export class WebAudioSynth implements Synth {
  private readonly master: GainNode;
  private active = new Set<{ oscs: OscillatorNode[]; gain: GainNode }>();

  constructor(private readonly ctx: AudioContext) {
    this.master = ctx.createGain();
    this.master.gain.value = 0.25;
    this.master.connect(ctx.destination);
  }

  play(note: ScheduledNote & { time: number }): void {
    if (note.channel === DRUM_CHANNEL) {
      this.playDrum(note);
      return;
    }
    const { time, durationSec } = note;
    const inst = instrumentVoice(note.program);
    const peak = (note.velocity / 127) ** 1.5;
    const end = time + durationSec;
    const held = Math.max(peak * inst.sustain, 0.0001);
    const release = Math.max(end, time + inst.attack + inst.decay);

    const gain = this.ctx.createGain();
    const filter = this.ctx.createBiquadFilter();
    filter.type = 'lowpass';
    filter.frequency.setValueAtTime(Math.min(8000, midiToHz(note.pitch) * inst.brightness), time);
    filter.frequency.exponentialRampToValueAtTime(Math.max(200, midiToHz(note.pitch) * 2), end);

    gain.gain.setValueAtTime(0, time);
    gain.gain.linearRampToValueAtTime(peak, time + inst.attack);
    gain.gain.exponentialRampToValueAtTime(held, time + inst.attack + inst.decay);
    gain.gain.setValueAtTime(held, release);
    gain.gain.exponentialRampToValueAtTime(0.0001, release + 0.15);

    const oscs = inst.waves.map(({ type, detune, level }) => {
      const osc = this.ctx.createOscillator();
      osc.type = type;
      osc.frequency.value = midiToHz(note.pitch);
      osc.detune.value = detune;
      const level_ = this.ctx.createGain();
      level_.gain.value = level;
      osc.connect(level_).connect(filter);
      osc.start(time);
      osc.stop(release + 0.2);
      return osc;
    });
    filter.connect(gain).connect(this.master);

    const voice = { oscs, gain };
    this.active.add(voice);
    oscs[0]?.addEventListener('ended', () => this.active.delete(voice));
  }

  /** GM percussion, voiced by drumVoice. */
  private playDrum(note: ScheduledNote & { time: number }): void {
    const { time } = note;
    const level = (note.velocity / 127) ** 1.2;
    const gain = this.ctx.createGain();
    gain.connect(this.master);
    const voice = drumVoice(note.pitch);
    switch (voice.kind) {
      case 'kick':
        this.membrane(gain, time, level, 150, 45, 0.25);
        return;
      case 'tom':
        this.membrane(gain, time, level * 0.9, voice.startHz, voice.endHz, 0.3);
        return;
      case 'stick':
        this.click(gain, time, level);
        return;
      case 'snare':
      case 'hat':
        this.noiseHit(gain, time, level, voice.kind === 'snare');
    }
  }

  /** A pitched drum head: a sine falling from startHz to endHz. */
  private membrane(
    gain: GainNode,
    time: number,
    level: number,
    startHz: number,
    endHz: number,
    length: number,
  ): void {
    const osc = this.ctx.createOscillator();
    osc.type = 'sine';
    osc.frequency.setValueAtTime(startHz, time);
    osc.frequency.exponentialRampToValueAtTime(endHz, time + length * 0.5);
    gain.gain.setValueAtTime(level, time);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + length);
    osc.connect(gain);
    osc.start(time);
    osc.stop(time + length + 0.05);
  }

  /** A side stick: a very short, woody tone with a touch of noise. */
  private click(gain: GainNode, time: number, level: number): void {
    const osc = this.ctx.createOscillator();
    osc.type = 'triangle';
    osc.frequency.setValueAtTime(1100, time);
    gain.gain.setValueAtTime(level * 0.6, time);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + 0.04);
    osc.connect(gain);
    osc.start(time);
    osc.stop(time + 0.06);
  }

  /** A snare (band-passed noise and a tone) or a hat (short high-passed noise). */
  private noiseHit(gain: GainNode, time: number, level: number, isSnare: boolean): void {
    const length = isSnare ? 0.18 : 0.06;
    const src = this.ctx.createBufferSource();
    src.buffer = this.noise();
    const filter = this.ctx.createBiquadFilter();
    filter.type = isSnare ? 'bandpass' : 'highpass';
    filter.frequency.value = isSnare ? 1800 : 7000;
    gain.gain.setValueAtTime(level * (isSnare ? 0.8 : 0.35), time);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + length);
    src.connect(filter).connect(gain);
    src.start(time);
    src.stop(time + length + 0.02);
    if (isSnare) {
      const tone = this.ctx.createOscillator();
      const toneGain = this.ctx.createGain();
      tone.type = 'triangle';
      tone.frequency.setValueAtTime(190, time);
      toneGain.gain.setValueAtTime(level * 0.4, time);
      toneGain.gain.exponentialRampToValueAtTime(0.0001, time + 0.1);
      tone.connect(toneGain).connect(this.master);
      tone.start(time);
      tone.stop(time + 0.12);
    }
  }

  private noiseBuffer: AudioBuffer | null = null;

  /** One second of white noise, reused for every hit. */
  private noise(): AudioBuffer {
    if (!this.noiseBuffer) {
      const rate = this.ctx.sampleRate;
      const buf = this.ctx.createBuffer(1, rate, rate);
      const data = buf.getChannelData(0);
      for (let i = 0; i < data.length; i++) data[i] = Math.random() * 2 - 1;
      this.noiseBuffer = buf;
    }
    return this.noiseBuffer;
  }

  allOff(): void {
    const now = this.ctx.currentTime;
    for (const voice of this.active) {
      voice.gain.gain.cancelScheduledValues(now);
      voice.gain.gain.setValueAtTime(voice.gain.gain.value, now);
      voice.gain.gain.linearRampToValueAtTime(0, now + 0.03);
      for (const osc of voice.oscs) osc.stop(now + 0.05);
    }
    this.active.clear();
  }
}
