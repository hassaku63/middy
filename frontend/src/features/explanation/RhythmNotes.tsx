import type { Rhythm, RhythmPart } from '@/shared/api/client';
import { colourNotes } from './explanationText';

/** The rhythm half of the explanation: what each applied pattern does and why it feels that way. */
export function RhythmNotes({ rhythm }: { rhythm: Rhythm }) {
  const parts: [string, RhythmPart | undefined][] = [
    ['伴奏', rhythm.accompaniment],
    ['ドラム', rhythm.drums],
  ];
  // A style recorded on plain block chords has nothing to explain here.
  const colour = colourNotes(rhythm);
  if (!rhythm.accompaniment && !rhythm.drums && !rhythm.feel && colour.length === 0) return null;
  return (
    <section className="explanation__rhythm" aria-label="リズム">
      <h3 className="explanation__subtitle">リズム</h3>
      {colour.map((text) => (
        <p key={text} className="explanation__text">
          {text}
        </p>
      ))}
      {parts.map(([label, part]) =>
        part ? (
          <section key={label} aria-label={`${label}: ${part.name}`} className="explanation__part">
            <h4 className="explanation__part-title">
              {label}: {part.name}
              {part.syncopated && <span className="explanation__badge">シンコペーション</span>}
            </h4>
            <p className="explanation__text">{part.description}</p>
            <p className="explanation__text">{part.explanation}</p>
          </section>
        ) : null,
      )}
      {rhythm.feel && (
        <section aria-label={`ノリ: ${rhythm.feel.name}`} className="explanation__part">
          <h4 className="explanation__part-title">ノリ: {rhythm.feel.name}</h4>
          <p className="explanation__text">{rhythm.feel.explanation}</p>
        </section>
      )}
    </section>
  );
}
