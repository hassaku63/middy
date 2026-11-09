import type { Timeline } from '@/shared/api/client';
import { ChordNotes, ChordStatus } from './ChordNotes';
import { explainedIndex } from './explainedIndex';
import './ExplanationPanel.css';

type Props = {
  timeline: Timeline;
  positionTick: number;
  selectedHarmony: number | null;
  onClearSelection: () => void;
};

/**
 * One chord's explanation in a frame of fixed size, above the timeline on
 * the desk under trial (BL-185, #279). The frame never grows or shrinks
 * with the chord's length or text; a text longer than it scrolls inside.
 * Following the playhead or pinned shows twice without a label: the
 * frame's edge colour and, on the chord lane, ▶ or ■ (WCAG 1.4.1).
 */
export function ChordFrame({ timeline, positionTick, selectedHarmony, onClearSelection }: Props) {
  const index = explainedIndex(timeline.harmony, positionTick, selectedHarmony);
  const event = index >= 0 ? timeline.harmony[index] : undefined;
  const pinned = selectedHarmony !== null && index === selectedHarmony;
  return (
    <section
      className={`explanation explanation--frame ${pinned ? 'is-pinned' : 'is-following'}`}
      aria-label="コードの解説"
      // A long text scrolls inside the frame; the keyboard can reach it (WCAG 2.1.1).
      tabIndex={0}
    >
      <ChordStatus event={event} />
      {event ? (
        <ChordNotes event={event} pinned={pinned} onClearSelection={onClearSelection} />
      ) : (
        <p className="explanation__empty">
          再生するか、コードを押すと解説が出ます。← → で前後のコードへ、Esc で再生に戻ります。
        </p>
      )}
    </section>
  );
}
