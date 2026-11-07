import type { Synth } from './TimelinePlayer';
import type { ScheduledNote } from '@/shared/lib/timing';

/** GM percussion channel (0-based). */
export const DRUM_CHANNEL = 9;

/** MIDI note number to frequency in Hz (A4 = 69 = 440 Hz). */
export function midiToHz(pitch: number): number {
  return 440 * 2 ** ((pitch - 69) / 12);
}

/**
 * A small subtractive "piano-ish" voice: two detuned oscillators through a
 * low-pass filter with a percussive envelope. Good enough to hear chords
 * clearly; sample-based instruments are a backlog idea, not a requirement.
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
    const peak = (note.velocity / 127) ** 1.5;
    const end = time + durationSec;

    const gain = this.ctx.createGain();
    const filter = this.ctx.createBiquadFilter();
    filter.type = 'lowpass';
    filter.frequency.setValueAtTime(Math.min(8000, midiToHz(note.pitch) * 6), time);
    filter.frequency.exponentialRampToValueAtTime(Math.max(200, midiToHz(note.pitch) * 2), end);

    gain.gain.setValueAtTime(0, time);
    gain.gain.linearRampToValueAtTime(peak, time + 0.01);
    gain.gain.exponentialRampToValueAtTime(Math.max(peak * 0.35, 0.0001), time + 0.25);
    gain.gain.setValueAtTime(Math.max(peak * 0.35, 0.0001), end);
    gain.gain.exponentialRampToValueAtTime(0.0001, end + 0.15);

    const oscs = [
      { type: 'triangle' as const, detune: 0, level: 1 },
      { type: 'sine' as const, detune: 4, level: 0.5 },
    ].map(({ type, detune, level }) => {
      const osc = this.ctx.createOscillator();
      osc.type = type;
      osc.frequency.value = midiToHz(note.pitch);
      osc.detune.value = detune;
      const level_ = this.ctx.createGain();
      level_.gain.value = level;
      osc.connect(level_).connect(filter);
      osc.start(time);
      osc.stop(end + 0.2);
      return osc;
    });
    filter.connect(gain).connect(this.master);

    const voice = { oscs, gain };
    this.active.add(voice);
    oscs[0]?.addEventListener('ended', () => this.active.delete(voice));
  }

  /**
   * GM percussion: kick = sine pitch drop, snare = noise burst with a
   * short tone, everything else = short high-passed noise (hats, bells).
   */
  private playDrum(note: ScheduledNote & { time: number }): void {
    const { time } = note;
    const level = (note.velocity / 127) ** 1.2;
    const gain = this.ctx.createGain();
    gain.connect(this.master);

    if (note.pitch === 35 || note.pitch === 36) {
      const osc = this.ctx.createOscillator();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(150, time);
      osc.frequency.exponentialRampToValueAtTime(45, time + 0.12);
      gain.gain.setValueAtTime(level, time);
      gain.gain.exponentialRampToValueAtTime(0.0001, time + 0.25);
      osc.connect(gain);
      osc.start(time);
      osc.stop(time + 0.3);
      return;
    }

    const isSnare = note.pitch === 38 || note.pitch === 40;
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
