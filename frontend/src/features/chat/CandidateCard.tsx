import type { IdeaCandidate } from '@/shared/api/client';
import { pitchClassName } from '@/shared/lib/notation';

type Props = {
  candidate: IdeaCandidate;
  open: boolean;
  onOpen: () => void;
};

/** One proposal: what it is, why it fits, and a button to load it. */
export function CandidateCard({ candidate, open, onOpen }: Props) {
  const { title, key, tempoBpm, progression, beatsPerChord, reason } = candidate;
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
      <p className="candidate__reason">{reason}</p>
      <button type="button" onClick={onOpen} aria-pressed={open}>
        {open ? '開いています' : 'この候補を開く'}
      </button>
    </article>
  );
}
