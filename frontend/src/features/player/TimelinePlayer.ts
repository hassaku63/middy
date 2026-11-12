import type { Timeline } from '@/shared/api/client';
import {
  quarterBpm,
  scheduleNotes,
  secondsToTicks,
  ticksToSeconds,
  timelineLengthTicks,
} from '@/shared/lib/timing';
import type { ScheduledNote } from '@/shared/lib/timing';

/** Something that can sound a note at an absolute audio-clock time. */
export type Synth = {
  play(note: ScheduledNote & { time: number }): void;
  /** Silence everything immediately (stop button). */
  allOff(): void;
};

/** Monotonic clock in seconds; in production this is AudioContext.currentTime. */
export type Clock = { now(): number };

export type PlayerState = 'stopped' | 'playing';

export type PlayerOptions = {
  synth: Synth;
  clock: Clock;
  /** How far ahead of the clock notes are handed to the synth (seconds). */
  lookAheadSec?: number;
  /** Scheduler period in milliseconds. */
  intervalMs?: number;
  setInterval?: IntervalScheduler['set'];
  clearInterval?: IntervalScheduler['clear'];
};

/** Narrow timer signatures (no string-eval overloads) so fakes are easy to type. */
type IntervalScheduler = {
  set: (fn: () => void, ms: number) => TimerHandle;
  clear: (handle: TimerHandle) => void;
};
type TimerHandle = ReturnType<typeof globalThis.setInterval>;

/** GM percussion (MIDI channel 10). */
const DRUM_CHANNEL = 9;

type Listener = (positionTick: number, state: PlayerState) => void;

/**
 * Plays a Timeline with a look-ahead scheduler: a coarse JS timer wakes up
 * every `intervalMs` and hands the synth every note that starts within the
 * next `lookAheadSec`, stamped with a precise audio-clock time. Position is
 * derived from the clock, never from timer counts, so it stays accurate.
 *
 * No Web Audio objects live here; that keeps the class testable with fakes.
 */
export class TimelinePlayer {
  private readonly synth: Synth;
  private readonly clock: Clock;
  private readonly lookAheadSec: number;
  private readonly intervalMs: number;
  private readonly setInterval: IntervalScheduler['set'];
  private readonly clearInterval: IntervalScheduler['clear'];

  private timeline: Timeline | null = null;
  private notes: ScheduledNote[] = [];
  private lengthSec = 0;
  private nextNote = 0;
  /** Audio-clock time of the timeline's tick 0 (in the past after a seek). */
  private startTime = 0;
  /** Where play starts (BL-327): where it paused; 0 after stop or the end (BL-334). */
  private cueTick = 0;
  private timer: TimerHandle | null = null;
  private state: PlayerState = 'stopped';
  private readonly listeners = new Set<Listener>();

  constructor(opts: PlayerOptions) {
    this.synth = opts.synth;
    this.clock = opts.clock;
    this.lookAheadSec = opts.lookAheadSec ?? 0.1;
    this.intervalMs = opts.intervalMs ?? 25;
    this.setInterval = opts.setInterval ?? ((fn, ms) => globalThis.setInterval(fn, ms));
    this.clearInterval = opts.clearInterval ?? ((id) => globalThis.clearInterval(id));
  }

  load(timeline: Timeline): void {
    if (this.state === 'playing') this.halt();
    this.cueTick = 0;
    this.timeline = timeline;
    this.notes = scheduleNotes(timeline);
    this.lengthSec = ticksToSeconds(
      timelineLengthTicks(timeline),
      timeline.ppq,
      quarterBpm(timeline),
    );
  }

  get isPlaying(): boolean {
    return this.state === 'playing';
  }

  /** Current position in ticks; stopped, where the next play starts. */
  get positionTick(): number {
    if (this.state !== 'playing' || !this.timeline) return this.cueTick;
    const elapsed = Math.max(0, this.clock.now() - this.startTime);
    return Math.min(
      secondsToTicks(elapsed, this.timeline.ppq, quarterBpm(this.timeline)),
      timelineLengthTicks(this.timeline),
    );
  }

