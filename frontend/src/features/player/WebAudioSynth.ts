import type { Synth } from './TimelinePlayer';
import type { ScheduledNote } from '@/shared/lib/timing';

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
