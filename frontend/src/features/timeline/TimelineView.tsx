import type { Timeline } from '@/shared/api/client';
import { ChordLane } from './ChordLane';
import { PianoRoll } from './PianoRoll';
import { computeLayout, harmonyAt } from './layout';
import './TimelineView.css';

export type TimelineViewProps = {
  timeline: Timeline;
  /** Playhead position in ticks. */
  positionTick: number;
  /** Index into timeline.harmony the user pinned by clicking, if any. */
  selectedHarmony?: number | null;
  onSelectHarmony?: (index: number) => void;
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
}: TimelineViewProps) {
  const layout = computeLayout(timeline);
  const notes = timeline.tracks.flatMap((t) => t.notes);

  return (
    <svg
      className="timeline"
      role="img"
      aria-label="タイムライン"
      width={layout.width}
      height={layout.height}
      viewBox={`0 0 ${layout.width} ${layout.height}`}
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
