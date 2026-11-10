import type { Cadence, HarmonyEvent } from '@/shared/api/client';
import { FUNCTION_NAME, harmonySymbol } from '@/shared/lib/notation';

const CADENCE_LABEL: Record<Cadence, string> = {
  authentic: '正格終止',
  plagal: '変格終止',
  half: '半終止',
  deceptive: '偽終止',
  modal: 'モーダル終止',
};

type Props = {
  event: HarmonyEvent;
  pinned: boolean;
  onClearSelection?: (() => void) | undefined;
  /** Label the chord by function; false in a mode (ADR-0045). */
  functional?: boolean;
};

/** The harmony half of the explanation: one chord's function, cadence and text. */
export function ChordNotes({ event, pinned, onClearSelection, functional = true }: Props) {
  const analysis = event.analysis;
  return (
    <>
      <header className="explanation__header">
        <h2 className="explanation__title">
          <span className="explanation__numeral">{event.romanNumeral}</span>
          <span className="explanation__symbol">{harmonySymbol(event)}</span>
        </h2>
        {analysis && functional && (
          <span className={`explanation__function is-${analysis.function}`}>
            {analysis.function} · {FUNCTION_NAME[analysis.function]}
          </span>
        )}
        {analysis?.cadence && (
          <span className="explanation__cadence">{CADENCE_LABEL[analysis.cadence]}</span>
        )}
      </header>
      <p className="explanation__text">
        {analysis?.explanation ?? 'このコードの分析はまだありません。'}
      </p>
      {pinned && onClearSelection && (
        <button type="button" className="explanation__unpin" onClick={onClearSelection}>
          固定を解除して再生に追従
        </button>
      )}
    </>
  );
}

/**
 * The one live region of the explanation (#274, WCAG 4.1.3): the sounding
 * or pinned chord's name and, in major and minor, its function, short enough to announce at the
 * pace of playback. The longer text is there to read, not announced.
 */
export function ChordStatus({
  event,
  functional = true,
}: {
  event: HarmonyEvent | undefined;
  /** Name the chord's function; false in a mode (ADR-0045). */
  functional?: boolean;
}) {
  const fn = functional ? event?.analysis?.function : undefined;
  return (
    <p className="explanation__live" aria-live="polite">
      {event
        ? `${event.romanNumeral} ${harmonySymbol(event)}${fn ? ` ${FUNCTION_NAME[fn]}` : ''}`
        : ''}
    </p>
  );
}
