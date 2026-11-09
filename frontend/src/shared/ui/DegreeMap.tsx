import type { DegreeChord, HarmonyEvent } from '@/shared/api/client';
import { FUNCTION_NAME } from '@/shared/lib/notation';
import { currentDegree } from './currentDegree';
import './DegreeMap.css';

export type DegreeMapProps = {
  degrees: DegreeChord[];
  /** The chord being explained (playing or pinned), if any. */
  event: HarmonyEvent | undefined;
};

/**
 * The degree map (BL-099): the seven chords of the example's mode in a
 * band, coloured by function. The ones the example plays stand out, and
 * the one being explained lights up, so the learner sees where in the mode
 * the music is. The data is the engine's (Overview.degrees). It takes no
 * place of its own, so any screen can put it where it fits.
 */
export function DegreeMap({ degrees, event }: DegreeMapProps) {
  if (degrees.length === 0) return null;
  const current = currentDegree(event);
  return (
    <ol className="degree-map" aria-label="度数の地図">
      {degrees.map((d) => (
        <li
          key={d.degree}
          className={[
            'degree-map__chord',
            `is-${d.function}`,
            d.used ? 'is-used' : '',
            d.degree === current ? 'is-current' : '',
          ]
            .filter(Boolean)
            .join(' ')}
          aria-current={d.degree === current ? 'true' : undefined}
          aria-label={`${d.romanNumeral}（${d.symbol}）、${FUNCTION_NAME[d.function]}${d.used ? '、この例で使う' : ''}`}
        >
          <span className="degree-map__numeral">{d.romanNumeral}</span>
          <span className="degree-map__symbol">{d.symbol}</span>
        </li>
      ))}
    </ol>
  );
}
