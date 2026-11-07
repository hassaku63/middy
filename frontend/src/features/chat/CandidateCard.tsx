import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { pitchClassName } from '@/shared/lib/notation';

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
  const accName = patterns?.accompaniments.find((a) => a.id === candidate.accompaniment)?.name;
  const drumName = patterns?.drums.find((d) => d.id === candidate.drums)?.name;
  return (
    <article className={`candidate${open ? ' is-open' : ''}`} aria-label={title}>
      <h3 className="candidate__title">{title}</h3>
      <p className="candidate__spec">
        <span>
          {pitchClassName(key.tonic)} {key.mode}
        </span>
        <span>{tempoBpm} BPM</span>
        <span>{beatsPerChord} 拍/コード</span>
      </p>
      <p className="candidate__progression">{progression.join(' - ')}</p>
      {(candidate.accompaniment || candidate.drums) && (
        <p className="candidate__spec">
          {candidate.accompaniment && <span>伴奏: {accName ?? candidate.accompaniment}</span>}
          {candidate.drums && <span>ドラム: {drumName ?? candidate.drums}</span>}
        </p>
      )}
      <p className="candidate__reason">{reason}</p>
      <button type="button" onClick={onOpen} aria-pressed={open}>
        {open ? '開いています' : 'この候補を開く'}
      </button>
    </article>
  );
}
