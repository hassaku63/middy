import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { keyName } from '@/shared/lib/notation';

type Props = {
  candidate: IdeaCandidate;
  open: boolean;
  onOpen: () => void;
  /** Catalogue used to show pattern names instead of ids. */
  patterns?: PatternsResponse | undefined;
};

/** One proposal: what it is, why it fits, and a button to load it. */
export function CandidateCard({ candidate, open, onOpen, patterns }: Props) {
  const { title, key, tempoBpm, progression, beatsPerChord, beats, reason } = candidate;
  const labels = patternLabels(candidate, patterns);
  return (
    <article className={`candidate${open ? ' is-open' : ''}`} aria-label={title}>
      <h3 className="candidate__title">{title}</h3>
      <p className="candidate__spec">
        <span>{keyName(key)}</span>
        <span>{tempoBpm} BPM</span>
        <span>{beats ? `長さ: ${beats.join('・')} 拍` : `${beatsPerChord} 拍/コード`}</span>
      </p>
      <p className="candidate__progression">{progression.join(' - ')}</p>
      {labels.length > 0 && (
        <p className="candidate__spec">
          {labels.map((l) => (
            <span key={l}>{l}</span>
          ))}
        </p>
      )}
      {/* Hear first, then read why (intent 4, BL-073): the reason follows the button. */}
      <button type="button" onClick={onOpen} aria-pressed={open}>
        {open ? '開いています' : 'この候補を開く'}
      </button>
      <p className="candidate__reason">{reason}</p>
    </article>
  );
}

/**
 * Names for the candidate's patterns, from the catalogue when it is loaded.
 * Straight feel and 4/4 are the defaults and go unsaid.
 */
function patternLabels(c: IdeaCandidate, patterns: PatternsResponse | undefined): string[] {
  const name = (list: { id: string; name: string }[] | undefined, id: string) =>
    list?.find((x) => x.id === id)?.name ?? id;
  const labels: string[] = [];
  if (c.style) labels.push(`スタイル: ${name(patterns?.styles, c.style)}`);
  if (c.accompaniment) labels.push(`伴奏: ${name(patterns?.accompaniments, c.accompaniment)}`);
  if (c.drums) labels.push(`ドラム: ${name(patterns?.drums, c.drums)}`);
  if (c.feel && c.feel !== 'straight') labels.push(`ノリ: ${name(patterns?.feels, c.feel)}`);
  if (c.meter && c.meter !== '4/4') labels.push(`拍子: ${c.meter}`);
  return labels;
}
