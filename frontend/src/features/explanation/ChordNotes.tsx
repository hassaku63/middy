import type { Cadence, HarmonicFunction, HarmonyEvent } from '@/shared/api/client';
import { harmonySymbol } from '@/shared/lib/notation';

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

type Props = {
  event: HarmonyEvent;
  pinned: boolean;
  onClearSelection?: (() => void) | undefined;
};

/** The harmony half of the explanation: one chord's function, cadence and text. */
export function ChordNotes({ event, pinned, onClearSelection }: Props) {
  const analysis = event.analysis;
  return (
    <>
      <header className="explanation__header">
        <h2 className="explanation__title">
          <span className="explanation__numeral">{event.romanNumeral}</span>
          <span className="explanation__symbol">{harmonySymbol(event)}</span>
        </h2>
        {analysis && (
          <span className={`explanation__function is-${analysis.function}`}>
            {analysis.function} · {FUNCTION_LABEL[analysis.function]}
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
