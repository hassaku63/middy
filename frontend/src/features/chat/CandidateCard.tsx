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
  const { title, key, tempoBpm, progression, beatsPerChord, reason } = candidate;
  const labels = patternLabels(candidate, patterns);
  return (
    <article className={`candidate${open ? ' is-open' : ''}`} aria-label={title}>
      <h3 className="candidate__title">{title}</h3>
      <p className="candidate__spec">
        <span>{keyName(key)}</span>
        <span>{tempoBpm} BPM</span>
        <span>{beatsPerChord} 拍/コード</span>
      </p>
      <p className="candidate__progression">{progression.join(' - ')}</p>
      {labels.length > 0 && (
        <p className="candidate__spec">
          {labels.map((l) => (
            <span key={l}>{l}</span>
          ))}
        </p>
      )}
      <p className="candidate__reason">{reason}</p>
      <button type="button" onClick={onOpen} aria-pressed={open}>
        {open ? '開いています' : 'この候補を開く'}
      </button>
    </article>
  );
}

/**
 * Names for the candidate's patterns, from the catalogue when it is loaded.
 * Straight is the default feel and goes unsaid.
 */
function patternLabels(c: IdeaCandidate, patterns: PatternsResponse | undefined): string[] {
  const name = (list: { id: string; name: string }[] | undefined, id: string) =>
    list?.find((x) => x.id === id)?.name ?? id;
  const labels: string[] = [];
  if (c.accompaniment) labels.push(`伴奏: ${name(patterns?.accompaniments, c.accompaniment)}`);
  if (c.drums) labels.push(`ドラム: ${name(patterns?.drums, c.drums)}`);
  if (c.feel && c.feel !== 'straight') labels.push(`ノリ: ${name(patterns?.feels, c.feel)}`);
  return labels;
}
