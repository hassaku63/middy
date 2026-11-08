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
  private startTime = 0;
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
    this.stop();
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

  /** Current position in ticks; 0 when stopped. */
  get positionTick(): number {
    if (this.state !== 'playing' || !this.timeline) return 0;
    const elapsed = Math.max(0, this.clock.now() - this.startTime);
    return Math.min(
      secondsToTicks(elapsed, this.timeline.ppq, quarterBpm(this.timeline)),
      timelineLengthTicks(this.timeline),
    );
  }

  /** Starts from the beginning. No-op without a loaded timeline. */
  play(): void {
    if (!this.timeline || this.state === 'playing') return;
    this.state = 'playing';
    this.nextNote = 0;
    this.startTime = this.clock.now();
    this.tick();
    this.timer = this.setInterval(() => this.tick(), this.intervalMs);
    this.notify();
  }

  /** Stops immediately and rewinds. */
  stop(): void {
    if (this.state === 'stopped') return;
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

  /** Natural end: timer stops, nothing is cut off, position resets. */
  private finish(): void {
    this.state = 'stopped';
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
