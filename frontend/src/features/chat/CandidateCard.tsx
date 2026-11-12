import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { patternLabels, specOf } from './cardText';
import { chordNamesOf, numeralsOf } from './progression';

type Props = {
  candidate: IdeaCandidate;
  open: boolean;
  onOpen: () => void;
  /** Catalogue used to show pattern names instead of ids. */
  patterns?: PatternsResponse | undefined;
};

/**
 * One proposal: what it is and why it fits. The whole card is what the
 * learner presses to open it (#313): the title is the button, and its
 * hit area covers the card.
 */
export function CandidateCard({ candidate, open, onOpen, patterns }: Props) {
  const { title, reason } = candidate;
  const names = chordNamesOf(candidate);
  const labels = patternLabels(candidate, patterns);
  return (
    <article className={`candidate${open ? ' is-open' : ''}`} aria-label={title}>
      <h3 className="candidate__title">
        <button type="button" className="candidate__open" onClick={onOpen} aria-pressed={open}>
          {title}
        </button>
      </h3>
      <p className="candidate__spec">
        {specOf(candidate).map((s) => (
          <span key={s}>{s}</span>
        ))}
      </p>
      <p className="candidate__progression">{numeralsOf(candidate).join(' - ')}</p>
      {names.length > 0 && <p className="candidate__chords">{names.join(' - ')}</p>}
      {labels.length > 0 && (
        <p className="candidate__spec">
          {labels.map((l) => (
            <span key={l}>{l}</span>
          ))}
        </p>
      )}
      {/* Hear first, then read why (intent 4, BL-073): the reason comes last. */}
      <p className="candidate__reason">{reason}</p>
    </article>
  );
}
