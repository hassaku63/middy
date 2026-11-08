/**
 * Pure tick/time conversions shared by the player and the timeline view.
 * Ticks are the backend's unit (see docs/architecture/domain-model.md);
 * seconds are Web Audio's.
 */
import type { TimeSignature, Timeline } from '@/shared/api/client';

/** Seconds for ticks at a tempo in quarter notes per minute (see quarterBpm). */
export function ticksToSeconds(tick: number, ppq: number, tempoBpm: number): number {
  return (tick / ppq) * (60 / tempoBpm);
}

export function secondsToTicks(seconds: number, ppq: number, tempoBpm: number): number {
  return (seconds / (60 / tempoBpm)) * ppq;
}

/** Ticks in one bar for the given signature. */
export function ticksPerBar(ppq: number, sig: TimeSignature): number {
  return ppq * 4 * (sig.numerator / sig.denominator);
}

/** Whether the signature is compound (6/8, 9/8, 12/8): its beat is a dotted quarter. */
function isCompound(sig: TimeSignature): boolean {
  return sig.denominator === 8 && sig.numerator > 3 && sig.numerator % 3 === 0;
}

/** Beats a musician counts in one bar: 6/8 is counted in two (music.BeatsPerBar). */
export function beatsPerBar(sig: TimeSignature): number {
  return isCompound(sig) ? sig.numerator / 3 : sig.numerator;
}

/** Ticks in one beat of the meter: a quarter in 4/4 and 3/4, a dotted quarter in 6/8. */
export function ticksPerBeat(ppq: number, sig: TimeSignature): number {
  return ticksPerBar(ppq, sig) / beatsPerBar(sig);
}

/**
 * The timeline's tempo in quarter notes per minute. `tempoBpm` counts the
 * meter's beat (ADR-0019), so 60 in 6/8 is 90 quarters a minute.
 */
export function quarterBpm(tl: Pick<Timeline, 'ppq' | 'tempoBpm' | 'timeSignature'>): number {
  return (tl.tempoBpm * ticksPerBeat(tl.ppq, tl.timeSignature)) / tl.ppq;
}

/** 1-based "bar.beat" position, e.g. tick 0 -> "1.1", one quarter later in 4/4 -> "1.2". */
export function formatPosition(tick: number, ppq: number, sig: TimeSignature): string {
  const perBar = ticksPerBar(ppq, sig);
  const perBeat = ticksPerBeat(ppq, sig);
  const bar = Math.floor(tick / perBar) + 1;
  const beat = Math.floor((tick % perBar) / perBeat) + 1;
  return `${bar}.${beat}`;
}

export type ScheduledNote = {
  pitch: number;
  velocity: number;
  /** MIDI channel of the source track; 9 is GM percussion. */
  channel: number;
  /** seconds from the start of the timeline */
  startSec: number;
  durationSec: number;
};

/** Flattens every track into one list of notes ordered by start time. */
export function scheduleNotes(tl: Timeline): ScheduledNote[] {
  const out: ScheduledNote[] = [];
  const bpm = quarterBpm(tl);
  for (const track of tl.tracks) {
    for (const n of track.notes) {
      out.push({
        pitch: n.pitch,
        velocity: n.velocity,
        channel: track.channel,
        startSec: ticksToSeconds(n.start, tl.ppq, bpm),
        durationSec: ticksToSeconds(n.duration, tl.ppq, bpm),
      });
    }
  }
  return out.sort((a, b) => a.startSec - b.startSec);
}

export function timelineLengthTicks(tl: Timeline): number {
  let end = 0;
  for (const track of tl.tracks) {
    for (const n of track.notes) end = Math.max(end, n.start + n.duration);
  }
  for (const h of tl.harmony) end = Math.max(end, h.start + h.duration);
  return end;
}
