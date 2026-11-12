import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { patternLabels, specOf } from './cardText';
import { chordNamesOf, numeralsOf } from './progression';
import './ExampleNotes.css';

type Props = {
  candidate: IdeaCandidate;
  /** The words the example answered. */
  image: string;
  patterns?: PatternsResponse | undefined;
  /** The key or style has changed since: the card describes the original. */
  changed: boolean;
};

/**
 * The open example's card, at the head of the explanation (BL-298, #441):
 * the words it answered, its key, tempo, length, progression, patterns and
 * reason, as plain text that can be selected and copied. It is the AI's
 * proposal, kept apart from the engine's notes below it (ADR-0003).
 */
export function ExampleNotes({ candidate, image, patterns, changed }: Props) {
  const names = chordNamesOf(candidate);
  const labels = patternLabels(candidate, patterns);
  return (
    <div className="example-notes" aria-label="この例（AI の提案）">
      <p className="example-notes__label">
        この例（AI の提案）
        {changed && <span>キーやスタイルを変える前の、元の例</span>}
      </p>
      {image && <p className="example-notes__line">イメージ: {image}</p>}
      <p className="example-notes__line">{specOf(candidate).join(' · ')}</p>
      <p className="example-notes__line">{numeralsOf(candidate).join(' - ')}</p>
      {names.length > 0 && <p className="example-notes__line">{names.join(' - ')}</p>}
      {labels.length > 0 && <p className="example-notes__line">{labels.join(' · ')}</p>}
      <p className="example-notes__line">{candidate.reason}</p>
    </div>
  );
}
