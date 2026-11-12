import { useMemo, useRef, type CSSProperties } from 'react';
import type { Timeline } from '@/shared/api/client';
import { ChordLane } from './ChordLane';
import { DrumLane } from './DrumLane';
import { PianoRoll } from './PianoRoll';
import { TimelineGrid } from './TimelineGrid';
import { TimelineRuler } from './TimelineRuler';
import {
  DRUM_CHANNEL,
  ROW_HEIGHT,
  computeLayout,
  fitBarWidth,
  harmonyAt,
  type Layout,
} from './layout';
import { useBoxSize } from './useBoxSize';
import { GRIP_HALF, useHeadGrip } from './useHeadGrip';
import { useSeekByClick, type Scrub } from './useSeekByClick';
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
  onSelectHarmony?: ((index: number) => void) | undefined;
  /** Clicking anywhere but a chord moves the playhead there (BL-328). */
  onSeek?: ((tick: number) => void) | undefined;
  /** A drag that moves the playhead starts or ends (BL-335). */
  onScrub?: Scrub | undefined;
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
  onSeek,
  onScrub,
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
  // A thin strip draws no bar numbers (#435).
  const natural = useMemo(() => computeLayout(timeline, undefined, !stretch), [timeline, stretch]);
  const barWidth = box ? fitBarWidth(natural, box) : natural.barWidth;
  const { layout, notes, drumNotes } = useMemo(
    () => ({
      layout: barWidth === natural.barWidth ? natural : computeLayout(timeline, barWidth, !stretch),
      notes: timeline.tracks.filter((t) => t.channel !== DRUM_CHANNEL).flatMap((t) => t.notes),
      drumNotes: timeline.tracks.filter((t) => t.channel === DRUM_CHANNEL).flatMap((t) => t.notes),
    }),
    [timeline, natural, barWidth, stretch],
  );

  const seek = useSeekByClick(ref, layout, onSeek, onSelectHarmony !== undefined, onScrub);
  const grip = useHeadGrip(ref, layout, positionTick, onSeek, onScrub);
  const sounding = playing ? positionTick : -1; // no chord or note sounds at -1

  return (
    <svg
      className={['timeline', onSeek && 'timeline--seekable', className].filter(Boolean).join(' ')}
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
      {...seek.handlers}
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
      {!stretch && <TimelineRuler layout={layout} />}
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
      {seek.hint !== null && <SeekHint layout={layout} tick={seek.hint} />}
      <Playhead layout={layout} tick={positionTick} grip={grip} />
    </svg>
  );
}

/** Where a click would move the playhead (BL-328), drawn like a thin playhead. */
function SeekHint({ layout, tick }: { layout: Layout; tick: number }) {
  return (
    <line
      className="timeline__seek-hint"
      x1={layout.x(tick)}
      x2={layout.x(tick)}
      y1={layout.rollTop}
      y2={layout.height}
    />
  );
}

/**
 * The playhead, and the band a finger holds it by (BL-337). It runs under
 * the chord lane, where the sounding chord already shows by its own mark:
 * across the lane it hid the chord's numeral at the start of each chord
 * (#402 I).
 */
function Playhead({
  layout,
  tick,
  grip,
}: {
  layout: Layout;
  tick: number;
  grip: ReturnType<typeof useHeadGrip>;
}) {
  const x = layout.x(tick);
  return (
    <>
      <line
        className="timeline__playhead"
        data-testid="playhead"
        x1={x}
        x2={x}
        y1={layout.rollTop}
        y2={layout.height}
      />
      {grip && (
        <rect
          className="timeline__head-grip"
          data-testid="head-grip"
          x={x - GRIP_HALF}
          y={layout.rollTop}
          width={GRIP_HALF * 2}
          height={layout.height - layout.rollTop}
          {...grip}
        />
      )}
    </>
  );
}
