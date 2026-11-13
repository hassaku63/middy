import type { IdeaCandidate, PatternsResponse } from '@/shared/api/client';
import { Fold } from '@/shared/ui/Fold';
import { cardLines, specOf } from './cardText';
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
 * proposal, kept apart from the engine's notes (ADR-0003). It says what the
 * card in the conversation says, so it is folded at first, with the key,
 * tempo and length beside its title (#511).
 */
export function ExampleNotes({ candidate, image, patterns, changed }: Props) {
  return (
    <div className="example-notes" aria-label="この例（AI の提案）">
      <Fold
        title={
          <>
            この例（AI の提案）
            {changed && (
              <span className="example-notes__changed">キーやスタイルを変える前の、元の例</span>
            )}
          </>
        }
        gist={specOf(candidate)
          .filter((s) => !s.includes('拍'))
          .join(' · ')}
      >
        {cardLines(candidate, image, patterns).map((line, i) => (
          <p key={i} className="example-notes__line">
            {line}
          </p>
        ))}
      </Fold>
    </div>
  );
}
