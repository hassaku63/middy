import type { Timeline } from '@/shared/api/client';
import { isFunctional } from '@/shared/lib/notation';
import { DegreeMap } from '@/shared/ui/DegreeMap';
import { ChordNotes, ChordStatus } from './ChordNotes';
import { explainedIndex } from './explainedIndex';
import './ExplanationPanel.css';

type Props = {
  timeline: Timeline;
  positionTick: number;
  selectedHarmony: number | null;
  onClearSelection: () => void;
  /** Plays from a tick: the pinned chord's start (BL-329). */
  onPlayFrom?: (tick: number) => void;
};

/**
 * One chord's explanation in a frame of fixed size (BL-185, #279), with
 * the degree map beside it: the two are read at the same time, so they
 * share one frame (#360, ADR-0044). The frame never grows or shrinks with
 * the chord's length or text; a text longer than it scrolls inside.
 * Following the playhead or pinned shows twice without a label: the
 * frame's edge colour and, on the chord lane, ▶ or ■ (WCAG 1.4.1).
 */
export function ChordFrame({
  timeline,
  positionTick,
  selectedHarmony,
  onClearSelection,
  onPlayFrom,
}: Props) {
  const index = explainedIndex(timeline.harmony, positionTick, selectedHarmony);
  const event = index >= 0 ? timeline.harmony[index] : undefined;
  const pinned = selectedHarmony !== null && index === selectedHarmony;
  // Functions (T, S, D) are a theory of major and minor; in a mode nothing
  // is labelled by function and the map groups the chords by their role
  // instead (ADR-0044, 0045, 0048).
  const functional = isFunctional(timeline.key.mode);
  return (
    <section
      className={`explanation explanation--frame ${pinned ? 'is-pinned' : 'is-following'}`}
      aria-label="コードの解説"
    >
      {timeline.overview && (
        <DegreeMap
          degrees={timeline.overview.degrees}
          event={event}
          flow={functional}
          functions={functional}
          roles={!functional}
          pinned={pinned}
        />
      )}
      {/* A long text scrolls inside; the keyboard can reach it (WCAG 2.1.1). */}
      <div className="explanation__chord" tabIndex={0}>
        <ChordStatus event={event} functional={functional} />
        {event ? (
          <ChordNotes
            event={event}
            pinned={pinned}
            onClearSelection={onClearSelection}
            onPlayFrom={onPlayFrom && (() => onPlayFrom(event.start))}
            functional={functional}
          />
        ) : (
          <p className="explanation__empty">
            再生するか、コードを押すと解説が出ます。← → で前後のコードへ、Esc で再生に戻ります。
          </p>
        )}
      </div>
    </section>
  );
}
