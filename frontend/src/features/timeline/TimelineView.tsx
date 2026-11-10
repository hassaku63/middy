import { useMemo, type CSSProperties } from 'react';
import type { Timeline } from '@/shared/api/client';
import { ChordLane } from './ChordLane';
import { DrumLane } from './DrumLane';
import { PianoRoll } from './PianoRoll';
import { DRUM_CHANNEL, computeLayout, harmonyAt } from './layout';
import './TimelineView.css';

export type TimelineViewProps = {
  timeline: Timeline;
  /** Playhead position in ticks. */
  positionTick: number;
  /**
   * Whether the timeline is playing. Stopped, nothing is marked as
   * sounding (the playhead still shows where play starts); omitted, the
   * marks follow the playhead.
   */
  playing?: boolean;
  /** Index into timeline.harmony the user pinned by clicking, if any. */
  selectedHarmony?: number | null;
  onSelectHarmony?: (index: number) => void;
  /** Outline this slice, e.g. the part in view next to an overview. */
  highlight?: { x: number; width: number };
  /** Fill the box it is given, out of proportion (an overview strip). */
  stretch?: boolean;
  /** Accessible name; several timelines on one screen need distinct names. */
  label?: string;
  /** ▶ on the chord followed, ■ on the pinned one (see ChordLane). */
  marks?: boolean;
  /** Hide from assistive tech (a duplicate drawing such as an overview). */
  decorative?: boolean;
  className?: string;
};

/**
 * Chord lane plus piano roll rendered as SVG. Highlights the chord and the
 * notes under the playhead; chord blocks are clickable so the explanation
 * panel can be pinned to one of them.
 */
export function TimelineView({
  timeline,
  positionTick,
  playing = true,
  selectedHarmony = null,
  onSelectHarmony,
  highlight,
  stretch = false,
  label = 'タイムライン',
  decorative = false,
  marks,
  className,
}: TimelineViewProps) {
  // The playhead re-renders this every frame; the layout only changes with the timeline.
  const { layout, notes, drumNotes } = useMemo(
    () => ({
      layout: computeLayout(timeline),
      notes: timeline.tracks.filter((t) => t.channel !== DRUM_CHANNEL).flatMap((t) => t.notes),
      drumNotes: timeline.tracks.filter((t) => t.channel === DRUM_CHANNEL).flatMap((t) => t.notes),
    }),
    [timeline],
  );

  const sounding = playing ? positionTick : -1; // no chord or note sounds at -1

  return (
    <svg
      className={className ? `timeline ${className}` : 'timeline'}
      // A group, not an img: an img's children are hidden from assistive
      // technology, and the chord lane inside is an interactive listbox.
      role="group"
      aria-label={label}
      aria-hidden={decorative || undefined}
      width={layout.width}
      height={layout.height}
      // The drawing's proportions, for CSS to cap the height (ADR-0041).
      style={{ '--timeline-aspect': layout.width / layout.height } as CSSProperties}
      viewBox={`0 0 ${layout.width} ${layout.height}`}
      preserveAspectRatio={stretch ? 'none' : undefined}
    >
      {Array.from({ length: layout.bars + 1 }, (_, i) => (
        <line
          key={`bar-${i}`}
          className="timeline__bar"
          x1={layout.x(i * layout.perBar)}
          x2={layout.x(i * layout.perBar)}
          y1={0}
          y2={layout.height}
        />
      ))}
      <ChordLane
        harmony={timeline.harmony}
        layout={layout}
        activeIndex={harmonyAt(timeline.harmony, sounding)}
        selectedIndex={selectedHarmony}
        onSelect={onSelectHarmony}
        focusable={!decorative}
        marks={marks}
        ppq={timeline.ppq}
      />
      <PianoRoll notes={notes} layout={layout} positionTick={sounding} />
      <DrumLane notes={drumNotes} layout={layout} positionTick={sounding} />
      {highlight && (
        <rect
          className="timeline__highlight"
          x={highlight.x}
          y={0}
          width={highlight.width}
          height={layout.height}
        />
      )}
      <line
        className="timeline__playhead"
        data-testid="playhead"
        x1={layout.x(positionTick)}
        x2={layout.x(positionTick)}
        y1={0}
        y2={layout.height}
      />
    </svg>
  );
}