  /** Starts at fromTick (the beginning by default). No-op without a loaded timeline. */
  play(fromTick = 0): void {
    if (!this.timeline || this.state === 'playing') return;
    this.state = 'playing';
    this.cueTick = this.clamp(fromTick);
    this.startAt(this.cueTick);
    this.tick();
    // Read through the getter: tick() may have finished an empty timeline,
    // which the narrowing from the assignment above cannot see.
    if (!this.isPlaying) return; // no timer to leak
    this.timer = this.setInterval(() => this.tick(), this.intervalMs);
    this.notify();
  }

  /**
   * Moves the position to tick (BL-327). Playing, the sounding notes stop
   * and play goes on from there; stopped, the next play starts there.
   */
  seek(tick: number): void {
    this.cueTick = this.clamp(tick);
    if (this.state === 'playing') {
      this.synth.allOff();
      this.startAt(this.cueTick);
      this.tick();
      return;
    }
    this.notify();
  }

  /** Stops where it is: the next play goes on from there (BL-334). */
  pause(): void {
    if (this.state !== 'playing') return;
    this.cueTick = this.clamp(this.positionTick);
    this.halt();
  }

  /** Stops and goes back to the beginning (BL-334, as a recorder's ■). */
  stop(): void {
    this.cueTick = 0;
    if (this.state === 'stopped') {
      this.notify();
      return;
    }
    this.halt();
  }

  /** Silences everything and reports the stop. */
  private halt(): void {
    this.state = 'stopped';
    if (this.timer !== null) {
      this.clearInterval(this.timer);
      this.timer = null;
    }
    this.synth.allOff();
    this.notify();
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  /** A tick inside the timeline; past its end is the beginning. */
  private clamp(tick: number): number {
    if (!this.timeline) return 0;
    const t = Math.max(0, Math.round(tick));
    return t < timelineLengthTicks(this.timeline) ? t : 0;
  }

  /**
   * Schedules from tick on: notes that start there or later as usual, and
   * the pitched notes already sounding there for what is left of them.
   * Drum hits are not struck again halfway (a late hit sounds as a new one).
   */
  private startAt(tick: number): void {
    const tl = this.timeline;
    if (!tl) return;
    const now = this.clock.now();
    const fromSec = ticksToSeconds(tick, tl.ppq, quarterBpm(tl));
    this.startTime = now - fromSec;
    this.nextNote = this.notes.findIndex((n) => n.startSec >= fromSec);
    if (this.nextNote < 0) this.nextNote = this.notes.length;
    for (const n of this.notes.slice(0, this.nextNote)) {
      const left = n.startSec + n.durationSec - fromSec;
      if (left > 0 && n.channel !== DRUM_CHANNEL) {
        this.synth.play({ ...n, startSec: fromSec, durationSec: left, time: now });
      }
    }
  }

  private tick(): void {
    const now = this.clock.now();
    const horizon = now - this.startTime + this.lookAheadSec;
    while (this.nextNote < this.notes.length) {
      const n = this.notes[this.nextNote];
      if (!n || n.startSec > horizon) break;
      this.synth.play({ ...n, time: this.startTime + n.startSec });
      this.nextNote += 1;
    }
    if (now - this.startTime >= this.lengthSec) {
      this.finish();
      return;
    }
    this.notify();
  }

  /** Natural end: timer stops, nothing is cut off, back to the beginning. */
  private finish(): void {
    this.state = 'stopped';
    this.cueTick = 0;
    if (this.timer !== null) {
      this.clearInterval(this.timer);
      this.timer = null;
    }
    this.notify();
  }

  private notify(): void {
    const pos = this.positionTick;
    for (const l of this.listeners) l(pos, this.state);
  }
}
