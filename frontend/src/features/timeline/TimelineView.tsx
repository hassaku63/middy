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
  /** Index into timeline.harmony the user pinned by clicking, if any. */
  selectedHarmony?: number | null;
  onSelectHarmony?: (index: number) => void;
  /**
   * Show only this horizontal slice of the drawing (user units, i.e. the
   * unscaled pixels of computeLayout). Used by the long-timeline layouts
   * under ./lab (BL-080 experiment); omitted, the whole timeline is drawn.
   */
  view?: { x: number; width: number };
  /** Outline this slice, e.g. the part shown in detail next to an overview. */
  highlight?: { x: number; width: number };
  /** Accessible name; several timelines on one screen need distinct names. */
  label?: string;
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
  selectedHarmony = null,
  onSelectHarmony,
  view,
  highlight,
  label = 'タイムライン',
  decorative = false,
  className,
}: TimelineViewProps) {
  const layout = computeLayout(timeline);
  const notes = timeline.tracks.filter((t) => t.channel !== DRUM_CHANNEL).flatMap((t) => t.notes);
  const drumNotes = timeline.tracks
    .filter((t) => t.channel === DRUM_CHANNEL)
    .flatMap((t) => t.notes);

  return (
    <svg
      className={className ? `timeline ${className}` : 'timeline'}
      role="img"
      aria-label={label}
      aria-hidden={decorative || undefined}
      width={view?.width ?? layout.width}
      height={layout.height}
      viewBox={`${view?.x ?? 0} 0 ${view?.width ?? layout.width} ${layout.height}`}
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
        activeIndex={harmonyAt(timeline.harmony, positionTick)}
        selectedIndex={selectedHarmony}
        onSelect={onSelectHarmony}
      />
      <PianoRoll notes={notes} layout={layout} positionTick={positionTick} />
      <DrumLane notes={drumNotes} layout={layout} positionTick={positionTick} />
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
