import type { Cadence, HarmonicFunction, Timeline } from '@/shared/api/client';
import { chordSymbol } from '@/shared/lib/notation';
import { explainedIndex } from './explainedIndex';
import './ExplanationPanel.css';

export type ExplanationPanelProps = {
  timeline: Timeline | null;
  /** Playhead position; the panel follows it unless a chord is pinned. */
  positionTick: number;
  /** Index into timeline.harmony pinned by the user, or null to follow playback. */
  selectedHarmony: number | null;
  onClearSelection?: () => void;
};

const FUNCTION_LABEL: Record<HarmonicFunction, string> = {
  T: 'トニック',
  S: 'サブドミナント',
  D: 'ドミナント',
};

const CADENCE_LABEL: Record<Cadence, string> = {
  authentic: '正格終止',
  plagal: '変格終止',
  half: '半終止',
  deceptive: '偽終止',
};

/**
 * Shows the analysis of one chord. All text comes from the backend's
 * deterministic Analysis (ADR-0003); this component only lays it out.
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
      {!event ? (
        <p className="explanation__empty">
          {timeline
            ? '再生するか、コードをクリックすると解説が表示されます。'
            : 'タイムラインがありません。'}
        </p>
      ) : (
        <>
          <header className="explanation__header">
            <h2 className="explanation__title">
              <span className="explanation__numeral">{event.romanNumeral}</span>
              <span className="explanation__symbol">{chordSymbol(event.chord)}</span>
            </h2>
            {event.analysis && (
              <span className={`explanation__function is-${event.analysis.function}`}>
                {event.analysis.function} · {FUNCTION_LABEL[event.analysis.function]}
              </span>
            )}
            {event.analysis?.cadence && (
              <span className="explanation__cadence">{CADENCE_LABEL[event.analysis.cadence]}</span>
            )}
          </header>
          <p className="explanation__text">
            {event.analysis?.explanation ?? 'このコードの分析はまだありません。'}
          </p>
          {pinned && onClearSelection && (
            <button type="button" className="explanation__unpin" onClick={onClearSelection}>
              固定を解除して再生に追従
            </button>
          )}
        </>
      )}
    </section>
  );
}
