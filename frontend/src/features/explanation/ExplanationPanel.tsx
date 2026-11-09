import type { Timeline } from '@/shared/api/client';
import { DegreeMap } from '@/shared/ui/DegreeMap';
import { ChordNotes } from './ChordNotes';
import { explainedIndex } from './explainedIndex';
import { OverviewNotes } from './OverviewNotes';
import { RhythmNotes } from './RhythmNotes';
import './ExplanationPanel.css';

export type ExplanationPanelProps = {
  timeline: Timeline | null;
  /** Playhead position; the panel follows it unless a chord is pinned. */
  positionTick: number;
  /** Index into timeline.harmony pinned by the user, or null to follow playback. */
  selectedHarmony: number | null;
  onClearSelection?: () => void;
};

/**
 * Shows the example as a whole (mode and flow), then the analysis of one
 * chord and, when patterns were applied, the rhythm notes. All text comes from the backend's deterministic analysis
 * (ADR-0003); this component only lays it out.
 */
export function ExplanationPanel({
  timeline,
  positionTick,
  selectedHarmony,
  onClearSelection,
}: ExplanationPanelProps) {
  const harmony = timeline?.harmony ?? [];
  const index = explainedIndex(harmony, positionTick, selectedHarmony);
  const event = index >= 0 ? harmony[index] : undefined;
  const pinned = selectedHarmony !== null && index === selectedHarmony;

  return (
    <section className="explanation" aria-label="解説" aria-live="polite">
      {timeline?.overview && <OverviewNotes overview={timeline.overview} />}
      {timeline?.overview && <DegreeMap degrees={timeline.overview.degrees} event={event} />}
      {event ? (
        <ChordNotes event={event} pinned={pinned} onClearSelection={onClearSelection} />
      ) : (
        <p className="explanation__empty">
          {timeline
            ? '再生するか、コードをクリックすると解説が表示されます。'
            : 'タイムラインがありません。'}
        </p>
      )}
      {timeline?.rhythm && <RhythmNotes rhythm={timeline.rhythm} />}
    </section>
  );
}
