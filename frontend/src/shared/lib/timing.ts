/**
 * Pure tick/time conversions shared by the player and the timeline view.
 * Ticks are the backend's unit (see docs/architecture/domain-model.md);
 * seconds are Web Audio's.
 */
import type { TimeSignature, Timeline } from '@/shared/api/client';

export function ticksToSeconds(tick: number, ppq: number, tempoBpm: number): number {
  return (tick / ppq) * (60 / tempoBpm);
}

export function secondsToTicks(seconds: number, ppq: number, tempoBpm: number): number {
  return (seconds / (60 / tempoBpm)) * ppq;
}

/** Ticks in one bar for the given signature (beats are quarter-note based). */
export function ticksPerBar(ppq: number, sig: TimeSignature): number {
  return ppq * 4 * (sig.numerator / sig.denominator);
}

/** 1-based "bar.beat" position, e.g. tick 0 -> "1.1", one quarter later in 4/4 -> "1.2". */
export function formatPosition(tick: number, ppq: number, sig: TimeSignature): string {
  const perBar = ticksPerBar(ppq, sig);
  const perBeat = perBar / sig.numerator;
  const bar = Math.floor(tick / perBar) + 1;
  const beat = Math.floor((tick % perBar) / perBeat) + 1;
  return `${bar}.${beat}`;
}

export type ScheduledNote = {
  pitch: number;
  velocity: number;
  /** seconds from the start of the timeline */
  startSec: number;
  durationSec: number;
};

/** Flattens every track into one list of notes ordered by start time. */
export function scheduleNotes(tl: Timeline): ScheduledNote[] {
  const out: ScheduledNote[] = [];
  for (const track of tl.tracks) {
    for (const n of track.notes) {
      out.push({
        pitch: n.pitch,
        velocity: n.velocity,
        startSec: ticksToSeconds(n.start, tl.ppq, tl.tempoBpm),
        durationSec: ticksToSeconds(n.duration, tl.ppq, tl.tempoBpm),
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
