import type { Cadence, HarmonyEvent } from '@/shared/api/client';
import { FUNCTION_NAME, harmonySymbol, roleName } from '@/shared/lib/notation';

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
  /** Plays from the pinned chord's start (BL-329), when given. */
  onPlayFrom?: (() => void) | undefined;
  /** Label the chord by function; false in a mode (ADR-0045). */
  functional?: boolean;
};

/** The harmony half of the explanation: one chord's function, cadence and text. */
export function ChordNotes({
  event,
  pinned,
  onClearSelection,
  onPlayFrom,
  functional = true,
}: Props) {
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
        {/* In a mode the chord is named by its role instead (ADR-0048). */}
        {analysis && !functional && roleName(analysis.role) && (
          <span className={`explanation__role is-role-${analysis.role}`}>
            {roleName(analysis.role)}
          </span>
        )}
        {analysis?.cadence && (
          <span className="explanation__cadence">{CADENCE_LABEL[analysis.cadence]}</span>
        )}
      </header>
      <p className="explanation__text">
        {analysis?.explanation ?? 'このコードの分析はまだありません。'}
      </p>
      {pinned && <PinnedActions onPlayFrom={onPlayFrom} onClearSelection={onClearSelection} />}
    </>
  );
}

/**
 * The one live region of the explanation (#274, WCAG 4.1.3): the sounding
 * or pinned chord's name and its function (in major and minor) or its role
 * (in a mode, ADR-0048), short enough to announce at the pace of playback. The longer text is there to read, not announced.
 */
export function ChordStatus({
  event,
  functional = true,
}: {
  event: HarmonyEvent | undefined;
  /** Name the chord's function; false in a mode, where its role is named (ADR-0045, 0048). */
  functional?: boolean;
}) {
  const analysis = event?.analysis;
  const name = analysis
    ? functional
      ? FUNCTION_NAME[analysis.function]
      : roleName(analysis.role)
    : '';
  return (
    <p className="explanation__live" aria-live="polite">
      {event ? `${event.romanNumeral} ${harmonySymbol(event)}${name ? ` ${name}` : ''}` : ''}
    </p>
  );
}

/**
 * What a pinned chord offers: play the music from it (BL-329, the
 * keyboard's way to the timeline's click, WCAG 2.1.1: ← → pin a chord,
 * this plays from it) and going back to following the playhead.
 */
function PinnedActions({
  onPlayFrom,
  onClearSelection,
}: Pick<Props, 'onPlayFrom' | 'onClearSelection'>) {
  if (!onPlayFrom && !onClearSelection) return null;
  return (
    <div className="explanation__actions">
      {onPlayFrom && (
        <button type="button" className="explanation__play-from" onClick={onPlayFrom}>
          ここから再生
        </button>
      )}
      {onClearSelection && (
        <button type="button" className="explanation__unpin" onClick={onClearSelection}>
          固定を解除して再生に追従
        </button>
      )}
    </div>
  );
}
