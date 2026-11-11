import { useMemo, useRef, type CSSProperties } from 'react';
import type { Timeline } from '@/shared/api/client';
import { ChordLane } from './ChordLane';
import { DrumLane } from './DrumLane';
import { PianoRoll } from './PianoRoll';
import { TimelineGrid } from './TimelineGrid';
import {
  CHORD_LANE_HEIGHT,
  DRUM_CHANNEL,
  ROW_HEIGHT,
  computeLayout,
  fitBarWidth,
  harmonyAt,
} from './layout';
import { useBoxSize } from './useBoxSize';
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
  /**
   * Fill the width of a box whose height is capped by widening the bars,
   * keeping the rows and the text at their size (#402 F).
   */
  fit?: boolean;
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
  fit = false,
  label = 'タイムライン',
  decorative = false,
  marks,
  className,
}: TimelineViewProps) {
  const ref = useRef<SVGSVGElement>(null);
  const box = useBoxSize(ref, fit);
  // The playhead re-renders this every frame; the layout only changes with
  // the timeline and the box it fills.
  const natural = useMemo(() => computeLayout(timeline), [timeline]);
  const barWidth = box ? fitBarWidth(natural, box) : natural.barWidth;
  const { layout, notes, drumNotes } = useMemo(
    () => ({
      layout: barWidth === natural.barWidth ? natural : computeLayout(timeline, barWidth),
      notes: timeline.tracks.filter((t) => t.channel !== DRUM_CHANNEL).flatMap((t) => t.notes),
      drumNotes: timeline.tracks.filter((t) => t.channel === DRUM_CHANNEL).flatMap((t) => t.notes),
    }),
    [timeline, natural, barWidth],
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
      ref={ref}
      // The drawing's own proportions at BAR_WIDTH, for CSS to cap the
      // height (ADR-0041); a fitted drawing widens its bars from there.
      style={{ '--timeline-aspect': natural.width / natural.height } as CSSProperties}
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
      <PianoRoll
        notes={notes}
        layout={layout}
        positionTick={sounding}
        grid={
          <TimelineGrid
            layout={layout}
            top={layout.rollTop}
            bottom={layout.rollTop + layout.rows * ROW_HEIGHT}
            beats={!stretch}
          />
        }
      />
      <DrumLane
        notes={drumNotes}
        layout={layout}
        positionTick={sounding}
        grid={
          <TimelineGrid
            layout={layout}
            top={layout.drumTop}
            bottom={layout.height}
            beats={!stretch}
          />
        }
      />
      {highlight && (
        <rect
          className="timeline__highlight"
          x={highlight.x}
          y={0}
          width={highlight.width}
          height={layout.height}
        />
      )}
      {/* The playhead runs under the chord lane, where the sounding chord
          already shows by its own mark: across the lane it hid the chord's
          numeral at the start of each chord (#402 I). */}
      <line
        className="timeline__playhead"
        data-testid="playhead"
        x1={layout.x(positionTick)}
        x2={layout.x(positionTick)}
        y1={CHORD_LANE_HEIGHT}
        y2={layout.height}
      />
    </svg>
  );
}
